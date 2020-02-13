package kudo

import "time"

// WaitOption changes a WaitConfig
type WaitOption func(*WaitConfig)

// WaitTimeout sets the timeout of a instance wait call.
func WaitTimeout(timeout time.Duration) WaitOption {
	return func(config *WaitConfig) {
		config.Timeout = timeout
	}
}
