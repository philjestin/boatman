package agent

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadSession(t *testing.T) {
	// Create a test session
	session := NewSession("test-persist-session", "/tmp/test-project")
	session.Model = "sonnet"
	session.conversationID = "conv-123"
	session.Status = SessionStatusIdle

	// Add some messages
	session.Messages = append(session.Messages, Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Hello",
		Timestamp: time.Now(),
	})

	session.Messages = append(session.Messages, Message{
		ID:        "msg-2",
		Role:      "assistant",
		Content:   "Hi there!",
		Timestamp: time.Now(),
	})

	// Add some tasks
	session.Tasks = append(session.Tasks, Task{
		ID:          "task-1",
		Subject:     "Test task",
		Description: "A test task",
		Status:      "pending",
	})

	// Save the session
	err := SaveSession(session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Load the session
	loaded, err := LoadSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	// Verify loaded session matches original
	if loaded.ID != session.ID {
		t.Errorf("ID mismatch: expected %s, got %s", session.ID, loaded.ID)
	}

	if loaded.ProjectPath != session.ProjectPath {
		t.Errorf("ProjectPath mismatch: expected %s, got %s", session.ProjectPath, loaded.ProjectPath)
	}

	if loaded.Model != session.Model {
		t.Errorf("Model mismatch: expected %s, got %s", session.Model, loaded.Model)
	}

	if loaded.conversationID != session.conversationID {
		t.Errorf("ConversationID mismatch: expected %s, got %s", session.conversationID, loaded.conversationID)
	}

	if len(loaded.Messages) != len(session.Messages) {
		t.Errorf("Messages count mismatch: expected %d, got %d", len(session.Messages), len(loaded.Messages))
	}

	if len(loaded.Tasks) != len(session.Tasks) {
		t.Errorf("Tasks count mismatch: expected %d, got %d", len(session.Tasks), len(loaded.Tasks))
	}

	// Verify loaded session has context initialized
	if loaded.ctx == nil {
		t.Error("Loaded session should have context initialized")
	}

	if loaded.cancel == nil {
		t.Error("Loaded session should have cancel function initialized")
	}

	// Cleanup
	DeleteSessionFile(session.ID)
}

func TestLoadSession_InitializesAgents(t *testing.T) {
	session := NewSession("test-agents", "/tmp/test")

	// Save session
	err := SaveSession(session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Load session
	loaded, err := LoadSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	// Verify agents are initialized
	if loaded.agents == nil {
		t.Fatal("Agents map should be initialized")
	}

	if loaded.currentAgentID == "" {
		t.Error("Current agent ID should be set")
	}

	if _, ok := loaded.agents["main"]; !ok {
		t.Error("Main agent should exist")
	}

	// Cleanup
	DeleteSessionFile(session.ID)
}

func TestLoadSession_ResetsStoppedStatus(t *testing.T) {
	session := NewSession("test-stopped", "/tmp/test")
	session.Status = SessionStatusStopped

	// Save session with stopped status
	err := SaveSession(session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Load session
	loaded, err := LoadSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	// Verify status is reset to idle
	if loaded.Status != SessionStatusIdle {
		t.Errorf("Expected status idle, got %s", loaded.Status)
	}

	// Cleanup
	DeleteSessionFile(session.ID)
}

func TestLoadSession_ResetsErrorStatus(t *testing.T) {
	session := NewSession("test-error", "/tmp/test")
	session.Status = SessionStatusError

	// Save session with error status
	err := SaveSession(session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Load session
	loaded, err := LoadSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	// Verify status is reset to idle
	if loaded.Status != SessionStatusIdle {
		t.Errorf("Expected status idle, got %s", loaded.Status)
	}

	// Cleanup
	DeleteSessionFile(session.ID)
}

func TestLoadAllSessions(t *testing.T) {
	// Create multiple test sessions
	session1 := NewSession("test-all-1", "/tmp/test1")
	session2 := NewSession("test-all-2", "/tmp/test2")

	// Save sessions
	if err := SaveSession(session1); err != nil {
		t.Fatalf("Failed to save session1: %v", err)
	}
	if err := SaveSession(session2); err != nil {
		t.Fatalf("Failed to save session2: %v", err)
	}

	// Load all sessions
	sessions, err := LoadAllSessions()
	if err != nil {
		t.Fatalf("Failed to load all sessions: %v", err)
	}

	// Verify we got our sessions
	if len(sessions) < 2 {
		t.Errorf("Expected at least 2 sessions, got %d", len(sessions))
	}

	// Check our sessions are present
	found1, found2 := false, false
	for _, s := range sessions {
		if s.ID == session1.ID {
			found1 = true
		}
		if s.ID == session2.ID {
			found2 = true
		}
	}

	if !found1 {
		t.Error("Session 1 not found in loaded sessions")
	}
	if !found2 {
		t.Error("Session 2 not found in loaded sessions")
	}

	// Cleanup
	DeleteSessionFile(session1.ID)
	DeleteSessionFile(session2.ID)
}

func TestDeleteSessionFile(t *testing.T) {
	session := NewSession("test-delete", "/tmp/test")

	// Save session
	if err := SaveSession(session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Verify file exists
	sessionsDir, _ := GetSessionsDir()
	filename := filepath.Join(sessionsDir, session.ID+".json")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatal("Session file should exist after saving")
	}

	// Delete session file
	if err := DeleteSessionFile(session.ID); err != nil {
		t.Fatalf("Failed to delete session file: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		t.Error("Session file should be deleted")
	}
}

func TestLoadSession_NonExistent(t *testing.T) {
	// Try to load non-existent session
	_, err := LoadSession("non-existent-session-id")
	if err == nil {
		t.Error("Expected error when loading non-existent session")
	}
}

func TestGetSessionsDir(t *testing.T) {
	dir, err := GetSessionsDir()
	if err != nil {
		t.Fatalf("Failed to get sessions dir: %v", err)
	}

	if dir == "" {
		t.Error("Sessions directory should not be empty")
	}

	// Verify directory exists
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Sessions directory should exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("Sessions path should be a directory")
	}
}

func TestSaveSession_PreservesConversationID(t *testing.T) {
	session := NewSession("test-conv-id", "/tmp/test")
	session.conversationID = "my-conversation-123"

	if err := SaveSession(session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	loaded, err := LoadSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	if loaded.conversationID != session.conversationID {
		t.Errorf("ConversationID not preserved: expected %s, got %s",
			session.conversationID, loaded.conversationID)
	}

	DeleteSessionFile(session.ID)
}
