package queue

import (
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

	"gitlab.com/trum/noteo/internal/domain"
)

var (
	// ErrQueueFull is returned when the queue is at capacity
	ErrQueueFull = errors.New("message queue is full")
)

// MessageSender is an interface for sending messages
type MessageSender interface {
	// SendMessage sends a message
	SendMessage(msg domain.Message) error
}

// Queue is an in-memory message queue
type Queue struct {
	config        *Config
	messageSender MessageSender
	messages      chan domain.Message
	wg            sync.WaitGroup
	stopCh        chan struct{}
}

// NewQueue creates a new message queue with the specified configuration
func NewQueue(cfg *Config, sender MessageSender) *Queue {
	return &Queue{
		messages:      make(chan domain.Message, cfg.Capacity),
		stopCh:        make(chan struct{}),
		config:        cfg,
		messageSender: sender,
	}
}

// Put adds a message to the queue, returning immediately
// Returns ErrQueueFull if the queue is at capacity
func (q *Queue) Put(msg domain.Message) error {
	select {
	case q.messages <- msg:
		return nil
	default:
		return ErrQueueFull
	}
}

// Start begins processing messages from the queue
func (q *Queue) Start() {
	slog.Info("Starting message queue", "capacity", q.config.Capacity)
	if q.messageSender == nil {
		slog.Error("Message sender not set")
		panic("message sender not set")
	}

	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		for msg := range q.messages {
			if err := q.sendWithRetry(msg); err != nil {
				slog.Error("Failed to send message after all retries, exiting application",
					"error", err,
					"chatId", msg.UserID)
				os.Exit(1)
			}
		}
	}()
}

// sendWithRetry attempts to send a message with exponential backoff retries
func (q *Queue) sendWithRetry(msg domain.Message) error {
	var err error
	delay := q.config.InitialRetryDelay

	for attempt := 0; attempt < q.config.MaxRetries; attempt++ {
		// Try to send the message
		err = q.messageSender.SendMessage(msg)
		if err == nil {
			return nil // Success!
		}

		// Log the error and prepare for retry
		slog.Warn("Failed to send message, will retry",
			"error", err,
			"attempt", attempt+1,
			"maxRetries", q.config.MaxRetries,
			"nextRetryDelay", delay,
			"chatId", msg.UserID)

		// Wait for the delay or until the queue is stopped
		select {
		case <-time.After(delay):
			// Continue with retry
		case <-q.stopCh:
			return errors.New("queue stopped during retry")
		}

		// Exponential backoff: double the delay for next attempt
		delay *= 2
		if delay > q.config.MaxRetryDelay {
			delay = q.config.MaxRetryDelay
		}
	}

	return err // Return the last error after all retries
}

// Stop gracefully shuts down the queue, waiting for all messages to be processed
func (q *Queue) Stop() {
	close(q.stopCh)   // Signal all retries to stop
	close(q.messages) // Stop accepting new messages
	q.wg.Wait()       // Wait for all processing to complete
	slog.Info("Message queue stopped")
}
