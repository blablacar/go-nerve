package nerve

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCurrentWeight(t *testing.T) {
	var status error = nil
	s := Service{currentWeightIndex: 0, Weight: 100, currentStatus: &status}
	require.Equal(t, s.CurrentWeight(), uint8(1))
	s.currentWeightIndex = 1
	require.Equal(t, s.CurrentWeight(), uint8(1))
	s.currentWeightIndex = 2
	require.Equal(t, s.CurrentWeight(), uint8(1))
	s.currentWeightIndex = 3
	require.Equal(t, s.CurrentWeight(), uint8(2))
	s.currentWeightIndex = 4
	require.Equal(t, s.CurrentWeight(), uint8(3))
	s.currentWeightIndex = 5
	require.Equal(t, s.CurrentWeight(), uint8(4))
	s.currentWeightIndex = 6
	require.Equal(t, s.CurrentWeight(), uint8(6))
	s.currentWeightIndex = 7
	require.Equal(t, s.CurrentWeight(), uint8(10))
	s.currentWeightIndex = 8
	require.Equal(t, s.CurrentWeight(), uint8(15))
	s.currentWeightIndex = 9
	require.Equal(t, s.CurrentWeight(), uint8(24))
	s.currentWeightIndex = 10
	require.Equal(t, s.CurrentWeight(), uint8(39))
	s.currentWeightIndex = 11
	require.Equal(t, s.CurrentWeight(), uint8(62))
	s.currentWeightIndex = 12
	require.Equal(t, s.CurrentWeight(), uint8(100))
	s.currentWeightIndex = 13
	require.Equal(t, s.CurrentWeight(), uint8(100))
}

func TestCurrentWeight2(t *testing.T) {
	var status error = nil
	s := Service{currentWeightIndex: 0, Weight: 1, currentStatus: &status}
	require.Equal(t, s.CurrentWeight(), uint8(1))
	s.currentWeightIndex = 1
	require.Equal(t, s.CurrentWeight(), uint8(1))
	s.currentWeightIndex = 5
	require.Equal(t, s.CurrentWeight(), uint8(1))
	s.currentWeightIndex = 10
	require.Equal(t, s.CurrentWeight(), uint8(1))
}
