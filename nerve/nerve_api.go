package nerve

import (
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/macaron.v1"
	"log"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
	"encoding/json"
)

func (n *Nerve) DisableServices(ctx *macaron.Context) (string, error) {
	allWait := sync.WaitGroup{}
	for _, service := range n.Services {
		allWait.Add(1)
		go service.Disable(&allWait)
	}
	allWait.Wait()
	return n.ServicesStatus(ctx)
}

func (n *Nerve) Weight(ctx *macaron.Context) (string, error) {
	weight := uint8(ctx.ParamsInt(":weight"))
	if weight <= 0 || weight > 255 {
		return "", errs.WithF(n.fields.WithField("weight", weight), "Invalid weight value")
	}
	for _, service := range n.Services {
		service.Weight = weight
		service.runNotify()
	}
	return n.ServicesStatus(ctx)
}

func (n *Nerve) EnableServices(ctx *macaron.Context) (string, error) {
	for _, service := range n.Services {
		forceStr := ctx.QueryTrim("force")
		forceStr = strings.ToUpper(forceStr)

		force := false
		if strings.HasPrefix(forceStr, "Y") || strings.HasPrefix(forceStr, "T") || strings.HasPrefix(forceStr, "1") {
			force = true
		}
		service.Enable(force)
	}
	return n.ServicesStatus(ctx)
}

func (n *Nerve) ServicesStatus(ctx *macaron.Context) (string, error) {
	var statuses []ServiceStatus
	for _, service := range n.Services {
		statuses = append(statuses, n.ServiceStatus(service))
	}
	res, err := json.Marshal(statuses)
	if err != nil {
		ctx.Req.Header.Set("Content-Type", "application/json")
	}
	return string(res), err
}

type ServiceStatus struct {
	Name string
	Host string
	Port int
	Status string
	MaxWeight int
	CurrentWeight int
	Available bool
}

func (n *Nerve) ServiceStatus(service *Service) ServiceStatus {
	s := ServiceStatus{}
	s.Name = service.Name
	s.Host = service.Host
	s.Port = service.Port

	if service.forceEnable {
		s.Status = "forced"
	} else if service.disabled == nil {
		s.Status = "enabled"
	} else {
		s.Status = "disabled"
	}
	s.MaxWeight = int(service.Weight)
	s.CurrentWeight = int(service.CurrentWeight())
	s.Available = service.CurrentWeight() > 0
	return s
}

var favicon_ico = "\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xa4\xd3\xcd\x2b\x74\x61\x18\xc7\xf1\x7b\x7a\xe6\x21\x42\x23\x8c\x97\x4c\x4d\x4a\x68\x08\x1b\x6f\x11\x8a\x85\xb7\x05\xa2\x88\xc4\x4a\x92\xc8\x4a\x51\x88\x12\xc3\x1f\x20\x1b\x3b\x0b\xb2\xb1\xb5\xb3\x16\xb1\x96\x8d\xb7\x94\xbc\xa4\x90\x1c\xdf\xcb\x5c\x8b\xd3\x71\xa6\x59\x38\xfa\xe4\x9c\x7b\xce\xef\x3a\xf7\xb9\xee\xfb\x18\xe3\xe1\xcf\xe7\x93\xff\x41\x33\xea\x35\xc6\x6f\x8c\x29\x04\x43\x8c\x44\xc6\x7f\x0e\x7e\x3b\x0a\x45\xd8\x0e\xeb\xb7\x78\x54\x20\x1b\x99\x28\x81\x07\xa9\x48\xd2\xf3\x5e\x14\xbb\x64\x85\x17\x4b\xa8\xc1\x0c\xea\x75\x7c\x02\xad\xe8\xc2\xb1\xd6\x75\xcb\x8b\x79\xac\x60\x5a\x9f\x17\xc0\x06\xd6\x70\x8b\x2b\x84\x62\xe4\x9b\x31\xab\xf7\x0d\x60\x0b\x17\xb0\xb4\x46\xb9\xd6\x76\x66\xe3\xb0\x80\x16\x4c\x6a\x2f\x0e\xf0\xa8\x59\x71\x89\x7c\x7d\x46\x50\xeb\x94\x21\x03\x43\xe8\xc3\x30\x4a\xf5\xfd\x6f\x6c\x59\x71\x08\x1f\x9a\x90\x85\x5a\x6c\xc2\xaf\xef\x9d\xa7\xe7\xd2\xab\x13\x47\x56\xcc\xe9\x5c\xab\xb5\xcf\xd2\x97\x7e\xa4\x21\x8c\x0e\x4c\xe1\xd4\x25\xfb\x84\x3a\xcd\xb7\x6b\x5f\xaa\xb4\x96\xac\xf7\x32\x3a\xa3\x64\x5f\xb0\x8d\x14\xcd\x8f\xe8\x1c\xff\xa1\x0d\x45\x58\xc7\xb9\x4b\xf6\x19\x3b\x18\xd3\xbe\xc8\x3e\xf9\xaf\xbd\xf3\x68\x8d\x80\xce\xe7\xdd\x25\xff\x89\x37\x7c\xe1\x1a\xab\x8e\x75\xcb\x41\xb7\xf6\x2e\xec\x92\xb7\x93\x1a\x7b\xb6\x6c\x32\xc6\x75\x2d\xe5\xba\x01\xaf\x51\xb2\x32\x8f\x7d\xe4\xda\xf2\x89\x28\xb0\x5d\xcb\x9e\xbb\x73\xe4\x3e\xb0\x8b\x41\xa4\x5b\xd1\xf7\xae\x48\xc0\x22\xce\x70\xaf\x6b\xf6\x80\x9e\x18\x39\xe7\x77\x28\xdf\x70\x25\x1a\xad\xc8\x5e\xf5\x3b\xef\xfb\xd3\xf1\x1d\x00\x00\xff\xff\xc2\xa4\x56\xd0\x7e\x04\x00\x00"

