package nerve

import (
	"github.com/stretchr/testify/require"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestTcpServiceAvailable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	check := NewCheckTcp()
	_, port, _ := net.SplitHostPort(ts.Listener.Addr().String())
	port_int, _ := strconv.ParseInt(port, 10, 0)
	check.Init(&Service{nerve: &Nerve{}, Port: int(port_int)})

	require.NoError(t, check.Check())
}

func TestTcpServiceUnavailable(t *testing.T) {
	check := NewCheckTcp()
	check.Init(&Service{nerve: &Nerve{}, Port: 80})

	require.Error(t, check.Check())
}
