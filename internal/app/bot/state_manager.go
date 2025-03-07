package bot

import (
	"sync"
	"time"
)

// UserState represents the current state of user interaction
type UserState int

const (
	StateNone UserState = iota
	StateCreatingProject
	StateMutingSubscription
	StateSuspendingSubscription
	StateUnsubscribing
	StateCustomMuteDuration
	StateCustomSuspendDuration
)

// UserContext stores the current state and data for a user
type UserContext struct {
	State     UserState
	CreatedAt time.Time
	Data      map[string]interface{} // For storing additional data like project ID
}

// StateManager handles all user state-related operations
type StateManager struct {
	userStates  map[int]UserContext
	stateMutex  sync.RWMutex
	cleanupDone chan struct{}
}

// NewStateManager creates a new instance of StateManager
func NewStateManager() *StateManager {
	sm := &StateManager{
		userStates:  make(map[int]UserContext),
		cleanupDone: make(chan struct{}),
	}
	go sm.cleanupStaleStates()
	return sm
}

// GetState retrieves the current state for a user
func (sm *StateManager) GetState(userID int) (UserState, map[string]interface{}, bool) {
	sm.stateMutex.RLock()
	defer sm.stateMutex.RUnlock()
	ctx, exists := sm.userStates[userID]
	if !exists {
		return StateNone, nil, false
	}
	return ctx.State, ctx.Data, true
}

// SetState sets the state for a user
func (sm *StateManager) SetState(userID int, state UserState, data map[string]interface{}) {
	sm.stateMutex.Lock()
	defer sm.stateMutex.Unlock()
	sm.userStates[userID] = UserContext{
		State:     state,
		CreatedAt: time.Now(),
		Data:      data,
	}
}

// ClearState removes the state for a user
func (sm *StateManager) ClearState(userID int) {
	sm.stateMutex.Lock()
	defer sm.stateMutex.Unlock()
	delete(sm.userStates, userID)
}

// cleanupStaleStates removes user states that are older than 5 minutes
func (sm *StateManager) cleanupStaleStates() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.stateMutex.Lock()
			now := time.Now()
			for userID, state := range sm.userStates {
				if now.Sub(state.CreatedAt) > 5*time.Minute {
					delete(sm.userStates, userID)
				}
			}
			sm.stateMutex.Unlock()
		case <-sm.cleanupDone:
			return
		}
	}
}

// Stop gracefully stops the state manager's background processes
func (sm *StateManager) Stop() {
	close(sm.cleanupDone)
}
