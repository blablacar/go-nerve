package nerve

import (
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"gopkg.in/macaron.v1"
	"net"
	"net/http"
	"strconv"
)

func (n *Nerve) DisableServices() error {
	for _, service := range n.Services {
		service.Disable()
	}
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

	m := macaron.Classic()
	m.Get("/enable", n.EnableServices)
	m.Get("/disable", n.DisableServices)
	m.Get("/status", n.ServiceStatus)
	m.Get("/", func() string {
		return `/enable
/disable
/status`
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
