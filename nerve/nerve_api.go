package nerve

import (
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"gopkg.in/macaron.v1"
	"net"
	"net/http"
	"strconv"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"github.com/n0rad/go-erlog/data"
	"sync/atomic"
)

func (n *Nerve) DisableServices() error {
	for _, service := range n.Services {
		service.Disable()
	}
	time.Sleep(time.Duration(*n.DisableWaitInMilli) * time.Millisecond)
	return nil
}

func (n *Nerve) EnableServices() error {
	for _, service := range n.Services {
		service.Enable()
	}
	return nil
}

func (n *Nerve) ServiceStatus() string {
	var statuses string
	for _, service := range n.Services {
		state := service.Host + ":" + strconv.Itoa(service.Port) + " "
		if service.disabled == nil {
			state += "enabled"
		} else {
			state += "disabled"
		}
		statuses += state + "\n"
	}
	return statuses
}

func (n *Nerve) startApi() error {
	var err error
	n.apiListener, err = net.Listen("tcp", n.ApiUrl)
	if err != nil {
		return errs.WithEF(err, n.fields.WithField("url", n.ApiUrl), "Failed to listen")
	}


	m := macaron.New()
	m.Use(Logger())
	m.Use(macaron.Recovery())
	m.Use(macaron.Static("public"))

	m.Get("/enable", n.EnableServices)
	m.Get("/disable", n.DisableServices)
	m.Get("/status", n.ServiceStatus)
	m.Get("/metrics", prometheus.Handler())
	m.Get("/", func() string {
		return `/enable
/disable
/status
/metrics`
	})

	logs.WithF(n.fields.WithField("url", n.ApiUrl)).Info("Starting api")
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
		WithField("ip", ctx.RemoteAddr()).WithField("id", atomic.AddInt64(&reqCounter, 1))
		if logs.IsDebugEnabled() {
			logs.WithF(fields).Debug("Request received")
		}

		rw := ctx.Resp.(macaron.ResponseWriter)
		ctx.Next()

		if logs.IsInfoEnabled() {
			fields = fields.WithField("duration", time.Since(start)).
				WithField("status", rw.Status())
				var lvl logs.Level
				if rw.Status() >= 500 && rw.Status() < 600 {
					lvl = logs.ERROR
				} else {
					lvl = logs.INFO
				}

				logs.LogEntry(&logs.Entry{
					Fields: fields,
					Level: lvl,
					Message: "Request completed",
				})


		}
	}
}
