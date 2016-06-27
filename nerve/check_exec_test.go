package nerve

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExecRequireCommand(t *testing.T) {
	check := NewCheckExec()
	require.EqualError(t, check.Init(&Service{nerve: &Nerve{}}), "Exec command type require a command\n")
}

func TestExecFalse(t *testing.T) {
	check := NewCheckExec()
	check.Command = []string{"/bin/false"}
	require.NoError(t, check.Init(&Service{nerve: &Nerve{}}))

	require.Contains(t, check.Check().Error(), "Command failed")
}

func TestExecTrue(t *testing.T) {
	check := NewCheckExec()
	check.Command = []string{"/bin/true"}
	require.NoError(t, check.Init(&Service{nerve: &Nerve{}}))

	require.NoError(t, check.Check())
}
