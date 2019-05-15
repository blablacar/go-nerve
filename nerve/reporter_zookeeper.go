package nerve

import (
	"strings"
	"sync"
	"time"

	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"github.com/samuel/go-zookeeper/zk"
)

type ReporterZookeeper struct {
	ReporterCommon
	Path                     string
	Hosts                    []string
	ConnectionTimeoutInMilli *int
	RefreshIntervalInMilli   *int
	ExposeOnUnavailable      bool

	report      Report
	reportMutex sync.Mutex
	stopChecker chan struct{}
	connection  *SharedZkConnection
	fullPath    string
	currentNode string
}

func NewReporterZookeeper() *ReporterZookeeper {
	refreshInterval := 5 * 60 * 1000
	connectionTimeout := 2000
	return &ReporterZookeeper{
		RefreshIntervalInMilli:   &refreshInterval, // 5min
		ConnectionTimeoutInMilli: &connectionTimeout,
	}
}

func (r *ReporterZookeeper) Init(s *Service) error {
	if r.Path == "" {
		return errs.WithF(r.fields, "Zookeeper reporter require a path to report to")
	}

	r.fields = r.fields.WithField("path", r.Path)
	r.fullPath = r.Path + "/" + s.Name + "_" + s.Host
	r.currentNode = r.fullPath

	conn, err := NewSharedZkConnection(r.Hosts, time.Duration(*r.ConnectionTimeoutInMilli)*time.Millisecond)
	if err != nil {
		return errs.WithEF(err, r.fields, "Failed to prepare connection to zookeeper")
	}
	r.connection = conn
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

	exists, _, err := r.connection.Conn.Exists(r.currentNode)
	if *r.report.Available || r.ExposeOnUnavailable {
		content, err := r.report.toJson()
		if err != nil {
			return errs.WithEF(err, r.fields, "Failed to prepare report")
		}

		if !exists {
			logs.WithF(r.fields).Debug("Does not exists")
			acl := zk.WorldACL(zk.PermAll)
			if err = r.mkdirStaticPath(acl); err != nil {
				return errs.WithEF(err, r.fields, "Cannot create static path")
			}

			if r.currentNode, err = r.connection.CreateEphemeral(r.fullPath, content, acl); err != nil {
				return errs.WithEF(err, r.fields.WithField("fullpath", r.fullPath), "Cannot create path")
			}
		} else {
			logs.WithF(r.fields).Debug("writting data")
			if _, err := r.connection.Conn.Set(r.currentNode, content, -1); err != nil {
				return errs.WithE(err, "Failed to write status")
			}
		}
	} else if exists {
		logs.WithF(r.fields).Debug("delete")
		err = r.connection.Conn.Delete(r.currentNode, -1)
		if err != nil {
			return errs.WithEF(err, r.fields.WithField("fullpath", r.fullPath), "Cannot delete node")
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
			time.Sleep(time.Duration(*r.RefreshIntervalInMilli) * time.Millisecond)
		}

		logs.WithFields(r.fields).Debug("Refreshing report")
		if err := r.sendReportToZk(); err != nil {
			logs.WithEF(err, r.fields).Error("Failed to refresh status in zookeeper")
		}
	}
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
