package nerve

import (
	"encoding/json"
	"github.com/n0rad/go-erlog/logs"
	_ "github.com/n0rad/go-erlog/register"
	"github.com/stretchr/testify/require"
	"gopkg.in/macaron.v1"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestSimple(t *testing.T) {
	logs.SetLevel(logs.DEBUG)

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

	m := macaron.Classic()
	m.Get("/there", func() string {
		return "Hello!"
	})
	ln, err := net.Listen("tcp", ":1234")
	if err != nil {
		logs.WithE(err).Fatal("failed to listen")
	}
	go http.Serve(ln, m)
	os.Remove("/tmp/nerve.report")

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

func TestDisableEnable(t *testing.T) {
	logs.SetLevel(logs.DEBUG)

	nerve := &Nerve{
		Services: []*Service{{
			Host:                 "127.0.0.1",
			Port:                 1234,
			CheckIntervalInMilli: 1000,
			Rise:                 2,
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

	m := macaron.Classic()
	m.Get("/there", func() string {
		return "Hello!"
	})
	ln, err := net.Listen("tcp", ":1234")
	go http.Serve(ln, m)
	os.Remove("/tmp/nerve.report")

	require.Nil(t, report())

	time.Sleep(3 * time.Second)
	require.True(t, *report())

	resp, err := http.Get("http://localhost" + nerve.ApiUrl + "/disable")
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, 200)
	require.False(t, *report())

	time.Sleep(3 * time.Second)

	resp, err = http.Get("http://localhost" + nerve.ApiUrl + "/enable")
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, 200)

	time.Sleep(2 * time.Second)
	require.True(t, *report())

	resp, err = http.Get("http://localhost" + nerve.ApiUrl + "/disable")
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, 200)
	require.False(t, *report())

	ln.Close()
	time.Sleep(2 * time.Second)

	require.False(t, *report())
	resp, err = http.Get("http://localhost" + nerve.ApiUrl + "/enable")
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, 200)
	require.False(t, *report())

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
	res, err := ioutil.ReadFile("/tmp/nerve.report")
	if err != nil {
		return nil
	}
	if len(res) == 0 {
		rr := false
		return &(rr)
	}

	r, err := NewReport(res)
	if err != nil {
		panic(err)
	}
	return &r.Available
}
