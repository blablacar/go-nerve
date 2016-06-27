package nerve

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHttpDefault(t *testing.T) {
	check := NewCheckHttp()
	check.Path = ""
	check.Init(&Service{nerve: &Nerve{}, Port: 80})

	require.Equal(t, check.url, "http://127.0.0.1:80/")
}
