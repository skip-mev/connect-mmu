package gecko

import (
	"sync"
	"time"
)

type APIRateLimiter struct {
	// maxCalls is the amount of calls you are able to make in the given interval.
	maxCalls int
	// calls is the current amount of calls made.
	calls int
	// interval is the amount of time you are allowed to make maxCalls.
	interval time.Duration
	// timeStarted is the time in which calls=1 was made.
	timeStarted *time.Time

	mut *sync.Mutex
}

// newRateLimiter returns a type that keeps track of calls made within a given duration.
// this is useful for services that have rate limits.
func newRateLimiter(maxCalls int, interval time.Duration) *APIRateLimiter {
	return &APIRateLimiter{
		maxCalls:    maxCalls,
		calls:       0,
		interval:    interval,
		timeStarted: nil,
		mut:         &sync.Mutex{},
	}
}

// WaitForNextAvailableCall will check if you are safe to make another call. If you've reached the limit, this function
// will sleep until the interval has completed.
// For example, if you can only make 10 calls every 30 seconds,
// and you are at call 11, with start time = 1:30:20, and current time 1:30:49,
// this function will sleep for one second, then reset.
//
// This function is safe for concurrent use.
func (rl *APIRateLimiter) WaitForNextAvailableCall() {
	rl.mut.Lock()
	defer rl.mut.Unlock()
	now := time.Now()

	// if this is the first call, just set and return.
	if rl.timeStarted == nil {
		rl.timeStarted = &now
		rl.calls++
		return
	}

	// if we're within the window of opportunity
	if window := now.Sub(*rl.timeStarted); window <= rl.interval {
		// but we've reached max calls, we need to sleep. we'll reset later.
		if rl.calls+1 > rl.maxCalls {
			// sleep the amount of time it will take to get to the next window of opportunity.
			time.Sleep(rl.timeStarted.Add(rl.interval).Sub(now))
		} else {
			// we're not within max calls? we simply inc and return.
			rl.calls++
			return
		}
	}

	// reset. we're now in a new window of opportunity.
	rl.calls = 1
	rl.timeStarted = &now
}
