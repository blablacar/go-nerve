package nerve

import (
	"net"
	"github.com/n0rad/go-erlog/errs"
	"net/http"
	"github.com/n0rad/go-erlog/logs"
	"strconv"
)


func (n *Nerve) apiDisable() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logs.WithFields(n.fields).Info("Api called /disable")
		for _, service := range n.Services {
			service.Disable()
		}
	}
}

func (n *Nerve) apiEnable() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logs.WithFields(n.fields).Info("Api called /enable")
		for _, service := range n.Services {
			service.Enable()
		}
	}
}

func (n *Nerve) apiStatus() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logs.WithFields(n.fields).Info("Api called /status")
		for _, service := range n.Services {
			// TODO json ?
			w.Write([]byte(service.Host))
			w.Write([]byte(":"))
			w.Write([]byte(strconv.Itoa(service.Port)))
			w.Write([]byte(" "))
			if service.disabled == nil {
				w.Write([]byte("enabled"))
			} else {
				w.Write([]byte("disabled"))
			}
			w.Write([]byte("\n"))
		}
	}
}

func (n *Nerve) apiMain() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logs.WithFields(n.fields).Info("Api called /")
		w.Write([]byte(`/enable
/disable
/status
`))
	}
}

func (n *Nerve) startApi() error {
	var err error
	n.apiListener, err = net.Listen("tcp", n.ApiUrl)
	if err != nil {
		return errs.WithEF(err, n.fields.WithField("url", n.ApiUrl), "Failed to listen")
	}

	http.HandleFunc("/", n.apiMain())
	http.HandleFunc("/status", n.apiStatus())
	http.HandleFunc("/disable", n.apiDisable())
	http.HandleFunc("/enable", n.apiEnable())

	logs.WithF(n.fields.WithField("url", n.ApiUrl)).Info("Api listening")

	go http.Serve(n.apiListener, nil)
	return nil
}

func (n *Nerve) stopApi() {
	if n.apiListener != nil {
		n.apiListener.Close()
	}
	n.apiListener = nil
}