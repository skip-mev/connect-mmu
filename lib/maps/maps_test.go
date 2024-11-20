package maps

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapInvert(t *testing.T) {
	foo := map[string]int{
		"hello":   40,
		"goodbye": 60,
		"boop":    60,
	}

	inverted := Invert(foo)
	require.Equal(t, inverted[40], []string{"hello"})
	require.Len(t, inverted[60], 2)
	require.Contains(t, inverted[60], "goodbye")
	require.Contains(t, inverted[60], "boop")
}
