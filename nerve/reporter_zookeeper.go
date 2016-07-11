package nerve

import (
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"sync"
	"time"
)

type ReporterZookeeper struct {
	ReporterCommon
	Path                   string
	Hosts                  []string
	RefreshIntervalInMilli int
	ExposeOnUnavailable    bool

	report      Report
	reportMutex sync.Mutex
	stopChecker chan struct{}
	connection  *SharedZkConnection
	fullPath    string
	currentNode string
}

func NewReporterZookeeper() *ReporterZookeeper {
	return &ReporterZookeeper{
		RefreshIntervalInMilli: 5 * 60 * 1000, // 5min
	}
}

func (r *ReporterZookeeper) Init(s *Service) error {
	if r.Path == "" {
		return errs.WithF(r.fields, "Zookeeper reporter require a path to report to")
	}
	r.fields = r.fields.WithField("path", r.Path)
	r.fullPath = r.Path + "/" + s.Name + "_" + s.Host + "_"
	r.currentNode = r.fullPath
	return nil
}

func (r *ReporterZookeeper) Destroy() {
	if r.connection != nil {
		r.connection.Close()
	}
	if r.stopChecker != nil {
		close(r.stopChecker)
	}
}

func (r *ReporterZookeeper) sendReportToZk() error {
	r.reportMutex.Lock()
	defer r.reportMutex.Unlock()

	state, err := r.Connect()
	logs.WithF(r.fields).Debug("Connected")
	if err != nil /*|| state != zk.StateHasSession*/ {
		return errs.WithEF(err, r.fields.WithField("state", state), "Failed to connect to zookeeper")
	}

	exists, _, _ := r.connection.Conn.Exists(r.currentNode)
	if r.report.Available || r.ExposeOnUnavailable {
		content, err := r.report.toJson()
		if err != nil {
			return errs.WithEF(err, r.fields, "Failed to prepare report")
		}

		logs.WithF(r.fields).Debug("will write status")
		if !exists {
			logs.WithF(r.fields).Debug("does not exists")
			acl := zk.WorldACL(zk.PermAll)
			//Create Full Static Path if not exists
			err = r.mkdirStaticPath(acl)
			if err != nil {
				return errs.WithEF(err, r.fields, "Cannot create static path")
			}
			//Don't use Create, as it's an ephemeral zk node, and there's some race condition to avoid
			r.currentNode, err = r.connection.Conn.CreateProtectedEphemeralSequential(r.fullPath, content, acl)
			if err != nil {
				return errs.WithEF(err, r.fields.WithField("fullpath", r.fullPath), "Cannot create path")
			}
		} else {
			logs.WithF(r.fields).Debug("writting data")
			_, err := r.connection.Conn.Set(r.currentNode, content, -1)
			if err != nil {
				return errs.WithE(err, "Failed to write status")
			}
		}
	} else {
		if exists {
			logs.WithF(r.fields).Debug("delete")
			err = r.connection.Conn.Delete(r.currentNode, -1)
			if err != nil {
				return errs.WithEF(err, r.fields.WithField("fullpath", r.fullPath), "Cannot delete node")
			}
		}
	}
	return nil
}

func (r *ReporterZookeeper) Report(report Report) error {
	r.report = report
	if r.stopChecker == nil {
		r.stopChecker = make(chan struct{})
		go r.refresher()
	}

	return r.sendReportToZk()
}

func (r *ReporterZookeeper) refresher() {
	for {
		select {
		case <-r.stopChecker:
			logs.WithFields(r.fields).Debug("Stop refresher requested")
			return
		default:
			time.Sleep(time.Duration(r.RefreshIntervalInMilli) * time.Millisecond)
		}

		logs.WithFields(r.fields).Debug("Refreshing report")
		if err := r.sendReportToZk(); err != nil {
			logs.WithEF(err, r.fields).Error("Failed to refresh status in zookeeper")
		}
	}
}

func (r *ReporterZookeeper) Connect() (zk.State, error) {
	if r.connection != nil {
		state := r.connection.Conn.State()
		switch state {
		case zk.StateUnknown, zk.StateConnectedReadOnly, zk.StateExpired, zk.StateAuthFailed, zk.StateConnecting:
			logs.WithFields(r.fields).WithField("state", state).Warn("Connection is in bad state")
			r.connection.Close()
		case zk.StateConnected, zk.StateHasSession:
			logs.WithFields(r.fields).WithField("state", state).Debug("Zookeeper connected")
			return state, nil
		case zk.StateDisconnected:
			logs.WithFields(r.fields).Info("Reporter Connection is Disconnected -> Reconnection")
		}
	}
	SharedConn, err := NewSharedZkConnection(r.Hosts, 10*time.Second)
	if err != nil {
		r.connection = nil
		return zk.StateDisconnected, errs.WithEF(err, r.fields, "Unable to connect to ZooKeeper")
	}
	r.connection = SharedConn
	state := r.connection.Conn.State()
	return state, nil
}

func (r *ReporterZookeeper) mkdirStaticPath(acl []zk.ACL) error {
	paths := strings.Split(r.Path, "/")
	full := ""
	for i, path := range paths {
		if i > 0 {
			full += "/"
		}
		full += path
		if exists, _, _ := r.connection.Conn.Exists(full); exists {
			continue
		}

		logs.WithFields(r.fields.WithField("full", full)).Debug("Creating zk path")

		_, err := r.connection.Conn.Create(full, []byte(""), int32(0), acl)
		if err != nil {
			return errs.WithEF(err, r.fields.WithField("path", full), "Cannot create path")
		}
	}
	return nil
}
