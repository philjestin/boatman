package agent

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSendMessage_NoDeadlock(t *testing.T) {
	session := NewSession("test-session", "/tmp/test")
	session.ctx, session.cancel = context.WithCancel(context.Background())
	defer session.cancel()

	// Track if handlers are called
	var messageHandlerCalled, statusHandlerCalled bool
	var mu sync.Mutex

	// Set up handlers that would cause deadlock if called while lock is held
	session.SetMessageHandler(func(msg Message) {
		mu.Lock()
		messageHandlerCalled = true
		mu.Unlock()
		// This would deadlock if session lock is still held
		time.Sleep(10 * time.Millisecond)
	})

	session.SetStatusHandler(func(status SessionStatus) {
		mu.Lock()
		statusHandlerCalled = true
		mu.Unlock()
		// This would deadlock if session lock is still held
		time.Sleep(10 * time.Millisecond)
	})

	// Send message - should not deadlock
	done := make(chan bool)
	go func() {
		err := session.SendMessage("test message", AuthConfig{
			Method:       "anthropic-api",
			APIKey:       "test-key",
			ApprovalMode: "full-auto",
		})
		if err != nil {
			t.Errorf("SendMessage failed: %v", err)
		}
		done <- true
	}()

	// Wait with timeout
	select {
	case <-done:
		// Success - no deadlock
	case <-time.After(2 * time.Second):
		t.Fatal("SendMessage deadlocked")
	}

	// Give handlers time to complete
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if !messageHandlerCalled {
		t.Error("Message handler was not called")
	}
	if !statusHandlerCalled {
		t.Error("Status handler was not called")
	}
	mu.Unlock()

	// Verify message was added
	session.mu.RLock()
	if len(session.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(session.Messages))
	}
	if session.Status != SessionStatusRunning {
		t.Errorf("Expected status running, got %s", session.Status)
	}
	session.mu.RUnlock()
}

func TestSendMessage_StoppedSession(t *testing.T) {
	session := NewSession("test-session", "/tmp/test")
	session.ctx, session.cancel = context.WithCancel(context.Background())
	defer session.cancel()

	// Stop the session
	session.Stop()

	// Try to send message - should fail
	err := session.SendMessage("test", AuthConfig{
		Method: "anthropic-api",
		APIKey: "test-key",
	})

	if err == nil {
		t.Error("Expected error when sending to stopped session")
	}
	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("Expected 'not available' error, got: %v", err)
	}
}

func TestSendMessage_NoContext(t *testing.T) {
	session := NewSession("test-session", "/tmp/test")
	// Don't set context

	err := session.SendMessage("test", AuthConfig{
		Method: "anthropic-api",
		APIKey: "test-key",
	})

	if err == nil {
		t.Error("Expected error when context not initialized")
	}
	if !strings.Contains(err.Error(), "context not initialized") {
		t.Errorf("Expected 'context not initialized' error, got: %v", err)
	}
}

func TestSendMessage_AddsUserMessage(t *testing.T) {
	session := NewSession("test-session", "/tmp/test")
	session.ctx, session.cancel = context.WithCancel(context.Background())
	defer session.cancel()

	content := "Hello, Claude!"
	err := session.SendMessage(content, AuthConfig{
		Method:       "anthropic-api",
		APIKey:       "test-key",
		ApprovalMode: "full-auto",
	})

	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	// Give time for message to be added
	time.Sleep(50 * time.Millisecond)

	session.mu.RLock()
	defer session.mu.RUnlock()

	if len(session.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(session.Messages))
	}

	msg := session.Messages[0]
	if msg.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", msg.Role)
	}
	if msg.Content != content {
		t.Errorf("Expected content '%s', got '%s'", content, msg.Content)
	}
	if msg.ID == "" {
		t.Error("Message ID should not be empty")
	}
}

func TestSendMessage_SetsStatusToRunning(t *testing.T) {
	session := NewSession("test-session", "/tmp/test")
	session.ctx, session.cancel = context.WithCancel(context.Background())
	defer session.cancel()

	var statusChanges []SessionStatus
	var mu sync.Mutex

	session.SetStatusHandler(func(status SessionStatus) {
		mu.Lock()
		statusChanges = append(statusChanges, status)
		mu.Unlock()
	})

	err := session.SendMessage("test", AuthConfig{
		Method:       "anthropic-api",
		APIKey:       "test-key",
		ApprovalMode: "full-auto",
	})

	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	// Give time for status to be set
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(statusChanges) == 0 {
		t.Fatal("Status handler was not called")
	}

	if statusChanges[0] != SessionStatusRunning {
		t.Errorf("Expected status running, got %s", statusChanges[0])
	}
}

func TestSendMessage_CallsMessageHandler(t *testing.T) {
	session := NewSession("test-session", "/tmp/test")
	session.ctx, session.cancel = context.WithCancel(context.Background())
	defer session.cancel()

	var receivedMessages []Message
	var mu sync.Mutex

	session.SetMessageHandler(func(msg Message) {
		mu.Lock()
		receivedMessages = append(receivedMessages, msg)
		mu.Unlock()
	})

	content := "Test message"
	err := session.SendMessage(content, AuthConfig{
		Method:       "anthropic-api",
		APIKey:       "test-key",
		ApprovalMode: "full-auto",
	})

	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	// Give time for handler to be called
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(receivedMessages) != 1 {
		t.Fatalf("Expected 1 message in handler, got %d", len(receivedMessages))
	}

	if receivedMessages[0].Content != content {
		t.Errorf("Expected content '%s', got '%s'", content, receivedMessages[0].Content)
	}
}

func TestSessionHandlers_CanSafelyAccessSession(t *testing.T) {
	session := NewSession("test-session", "/tmp/test")
	session.ctx, session.cancel = context.WithCancel(context.Background())
	defer session.cancel()

	// Handler that reads session data (simulating SaveSession behavior)
	session.SetMessageHandler(func(msg Message) {
		session.mu.RLock()
		defer session.mu.RUnlock()
		// Simulate reading session data
		_ = session.Status
		_ = len(session.Messages)
	})

	// This should not deadlock
	done := make(chan bool)
	go func() {
		err := session.SendMessage("test", AuthConfig{
			Method:       "anthropic-api",
			APIKey:       "test-key",
			ApprovalMode: "full-auto",
		})
		if err != nil {
			t.Errorf("SendMessage failed: %v", err)
		}
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Deadlock detected - handler cannot access session")
	}
}
