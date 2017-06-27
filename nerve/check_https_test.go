package nerve

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHttpsServiceAvailable(t *testing.T) {

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	check := NewCheckHttps()
	check.Path = ""
	check.Init(&Service{nerve: &Nerve{}, Port: 443})
	check.url = ts.URL

	require.NoError(t, check.Check())
}

func TestHttpsServiceUnavailable(t *testing.T) {

	check := NewCheckHttps()
	check.Path = ""
	check.Init(&Service{nerve: &Nerve{}, Port: 443})

	require.Error(t, check.Check())
}
