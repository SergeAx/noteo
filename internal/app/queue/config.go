package queue

import "time"

// Config holds configuration for the message queue
type Config struct {
	Capacity          int
	InitialRetryDelay time.Duration
	MaxRetryDelay     time.Duration
	MaxRetries        int
}
