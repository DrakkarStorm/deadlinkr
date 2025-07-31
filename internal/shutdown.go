package internal

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/DrakkarStorm/deadlinkr/logger"
)

// ShutdownManager handles graceful shutdown of the application
type ShutdownManager struct {
	ctx              context.Context
	cancel           context.CancelFunc
	shutdownSignal   chan os.Signal
	shutdownComplete chan bool
	isShuttingDown   bool
	mutex            sync.RWMutex
	
	// Shutdown hooks
	shutdownHooks    []func() error
	hooksMutex       sync.Mutex
	
	// Timeouts
	gracePeriod      time.Duration
	forceTimeout     time.Duration
}

// NewShutdownManager creates a new shutdown manager
func NewShutdownManager() *ShutdownManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	sm := &ShutdownManager{
		ctx:              ctx,
		cancel:           cancel,
		shutdownSignal:   make(chan os.Signal, 1),
		shutdownComplete: make(chan bool, 1),
		isShuttingDown:   false,
		shutdownHooks:    make([]func() error, 0),
		gracePeriod:      30 * time.Second, // 30 seconds for graceful shutdown
		forceTimeout:     5 * time.Second,  // 5 seconds for forced shutdown
	}
	
	// Register signal handlers
	signal.Notify(sm.shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	
	return sm
}

// Context returns the shutdown context
func (sm *ShutdownManager) Context() context.Context {
	return sm.ctx
}

// IsShuttingDown returns true if shutdown is in progress
func (sm *ShutdownManager) IsShuttingDown() bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.isShuttingDown
}

// AddShutdownHook adds a function to be called during shutdown
func (sm *ShutdownManager) AddShutdownHook(hook func() error) {
	sm.hooksMutex.Lock()
	defer sm.hooksMutex.Unlock()
	sm.shutdownHooks = append(sm.shutdownHooks, hook)
}

// WaitForShutdown blocks until a shutdown signal is received
func (sm *ShutdownManager) WaitForShutdown() {
	select {
	case sig := <-sm.shutdownSignal:
		logger.Infof("Received shutdown signal: %v", sig)
		sm.initiateShutdown()
	case <-sm.ctx.Done():
		logger.Debugf("Shutdown context cancelled")
	}
}

// InitiateShutdown manually triggers shutdown (for testing or programmatic shutdown)
func (sm *ShutdownManager) InitiateShutdown() {
	sm.initiateShutdown()
}

// initiateShutdown starts the graceful shutdown process
func (sm *ShutdownManager) initiateShutdown() {
	sm.mutex.Lock()
	if sm.isShuttingDown {
		sm.mutex.Unlock()
		return // Already shutting down
	}
	sm.isShuttingDown = true
	sm.mutex.Unlock()
	
	logger.Infof("Initiating graceful shutdown...")
	
	// Start shutdown process in goroutine
	go sm.performShutdown()
}

// performShutdown executes the shutdown sequence
func (sm *ShutdownManager) performShutdown() {
	defer func() {
		sm.shutdownComplete <- true
	}()
	
	// Create timeout context for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), sm.gracePeriod)
	defer shutdownCancel()
	
	// Execute shutdown hooks
	hookComplete := make(chan bool, 1)
	go func() {
		sm.executeShutdownHooks()
		hookComplete <- true
	}()
	
	// Wait for hooks to complete or timeout
	select {
	case <-hookComplete:
		logger.Infof("All shutdown hooks completed successfully")
	case <-shutdownCtx.Done():
		logger.Warnf("Shutdown hooks timed out, forcing shutdown")
	}
	
	// Cancel the main context to signal all components
	sm.cancel()
	
	// Wait a bit for components to clean up
	select {
	case <-time.After(sm.forceTimeout):
		logger.Warnf("Force shutdown timeout reached")
	case <-sm.ctx.Done():
		// Context already cancelled
	}
	
	logger.Infof("Shutdown completed")
}

// executeShutdownHooks runs all registered shutdown hooks
func (sm *ShutdownManager) executeShutdownHooks() {
	sm.hooksMutex.Lock()
	hooks := make([]func() error, len(sm.shutdownHooks))
	copy(hooks, sm.shutdownHooks)
	sm.hooksMutex.Unlock()
	
	logger.Infof("Executing %d shutdown hooks...", len(hooks))
	
	for i, hook := range hooks {
		if err := hook(); err != nil {
			logger.Errorf("Shutdown hook %d failed: %v", i+1, err)
		} else {
			logger.Debugf("Shutdown hook %d completed successfully", i+1)
		}
	}
}

// WaitForCompletion waits for shutdown to complete
func (sm *ShutdownManager) WaitForCompletion() {
	<-sm.shutdownComplete
}

// SetGracePeriod sets the graceful shutdown timeout
func (sm *ShutdownManager) SetGracePeriod(duration time.Duration) {
	sm.gracePeriod = duration
}

// SetForceTimeout sets the force shutdown timeout
func (sm *ShutdownManager) SetForceTimeout(duration time.Duration) {
	sm.forceTimeout = duration
}

// Cleanup performs final cleanup
func (sm *ShutdownManager) Cleanup() {
	// Stop signal handling
	signal.Stop(sm.shutdownSignal)
	close(sm.shutdownSignal)
	
	// Cancel context if not already cancelled
	sm.cancel()
	
	logger.Debugf("Shutdown manager cleanup completed")
}

// ShutdownError represents an error during shutdown
type ShutdownError struct {
	Component string
	Err       error
}

func (se ShutdownError) Error() string {
	return "shutdown error in " + se.Component + ": " + se.Err.Error()
}

// NewShutdownError creates a new shutdown error
func NewShutdownError(component string, err error) ShutdownError {
	return ShutdownError{
		Component: component,
		Err:       err,
	}
}