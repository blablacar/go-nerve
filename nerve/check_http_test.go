package nerve

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHttpDefault(t *testing.T) {
	check := NewCheckHttp()
	check.Path = ""
	check.CheckCommon.Init(&Service{Port: 80})
	check.Init(&Service{})

	require.Equal(t, check.url, "http://127.0.0.1:80/")
}
