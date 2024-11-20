package gecko

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRateLimiter(t *testing.T) {
	// max calls is 3. we're going to make 4 calls in quick succession
	// which should make the limiter wait, reset, and then we'll be at 1 again.
	maximum := 3
	madeCalls := 4
	now := time.Now()
	limiter := newRateLimiter(maximum, 3*time.Second)
	for range madeCalls {
		limiter.WaitForNextAvailableCall()
	}
	require.Equal(t, limiter.calls, madeCalls-maximum)
	later := time.Now()
	delta := later.Sub(now)
	// should've taken around 3 seconds as it will sleep the interval off.
	require.InDelta(t, 3.0, delta.Seconds(), 0.1)
}
