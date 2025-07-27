package internal

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/stretchr/testify/assert"
)

func TestShutdownManager_Creation(t *testing.T) {
	// Initialize logger for tests
	logger.InitLogger("error")
	defer logger.CloseLogger()
	
	sm := NewShutdownManager()
	defer sm.Cleanup()
	
	assert.NotNil(t, sm)
	assert.False(t, sm.IsShuttingDown())
	assert.NotNil(t, sm.Context())
}

func TestShutdownManager_ShutdownHooks(t *testing.T) {
	logger.InitLogger("error")
	defer logger.CloseLogger()
	
	sm := NewShutdownManager()
	defer sm.Cleanup()
	
	// Track hook execution
	hookExecuted := false
	
	sm.AddShutdownHook(func() error {
		hookExecuted = true
		return nil
	})
	
	// Trigger shutdown manually
	sm.InitiateShutdown()
	sm.WaitForCompletion()
	
	assert.True(t, sm.IsShuttingDown())
	assert.True(t, hookExecuted)
}

func TestShutdownManager_MultipleHooks(t *testing.T) {
	logger.InitLogger("error")
	defer logger.CloseLogger()
	
	sm := NewShutdownManager()
	defer sm.Cleanup()
	
	// Track hook execution order
	executionOrder := make([]int, 0)
	
	sm.AddShutdownHook(func() error {
		executionOrder = append(executionOrder, 1)
		return nil
	})
	
	sm.AddShutdownHook(func() error {
		executionOrder = append(executionOrder, 2)
		return nil
	})
	
	sm.AddShutdownHook(func() error {
		executionOrder = append(executionOrder, 3)
		return nil
	})
	
	// Trigger shutdown
	sm.InitiateShutdown()
	sm.WaitForCompletion()
	
	// Hooks should execute in order
	assert.Equal(t, []int{1, 2, 3}, executionOrder)
}

func TestShutdownManager_ContextCancellation(t *testing.T) {
	logger.InitLogger("error")
	defer logger.CloseLogger()
	
	sm := NewShutdownManager()
	defer sm.Cleanup()
	
	ctx := sm.Context()
	
	// Context should be active initially
	select {
	case <-ctx.Done():
		t.Error("Context should not be cancelled initially")
	default:
		// Good, context is not cancelled
	}
	
	// Trigger shutdown
	sm.InitiateShutdown()
	
	// Wait a bit for shutdown to process
	time.Sleep(100 * time.Millisecond)
	
	// Context should be cancelled after shutdown
	select {
	case <-ctx.Done():
		// Good, context is cancelled
	case <-time.After(1 * time.Second):
		t.Error("Context should be cancelled after shutdown")
	}
}

func TestShutdownManager_TimeoutSettings(t *testing.T) {
	logger.InitLogger("error")
	defer logger.CloseLogger()
	
	sm := NewShutdownManager()
	defer sm.Cleanup()
	
	// Test setting custom timeouts
	sm.SetGracePeriod(5 * time.Second)
	sm.SetForceTimeout(2 * time.Second)
	
	// Values should be set (we can't directly access them, but we can test they don't panic)
	assert.NotPanics(t, func() {
		sm.InitiateShutdown()
		sm.WaitForCompletion()
	})
}

func TestShutdownManager_HookError(t *testing.T) {
	logger.InitLogger("error")
	defer logger.CloseLogger()
	
	sm := NewShutdownManager()
	defer sm.Cleanup()
	
	hookWithErrorExecuted := false
	normalHookExecuted := false
	
	// Add hook that returns error
	sm.AddShutdownHook(func() error {
		hookWithErrorExecuted = true
		return NewShutdownError("test-component", assert.AnError)
	})
	
	// Add normal hook
	sm.AddShutdownHook(func() error {
		normalHookExecuted = true
		return nil
	})
	
	// Trigger shutdown
	sm.InitiateShutdown()
	sm.WaitForCompletion()
	
	// Both hooks should execute despite error
	assert.True(t, hookWithErrorExecuted)
	assert.True(t, normalHookExecuted)
}

func TestShutdownManager_SignalHandling(t *testing.T) {
	logger.InitLogger("error")
	defer logger.CloseLogger()
	
	sm := NewShutdownManager()
	defer sm.Cleanup()
	
	hookExecuted := false
	
	sm.AddShutdownHook(func() error {
		hookExecuted = true
		return nil
	})
	
	// Start signal monitoring in background
	go sm.WaitForShutdown()
	
	// Give it time to set up signal handling
	time.Sleep(50 * time.Millisecond)
	
	// Send SIGTERM to current process
	process, err := os.FindProcess(os.Getpid())
	assert.NoError(t, err)
	
	err = process.Signal(syscall.SIGTERM)
	assert.NoError(t, err)
	
	// Wait for shutdown to complete
	sm.WaitForCompletion()
	
	assert.True(t, sm.IsShuttingDown())
	assert.True(t, hookExecuted)
}

func TestShutdownManager_Cleanup(t *testing.T) {
	logger.InitLogger("error")
	defer logger.CloseLogger()
	
	sm := NewShutdownManager()
	
	// Test cleanup doesn't panic
	assert.NotPanics(t, func() {
		sm.Cleanup()
	})
}

func TestShutdownError_Error(t *testing.T) {
	err := NewShutdownError("test-component", assert.AnError)
	
	expected := "shutdown error in test-component: " + assert.AnError.Error()
	assert.Equal(t, expected, err.Error())
}