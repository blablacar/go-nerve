package nerve

import (
	"fmt"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/logs"
	"github.com/samuel/go-zookeeper/zk"
	"sort"
	"strings"
	"sync"
	"time"
	"io"
	"crypto/rand"
)

var zkConnectionsMutex sync.Mutex
var zkConnections map[string]*SharedZkConnection

func init() {
	zkConnections = make(map[string]*SharedZkConnection)
}

type SharedZkConnection struct {
	hosts      []string
	hash       string
	Conn       *zk.Conn
	err        error
	syncMutex  sync.Mutex
	recipients []chan zk.Event
	sourceChan <-chan zk.Event
	closed     bool
	connected  bool
}

type ZKLogger struct {
}

func (zl ZKLogger) Printf(format string, data ...interface{}) {
	logs.Debug("Zookeeper: " + fmt.Sprintf(format, data))
}

// this reuse zk connection if host list is the same
// a new dedicated event chan is created for each call
// zk events are duplicated to all those channels
func NewSharedZkConnection(hosts []string, timeout time.Duration) (*SharedZkConnection, error) {
	zkConnectionsMutex.Lock()
	defer zkConnectionsMutex.Unlock()

	sort.Strings(hosts)
	hash := strings.Join(hosts, "")
	if _, ok := zkConnections[hash]; !ok {
		conn, channel, err := zk.Connect(hosts, timeout)
		conn.SetLogger(ZKLogger{})
		zkConnections[hash] = &SharedZkConnection{
			hash:       hash,
			Conn:       conn,
			err:        err,
			sourceChan: channel,
		}

		go func(sharedZk *SharedZkConnection) {
			events := sharedZk.Subscribe()
			connectingCount := 0
			for {
				select {
				case e, ok := <-events:
					if !ok {
						return
					}
					if e.Type == zk.EventSession && e.State == zk.StateHasSession {
						if sharedZk.connected == false {
							logs.WithF(data.WithField("servers", hosts)).Info("Connected to zk")
							connectingCount = 0
						}
						sharedZk.connected = true
					} else if (e.Type == zk.EventSession || e.Type == zk.EventType(0)) &&
						(e.State == zk.StateDisconnected || e.State == zk.StateExpired) {
						if sharedZk.connected == true {
							logs.WithF(data.WithField("servers", hosts)).Warn("Connection lost to zk")
							connectingCount = 0
						}
						sharedZk.connected = false
					} else if (e.Type == zk.EventSession || e.Type == zk.EventType(0)) &&
						(e.State == zk.StateAuthFailed) {
						logs.WithF(data.WithField("servers", hosts)).Error("Authentication failure on zk")
					} else if (e.Type == zk.EventSession || e.Type == zk.EventType(0)) &&
						(e.State == zk.StateConnecting) {
						if connectingCount == 1 {
							logs.WithF(data.WithField("server", conn.Server())).Warn("Failed to connect to zk. Not reporting nexts servers try until connected")
						}
						connectingCount++
					}
				}
			}
		}(zkConnections[hash])
	}
	go zkConnections[hash].recipientListPublish()

	return zkConnections[hash], zkConnections[hash].err
}

func (z *SharedZkConnection) Close() {
	z.syncMutex.Lock()
	defer z.syncMutex.Unlock()

	if z.closed {
		z.Conn.Close()
		for _, newChan := range z.recipients {
			close(newChan)
		}
	}
}

func (z *SharedZkConnection) Subscribe() <-chan zk.Event {
	z.syncMutex.Lock()
	defer z.syncMutex.Unlock()

	newChan := make(chan zk.Event, 6)
	z.recipients = append(z.recipients, newChan)

	return newChan
}

func (z *SharedZkConnection) recipientListPublish() {
	for {
		select {
		case srcEvent, ok := <-z.sourceChan:
			if !ok {
				zkConnectionsMutex.Lock()
				delete(zkConnections, z.hash)
				zkConnectionsMutex.Unlock()
				return
			}
			z.syncMutex.Lock()
			for _, newChan := range z.recipients {
				select {
				case newChan <- srcEvent:
				default:
				}
			}
			z.syncMutex.Unlock()
		}
	}
}

const protectedPrefix = "_c_"

func (z *SharedZkConnection) CreateEphemeral(path string, data []byte, acl []zk.ACL) (string, error) {
	var guid [16]byte
	_, err := io.ReadFull(rand.Reader, guid[:16])
	if err != nil {
		return "", err
	}
	guidStr := fmt.Sprintf("%x", guid)

	parts := strings.Split(path, "/")
	parts[len(parts)-1] = fmt.Sprintf("%s%s%s", parts[len(parts)-1], protectedPrefix, guidStr)
	rootPath := strings.Join(parts[:len(parts)-1], "/")
	protectedPath := strings.Join(parts, "/")

	var newPath string
	for i := 0; i < 3; i++ {
		newPath, err = z.Conn.Create(protectedPath, data, zk.FlagEphemeral|zk.FlagSequence, acl)
		switch err {
		case zk.ErrSessionExpired:
		// No need to search for the node since it can't exist. Just try again.
		case zk.ErrConnectionClosed:
			children, _, err := z.Conn.Children(rootPath)
			if err != nil {
				return "", err
			}
			for _, p := range children {
				parts := strings.Split(p, "/")
				if pth := parts[len(parts)-1]; strings.HasPrefix(pth, protectedPrefix) {
					if g := pth[len(protectedPrefix) : len(protectedPrefix)+32]; g == guidStr {
						return rootPath + "/" + p, nil
					}
				}
			}
		case nil:
			return newPath, nil
		default:
			return "", err
		}
	}
	return "", err
}