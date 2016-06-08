package nerve

import (
	"encoding/json"
	"fmt"
	"github.com/n0rad/go-erlog/logs"
	_ "github.com/n0rad/go-erlog/register"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestSimple(t *testing.T) {
	logs.SetLevel(logs.DEBUG)
	os.Remove("/tmp/nerve.report")

	nerve := &Nerve{
		Services: []*Service{{
			Host:                 "127.0.0.1",
			Port:                 1234,
			CheckIntervalInMilli: 1000,
			Rise:                 5,
			Fall:                 2,
			Reporters: []json.RawMessage{marshallNoError(ReporterFile{
				ReporterCommon: ReporterCommon{
					Type: "file",
				},
				Path: "/tmp/nerve.report",
			})},
			Checks: []json.RawMessage{marshallNoError(CheckHttp{
				CheckCommon: CheckCommon{
					Type:           "http",
					Port:           1234,
					Host:           "127.0.0.2",
					TimeoutInMilli: 2000,
				},
				Path: "/there",
			})},
		}},
	}
	require.NoError(t, nerve.Init())
	require.NoError(t, nerve.Start())
	defer nerve.Stop()

	http.HandleFunc("/there", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello astaxie!")
	})
	ln, err := net.Listen("tcp", ":1234")
	if err != nil {
		logs.WithE(err).Fatal("failed to listen")
	}
	go http.Serve(ln, nil)

	require.Nil(t, report())

	time.Sleep(2 * time.Second)
	require.Nil(t, report())

	time.Sleep(10 * time.Second)
	require.True(t, *report())

	ln.Close()
	time.Sleep(1 * time.Second)
	require.True(t, *report())

	time.Sleep(2 * time.Second)
	require.False(t, *report())
}

func marshallNoError(obj interface{}) []byte {
	data, err := json.Marshal(obj)

	if err != nil {
		logs.WithE(err).Fatal("Conversion failed")
	}
	return data
}

func report() *bool {
	res, _ := ioutil.ReadFile("/tmp/nerve.report")
	var rr bool
	switch string(res) {
	case "OK\n":
		rr = true
	case "KO\n":
		rr = false
	default:
		return nil
	}
	return &rr
}
