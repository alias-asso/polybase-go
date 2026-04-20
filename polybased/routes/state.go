package routes

import (
	"sync"
	"time"
)

var (
	stateStore    = make(map[string]time.Time)
	stateMutex    = &sync.Mutex{}
	cleanupTicker *time.Ticker
)

// startStateCleanupRoutine begins a periodic cleanup goroutine
func startStateCleanupRoutine() {
	cleanupTicker = time.NewTicker(1 * time.Minute)
	go func() {
		for range cleanupTicker.C {
			cleanupOIDCState()
		}
	}()
}

// cleanupOIDCState removes all expired entries from the state store
func cleanupOIDCState() {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	now := time.Now()
	for state, expiry := range stateStore {
		if now.After(expiry) {
			delete(stateStore, state)
		}
	}
}

// setOIDCState stores a state with 5-minute expiry
func setOIDCState(state string) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	stateStore[state] = time.Now().Add(5 * time.Minute)
}

// validOIDCState checks if state exists and is not expired, then deletes it
func validOIDCState(state string) bool {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	expiry, exists := stateStore[state]
	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		delete(stateStore, state)
		return false
	}

	delete(stateStore, state)
	return true
}
