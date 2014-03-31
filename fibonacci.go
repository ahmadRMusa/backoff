package backoff

import (
	"time"
)

type FibonacciBackoff struct {
	Retries    int
	MaxRetries int
	Delay      time.Duration
	Interval   time.Duration // time.Second, time.Millisecond, etc.
	Slots      []time.Duration
}

// Fibonacci creates a new instance of FibonacciBackoff.
func Fibonacci() *FibonacciBackoff {
	return &FibonacciBackoff{
		Retries:    0,
		MaxRetries: 5,
		Delay:      time.Duration(0),
	}
}

/*
Next gets the next backoff delay. This method will increment the retries and check
if the maximum number of retries has been met. If this condition is satisfied, then
the function will return. Otherwise, the next backoff delay will be computed.

The fibonacci backoff delay is computed as follows:
`n = fib(c - 1) + fib(c - 2); f(0) = 0, f(1) = 1; n >= 0.` where
`n` is the backoff delay and `c` is the retry slot.

This method maintains a slice of time.Duration to save on fibonacci computation.

Example, given a 1 second interval:

  Retry #        Backoff delay (in seconds)
    0                   0
    1                   1
    2                   1
    3                   2
    4                   3
    5                   5
    6                   8
    7                   13
*/
func (fb *FibonacciBackoff) Next() bool {
	fb.Retries++

	if fb.Retries >= fb.MaxRetries {
		return false
	}

	if fb.Retries == 1 {
		fb.Slots = append(fb.Slots, time.Duration(0*fb.Interval))
		fb.Slots = append(fb.Slots, time.Duration(1*fb.Interval))
		fb.Delay = time.Duration(1 * fb.Interval)
	} else {
		fb.Delay = fb.Slots[fb.Retries-1] + fb.Slots[fb.Retries-2]
		fb.Slots = append(fb.Slots, fb.Delay)
	}

	return true
}

/*
Retry will retry a function until the maximum number of retries is met. This method expects
the function `f` to return an error. If the failure condition is met, this method
will surface the error outputted from `f`, otherwise nil will be returned as normal.
*/
func (fb *FibonacciBackoff) Retry(f func() error) error {
	err := f()

	if err == nil {
		return nil
	}

	for fb.Next() {
		if err := f(); err == nil {
			return nil
		}

		time.Sleep(fb.Delay)
	}

	return err
}

// Reset will reset the retry count, the backoff delay, and backoff slots back to its initial state.
func (fb *FibonacciBackoff) Reset() {
	fb.Retries = 0
	fb.Delay = time.Duration(0 * time.Second)
	fb.Slots = nil
	fb.Slots = append(fb.Slots, time.Duration(0*fb.Interval))
}