func (n *Nerve) startApi() error {
	var err error
	url := n.ApiHost + ":" + strconv.Itoa(n.ApiPort)
	n.apiListener, err = net.Listen("tcp", url)
	if err != nil {
		return errs.WithEF(err, n.fields.WithField("url", url), "Failed to listen")
	}

	m := macaron.New()
	m.Use(Logger())
	m.Use(macaron.Recovery())
	m.Use(macaron.Static("public"))

	m.Get("/favicon.ico", func(resp http.ResponseWriter) {
		resp.Header().Set("Content-Encoding", "gzip")
		resp.Header().Set("Content-Type", "image/x-icon")
		resp.Header().Set("Cache-Control", "public, max-age=7776000")

		var empty [0]byte
		sx := (*reflect.StringHeader)(unsafe.Pointer(&favicon_ico))
		b := empty[:]
		bx := (*reflect.SliceHeader)(unsafe.Pointer(&b))
		bx.Data = sx.Data
		bx.Len = len(favicon_ico)
		bx.Cap = bx.Len
		resp.Write(b)
	})
	m.Get("/version", func(resp http.ResponseWriter) {
		resp.Write([]byte("version: "))
		resp.Write([]byte(n.nerveVersion))
		resp.Write([]byte("\n"))
		resp.Write([]byte("buildDate: "))
		resp.Write([]byte(n.nerveBuildTime))
		resp.Write([]byte("\n"))
	})

	m.Get("/enable", n.EnableServices)
	m.Get("/disable", n.DisableServices)
	m.Get("/status", n.ServicesStatus)
	m.Get("/metrics", prometheus.Handler())
	m.Get("/weight/:weight", n.Weight)
	m.Get("/", func() string {
		return `/enable[?force=true]
/weight/:weight
/disable
/status
/metrics
/version`
	})

	logs.WithF(n.fields.WithField("url", url)).Info("Starting api")
	go http.Serve(n.apiListener, m)
	return nil
}

func (n *Nerve) stopApi() {
	if n.apiListener != nil {
		n.apiListener.Close()
	}
	n.apiListener = nil
}

func Logger() macaron.Handler {
	var reqCounter int64
	return func(ctx *macaron.Context, log *log.Logger) {
		start := time.Now()

		fields := data.WithField("method", ctx.Req.Method).
			WithField("uri", ctx.Req.RequestURI).
			WithField("ip", ctx.RemoteAddr()).
			WithField("id", atomic.AddInt64(&reqCounter, 1))
		if logs.IsDebugEnabled() {
			logs.WithF(fields).Debug("Request received")
		}

		rw := ctx.Resp.(macaron.ResponseWriter)
		ctx.Next()

		if logs.IsInfoEnabled() {
			fields = fields.WithField("duration", time.Since(start)).WithField("status", rw.Status())
			var lvl logs.Level
			if rw.Status() >= 500 && rw.Status() < 600 {
				lvl = logs.ERROR
			} else {
				lvl = logs.INFO
			}

			logs.LogEntry(&logs.Entry{
				Fields:  fields,
				Level:   lvl,
				Message: "Request completed",
			})

		}
	}
}
