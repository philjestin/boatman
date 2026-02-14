package agent

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestNewManager tests the initialization of a new Manager
func TestNewManager(t *testing.T) {
	m := NewManager()

	if m == nil {
		t.Fatal("NewManager returned nil")
	}

	if m.sessions == nil {
		t.Error("sessions map is nil")
	}

	if len(m.sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(m.sessions))
	}

	if m.defaultModel != "sonnet" {
		t.Errorf("expected default model 'sonnet', got %s", m.defaultModel)
	}
}

// TestSetContext tests setting the Wails runtime context
func TestSetContext(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	m.SetContext(ctx)

	if m.ctx == nil {
		t.Error("context was not set")
	}
}

// TestSetDefaultModel tests setting the default model
func TestSetDefaultModel(t *testing.T) {
	m := NewManager()

	tests := []struct {
		name  string
		model string
	}{
		{"opus model", "opus"},
		{"sonnet model", "sonnet"},
		{"haiku model", "haiku"},
		{"custom model", "claude-3-5-sonnet-20241022"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.SetDefaultModel(tt.model)

			m.mu.RLock()
			actual := m.defaultModel
			m.mu.RUnlock()

			if actual != tt.model {
				t.Errorf("expected model %s, got %s", tt.model, actual)
			}
		})
	}
}

// TestSetAuthConfigGetter tests setting the auth config getter
func TestSetAuthConfigGetter(t *testing.T) {
	m := NewManager()

	expectedConfig := AuthConfig{
		Method:       "anthropic-api",
		APIKey:       "test-key",
		ApprovalMode: "suggest",
	}

	m.SetAuthConfigGetter(func() AuthConfig {
		return expectedConfig
	})

	if m.authConfigGetter == nil {
		t.Fatal("authConfigGetter was not set")
	}

	actualConfig := m.authConfigGetter()
	if actualConfig.Method != expectedConfig.Method {
		t.Errorf("expected method %s, got %s", expectedConfig.Method, actualConfig.Method)
	}
	if actualConfig.APIKey != expectedConfig.APIKey {
		t.Errorf("expected API key %s, got %s", expectedConfig.APIKey, actualConfig.APIKey)
	}
}

// TestSetAPIKeyGetter tests the deprecated API key getter
func TestSetAPIKeyGetter(t *testing.T) {
	m := NewManager()
	expectedKey := "test-api-key"

	m.SetAPIKeyGetter(func() string {
		return expectedKey
	})

	if m.authConfigGetter == nil {
		t.Fatal("authConfigGetter was not set")
	}

	config := m.authConfigGetter()
	if config.Method != "anthropic-api" {
		t.Errorf("expected method 'anthropic-api', got %s", config.Method)
	}
	if config.APIKey != expectedKey {
		t.Errorf("expected API key %s, got %s", expectedKey, config.APIKey)
	}
}

// TestCreateSession tests creating new sessions
func TestCreateSession(t *testing.T) {
	tests := []struct {
		name        string
		projectPath string
	}{
		{"simple path", "/path/to/project"},
		{"home directory", "~/projects/test"},
		{"relative path", "./myproject"},
		{"windows path", "C:\\Users\\test\\project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			m.SetContext(context.Background())

			session, err := m.CreateSession(tt.projectPath)

			if err != nil {
				t.Fatalf("CreateSession failed: %v", err)
			}

			if session == nil {
				t.Fatal("CreateSession returned nil session")
			}

			if session.ID == "" {
				t.Error("session ID is empty")
			}

			if session.ProjectPath != tt.projectPath {
				t.Errorf("expected project path %s, got %s", tt.projectPath, session.ProjectPath)
			}

			if session.Status != SessionStatusIdle {
				t.Errorf("expected status %s, got %s", SessionStatusIdle, session.Status)
			}

			// Verify session was added to manager
			m.mu.RLock()
			storedSession, exists := m.sessions[session.ID]
			m.mu.RUnlock()

			if !exists {
				t.Error("session was not added to manager")
			}

			if storedSession != session {
				t.Error("stored session does not match returned session")
			}
		})
	}
}

// TestCreateSessionEventHandlers tests that event handlers are set up correctly
func TestCreateSessionEventHandlers(t *testing.T) {
	m := NewManager()
	m.SetContext(context.Background())

	session, err := m.CreateSession("/test/path")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Verify handlers were set
	session.mu.RLock()
	hasMessageHandler := session.onMessage != nil
	hasTaskHandler := session.onTask != nil
	hasStatusHandler := session.onStatus != nil
	session.mu.RUnlock()

	if !hasMessageHandler {
		t.Error("message handler was not set")
	}
	if !hasTaskHandler {
		t.Error("task handler was not set")
	}
	if !hasStatusHandler {
		t.Error("status handler was not set")
	}
}

// TestGetSession tests retrieving sessions
func TestGetSession(t *testing.T) {
	m := NewManager()
	m.SetContext(context.Background())

	// Create a session
	session, err := m.CreateSession("/test/path")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	t.Run("existing session", func(t *testing.T) {
		retrieved, err := m.GetSession(session.ID)
		if err != nil {
			t.Fatalf("GetSession failed: %v", err)
		}

		if retrieved.ID != session.ID {
			t.Errorf("expected session ID %s, got %s", session.ID, retrieved.ID)
		}

		if retrieved != session {
			t.Error("retrieved session does not match original")
		}
	})

	t.Run("non-existing session", func(t *testing.T) {
		_, err := m.GetSession("non-existent-id")
		if err == nil {
			t.Error("expected error for non-existent session, got nil")
		}

		expectedError := "session not found: non-existent-id"
		if err.Error() != expectedError {
			t.Errorf("expected error %q, got %q", expectedError, err.Error())
		}
	})
}

// TestListSessions tests listing all sessions
func TestListSessions(t *testing.T) {
	m := NewManager()
	m.SetContext(context.Background())

	t.Run("empty list", func(t *testing.T) {
		sessions := m.ListSessions()
		if sessions == nil {
			t.Error("ListSessions returned nil")
		}
		if len(sessions) != 0 {
			t.Errorf("expected 0 sessions, got %d", len(sessions))
		}
	})

	t.Run("multiple sessions", func(t *testing.T) {
		// Create multiple sessions
		paths := []string{"/path1", "/path2", "/path3"}
		createdSessions := make([]*Session, 0, len(paths))

		for _, path := range paths {
			session, err := m.CreateSession(path)
			if err != nil {
				t.Fatalf("CreateSession failed: %v", err)
			}
			createdSessions = append(createdSessions, session)
		}

		sessions := m.ListSessions()
		if len(sessions) != len(paths) {
			t.Errorf("expected %d sessions, got %d", len(paths), len(sessions))
		}

		// Verify all sessions are present
		sessionMap := make(map[string]*Session)
		for _, s := range sessions {
			sessionMap[s.ID] = s
		}

		for _, created := range createdSessions {
			if _, exists := sessionMap[created.ID]; !exists {
				t.Errorf("created session %s not found in list", created.ID)
			}
		}
	})
}

// TestStartSession tests starting sessions
func TestStartSession(t *testing.T) {
	m := NewManager()
	// Don't set context to avoid Wails runtime errors in tests
	m.SetDefaultModel("opus")

	t.Run("success", func(t *testing.T) {
		session, err := m.CreateSession("/test/path")
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		err = m.StartSession(session.ID)
		if err != nil {
			t.Fatalf("StartSession failed: %v", err)
		}

		// Verify session is idle and model is set
		if session.Status != SessionStatusIdle {
			t.Errorf("expected status %s, got %s", SessionStatusIdle, session.Status)
		}

		if session.Model != "opus" {
			t.Errorf("expected model 'opus', got %s", session.Model)
		}
	})

	t.Run("non-existing session", func(t *testing.T) {
		err := m.StartSession("non-existent-id")
		if err == nil {
			t.Error("expected error for non-existent session, got nil")
		}
	})
}

// TestStopSession tests stopping sessions
func TestStopSession(t *testing.T) {
	m := NewManager()
	// Don't set context to avoid Wails runtime errors in tests

	t.Run("success", func(t *testing.T) {
		session, err := m.CreateSession("/test/path")
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		err = m.StartSession(session.ID)
		if err != nil {
			t.Fatalf("StartSession failed: %v", err)
		}

		err = m.StopSession(session.ID)
		if err != nil {
			t.Fatalf("StopSession failed: %v", err)
		}

		if session.Status != SessionStatusStopped {
			t.Errorf("expected status %s, got %s", SessionStatusStopped, session.Status)
		}
	})

	t.Run("non-existing session", func(t *testing.T) {
		err := m.StopSession("non-existent-id")
		if err == nil {
			t.Error("expected error for non-existent session, got nil")
		}
	})
}

// TestDeleteSession tests deleting sessions
func TestDeleteSession(t *testing.T) {
	m := NewManager()
	// Don't set context to avoid Wails runtime errors in tests

	t.Run("success", func(t *testing.T) {
		session, err := m.CreateSession("/test/path")
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		sessionID := session.ID

		err = m.DeleteSession(sessionID)
		if err != nil {
			t.Fatalf("DeleteSession failed: %v", err)
		}

		// Verify session was removed
		m.mu.RLock()
		_, exists := m.sessions[sessionID]
		m.mu.RUnlock()

		if exists {
			t.Error("session still exists after deletion")
		}

		// Verify GetSession returns error
		_, err = m.GetSession(sessionID)
		if err == nil {
			t.Error("GetSession should return error for deleted session")
		}
	})

	t.Run("non-existing session", func(t *testing.T) {
		err := m.DeleteSession("non-existent-id")
		if err == nil {
			t.Error("expected error for non-existent session, got nil")
		}
	})

	t.Run("stops session before deleting", func(t *testing.T) {
		session, err := m.CreateSession("/test/path")
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		err = m.StartSession(session.ID)
		if err != nil {
			t.Fatalf("StartSession failed: %v", err)
		}

		err = m.DeleteSession(session.ID)
		if err != nil {
			t.Fatalf("DeleteSession failed: %v", err)
		}

		// Session should be stopped
		if session.Status != SessionStatusStopped {
			t.Errorf("expected status %s, got %s", SessionStatusStopped, session.Status)
		}
	})
}

// TestSendMessage tests sending messages to sessions
func TestSendMessage(t *testing.T) {
	m := NewManager()
	// Don't set context to avoid Wails runtime errors in tests

	authConfig := AuthConfig{
		Method:       "anthropic-api",
		APIKey:       "test-key",
		ApprovalMode: "suggest",
	}
	m.SetAuthConfigGetter(func() AuthConfig {
		return authConfig
	})

	t.Run("non-existing session", func(t *testing.T) {
		err := m.SendMessage("non-existent-id", "test message")
		if err == nil {
			t.Error("expected error for non-existent session, got nil")
		}
	})

	t.Run("session not started", func(t *testing.T) {
		session, err := m.CreateSession("/test/path")
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		err = m.SendMessage(session.ID, "test message")
		if err == nil {
			t.Error("expected error for non-started session, got nil")
		}
	})
}

// TestApproveAction tests approving actions
func TestApproveAction(t *testing.T) {
	m := NewManager()
	// Don't set context to avoid Wails runtime errors in tests

	t.Run("non-existing session", func(t *testing.T) {
		err := m.ApproveAction("non-existent-id", "action-123")
		if err == nil {
			t.Error("expected error for non-existent session, got nil")
		}
	})

	t.Run("existing session", func(t *testing.T) {
		session, err := m.CreateSession("/test/path")
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		err = m.ApproveAction(session.ID, "action-123")
		// Expected to fail as approval is not supported in current implementation
		if err == nil {
			t.Error("expected error as approval is not supported")
		}
	})
}

// TestRejectAction tests rejecting actions
func TestRejectAction(t *testing.T) {
	m := NewManager()
	// Don't set context to avoid Wails runtime errors in tests

	t.Run("non-existing session", func(t *testing.T) {
		err := m.RejectAction("non-existent-id", "action-123")
		if err == nil {
			t.Error("expected error for non-existent session, got nil")
		}
	})

	t.Run("existing session", func(t *testing.T) {
		session, err := m.CreateSession("/test/path")
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		err = m.RejectAction(session.ID, "action-123")
		// Expected to fail as rejection is not supported in current implementation
		if err == nil {
			t.Error("expected error as rejection is not supported")
		}
	})
}

// TestGetSessionMessages tests retrieving session messages
func TestGetSessionMessages(t *testing.T) {
	m := NewManager()
	// Don't set context to avoid Wails runtime errors in tests

	t.Run("non-existing session", func(t *testing.T) {
		_, err := m.GetSessionMessages("non-existent-id")
		if err == nil {
			t.Error("expected error for non-existent session, got nil")
		}
	})

	t.Run("existing session with no messages", func(t *testing.T) {
		session, err := m.CreateSession("/test/path")
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		messages, err := m.GetSessionMessages(session.ID)
		if err != nil {
			t.Fatalf("GetSessionMessages failed: %v", err)
		}

		if messages == nil {
			t.Error("messages is nil")
		}

		if len(messages) != 0 {
			t.Errorf("expected 0 messages, got %d", len(messages))
		}
	})

	t.Run("existing session with messages", func(t *testing.T) {
		session, err := m.CreateSession("/test/path")
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		// Add some messages directly
		session.Messages = []Message{
			{ID: "1", Role: "user", Content: "Hello", Timestamp: time.Now()},
			{ID: "2", Role: "assistant", Content: "Hi", Timestamp: time.Now()},
		}

		messages, err := m.GetSessionMessages(session.ID)
		if err != nil {
			t.Fatalf("GetSessionMessages failed: %v", err)
		}

		if len(messages) != 2 {
			t.Errorf("expected 2 messages, got %d", len(messages))
		}
	})
}

// TestGetSessionTasks tests retrieving session tasks
func TestGetSessionTasks(t *testing.T) {
	m := NewManager()
	// Don't set context to avoid Wails runtime errors in tests

	t.Run("non-existing session", func(t *testing.T) {
		_, err := m.GetSessionTasks("non-existent-id")
		if err == nil {
			t.Error("expected error for non-existent session, got nil")
		}
	})

	t.Run("existing session with no tasks", func(t *testing.T) {
		session, err := m.CreateSession("/test/path")
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		tasks, err := m.GetSessionTasks(session.ID)
		if err != nil {
			t.Fatalf("GetSessionTasks failed: %v", err)
		}

		if tasks == nil {
			t.Error("tasks is nil")
		}

		if len(tasks) != 0 {
			t.Errorf("expected 0 tasks, got %d", len(tasks))
		}
	})

	t.Run("existing session with tasks", func(t *testing.T) {
		session, err := m.CreateSession("/test/path")
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}

		// Add some tasks directly
		session.Tasks = []Task{
			{ID: "1", Subject: "Task 1", Status: "pending"},
			{ID: "2", Subject: "Task 2", Status: "completed"},
		}

		tasks, err := m.GetSessionTasks(session.ID)
		if err != nil {
			t.Fatalf("GetSessionTasks failed: %v", err)
		}

		if len(tasks) != 2 {
			t.Errorf("expected 2 tasks, got %d", len(tasks))
		}
	})
}

// TestStopAllSessions tests stopping all sessions
func TestStopAllSessions(t *testing.T) {
	m := NewManager()
	// Don't set context to avoid Wails runtime errors in tests

	// Create multiple sessions
	sessions := make([]*Session, 3)
	for i := 0; i < 3; i++ {
		session, err := m.CreateSession("/test/path")
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}
		err = m.StartSession(session.ID)
		if err != nil {
			t.Fatalf("StartSession failed: %v", err)
		}
		sessions[i] = session
	}

	m.StopAllSessions()

	// Verify all sessions are stopped
	for i, session := range sessions {
		if session.Status != SessionStatusStopped {
			t.Errorf("session %d: expected status %s, got %s", i, SessionStatusStopped, session.Status)
		}
	}
}

// TestConcurrentSessionAccess tests concurrent access to sessions
func TestConcurrentSessionAccess(t *testing.T) {
	m := NewManager()
	// Don't set context to avoid Wails runtime errors in tests

	// Create initial sessions
	numSessions := 10
	sessionIDs := make([]string, numSessions)
	for i := 0; i < numSessions; i++ {
		session, err := m.CreateSession("/test/path")
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}
		sessionIDs[i] = session.ID
	}

	var wg sync.WaitGroup
	errors := make(chan error, numSessions*3)

	// Concurrent reads
	for i := 0; i < numSessions; i++ {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			_, err := m.GetSession(id)
			if err != nil {
				errors <- err
			}
		}(sessionIDs[i])
	}

	// Concurrent list operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sessions := m.ListSessions()
			if len(sessions) < numSessions {
				errors <- nil
			}
		}()
	}

	// Concurrent creates
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := m.CreateSession("/test/path")
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Errorf("concurrent operation failed: %v", err)
		}
	}
}

// TestMultiSessionManagement tests managing multiple sessions
func TestMultiSessionManagement(t *testing.T) {
	m := NewManager()
	m.SetContext(context.Background())

	// Create sessions for different projects
	projects := []string{"/project1", "/project2", "/project3"}
	sessionMap := make(map[string]*Session)

	for _, proj := range projects {
		session, err := m.CreateSession(proj)
		if err != nil {
			t.Fatalf("CreateSession failed for %s: %v", proj, err)
		}
		sessionMap[proj] = session
	}

	// Start some sessions
	err := m.StartSession(sessionMap["/project1"].ID)
	if err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	err = m.StartSession(sessionMap["/project2"].ID)
	if err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	// Verify session states
	if sessionMap["/project1"].Status != SessionStatusIdle {
		t.Errorf("project1: expected status %s, got %s", SessionStatusIdle, sessionMap["/project1"].Status)
	}

	if sessionMap["/project2"].Status != SessionStatusIdle {
		t.Errorf("project2: expected status %s, got %s", SessionStatusIdle, sessionMap["/project2"].Status)
	}

	if sessionMap["/project3"].Status != SessionStatusIdle {
		t.Errorf("project3: expected status %s, got %s", SessionStatusIdle, sessionMap["/project3"].Status)
	}

	// Stop one session
	err = m.StopSession(sessionMap["/project1"].ID)
	if err != nil {
		t.Fatalf("StopSession failed: %v", err)
	}

	if sessionMap["/project1"].Status != SessionStatusStopped {
		t.Errorf("project1: expected status %s after stop, got %s", SessionStatusStopped, sessionMap["/project1"].Status)
	}

	// Delete one session
	err = m.DeleteSession(sessionMap["/project3"].ID)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	// Verify remaining sessions
	sessions := m.ListSessions()
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions after deletion, got %d", len(sessions))
	}
}

// TestEventEmission tests that events are properly emitted
func TestEventEmission(t *testing.T) {
	m := NewManager()
	m.SetContext(context.Background())

	session, err := m.CreateSession("/test/path")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Test message handler
	t.Run("message handler", func(t *testing.T) {
		messageReceived := false
		var receivedMsg Message
		var mu sync.Mutex

		session.SetMessageHandler(func(msg Message) {
			mu.Lock()
			messageReceived = true
			receivedMsg = msg
			mu.Unlock()
		})

		// Simulate adding a message
		testMsg := Message{
			ID:        "test-1",
			Role:      "user",
			Content:   "test content",
			Timestamp: time.Now(),
		}

		session.mu.Lock()
		session.Messages = append(session.Messages, testMsg)
		if session.onMessage != nil {
			session.onMessage(testMsg)
		}
		session.mu.Unlock()

		// Wait a bit for async processing
		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		if !messageReceived {
			t.Error("message handler was not called")
		}
		if receivedMsg.ID != testMsg.ID {
			t.Errorf("expected message ID %s, got %s", testMsg.ID, receivedMsg.ID)
		}
		mu.Unlock()
	})

	// Test task handler
	t.Run("task handler", func(t *testing.T) {
		taskReceived := false
		var receivedTask Task
		var mu sync.Mutex

		session.SetTaskHandler(func(task Task) {
			mu.Lock()
			taskReceived = true
			receivedTask = task
			mu.Unlock()
		})

		// Simulate adding a task
		testTask := Task{
			ID:      "task-1",
			Subject: "Test task",
			Status:  "pending",
		}

		session.mu.Lock()
		session.Tasks = append(session.Tasks, testTask)
		if session.onTask != nil {
			session.onTask(testTask)
		}
		session.mu.Unlock()

		// Wait a bit for async processing
		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		if !taskReceived {
			t.Error("task handler was not called")
		}
		if receivedTask.ID != testTask.ID {
			t.Errorf("expected task ID %s, got %s", testTask.ID, receivedTask.ID)
		}
		mu.Unlock()
	})

	// Test status handler
	t.Run("status handler", func(t *testing.T) {
		statusReceived := false
		var receivedStatus SessionStatus
		var mu sync.Mutex

		session.SetStatusHandler(func(status SessionStatus) {
			mu.Lock()
			statusReceived = true
			receivedStatus = status
			mu.Unlock()
		})

		// Start the session to trigger status change
		err := m.StartSession(session.ID)
		if err != nil {
			t.Fatalf("StartSession failed: %v", err)
		}

		// Wait a bit for async processing
		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		if !statusReceived {
			t.Error("status handler was not called")
		}
		if receivedStatus != SessionStatusIdle {
			t.Errorf("expected status %s, got %s", SessionStatusIdle, receivedStatus)
		}
		mu.Unlock()
	})
}

// TestAuthConfigPropagation tests that auth config is properly propagated
func TestAuthConfigPropagation(t *testing.T) {
	tests := []struct {
		name   string
		config AuthConfig
	}{
		{
			name: "anthropic api",
			config: AuthConfig{
				Method:       "anthropic-api",
				APIKey:       "sk-test-key-123",
				ApprovalMode: "suggest",
			},
		},
		{
			name: "google cloud",
			config: AuthConfig{
				Method:       "google-cloud",
				GCPProjectID: "my-project",
				GCPRegion:    "us-central1",
				ApprovalMode: "auto-edit",
			},
		},
		{
			name: "full auto mode",
			config: AuthConfig{
				Method:       "anthropic-api",
				APIKey:       "sk-test-key-456",
				ApprovalMode: "full-auto",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			m.SetContext(context.Background())

			m.SetAuthConfigGetter(func() AuthConfig {
				return tt.config
			})

			// Verify the config is returned correctly
			m.mu.RLock()
			if m.authConfigGetter == nil {
				t.Fatal("auth config getter is nil")
			}
			retrievedConfig := m.authConfigGetter()
			m.mu.RUnlock()

			if retrievedConfig.Method != tt.config.Method {
				t.Errorf("expected method %s, got %s", tt.config.Method, retrievedConfig.Method)
			}
			if retrievedConfig.APIKey != tt.config.APIKey {
				t.Errorf("expected API key %s, got %s", tt.config.APIKey, retrievedConfig.APIKey)
			}
			if retrievedConfig.GCPProjectID != tt.config.GCPProjectID {
				t.Errorf("expected GCP project ID %s, got %s", tt.config.GCPProjectID, retrievedConfig.GCPProjectID)
			}
			if retrievedConfig.ApprovalMode != tt.config.ApprovalMode {
				t.Errorf("expected approval mode %s, got %s", tt.config.ApprovalMode, retrievedConfig.ApprovalMode)
			}
		})
	}
}

// TestSessionLifecycle tests the complete lifecycle of a session
func TestSessionLifecycle(t *testing.T) {
	m := NewManager()
	m.SetContext(context.Background())
	m.SetDefaultModel("sonnet")

	// 1. Create session
	session, err := m.CreateSession("/test/project")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if session.Status != SessionStatusIdle {
		t.Errorf("new session: expected status %s, got %s", SessionStatusIdle, session.Status)
	}

	sessionID := session.ID

	// 2. Verify it's in the list
	sessions := m.ListSessions()
	found := false
	for _, s := range sessions {
		if s.ID == sessionID {
			found = true
			break
		}
	}
	if !found {
		t.Error("newly created session not found in list")
	}

	// 3. Start session
	err = m.StartSession(sessionID)
	if err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	if session.Status != SessionStatusIdle {
		t.Errorf("started session: expected status %s, got %s", SessionStatusIdle, session.Status)
	}

	if session.Model != "sonnet" {
		t.Errorf("started session: expected model 'sonnet', got %s", session.Model)
	}

	// 4. Verify messages and tasks are accessible
	messages, err := m.GetSessionMessages(sessionID)
	if err != nil {
		t.Fatalf("GetSessionMessages failed: %v", err)
	}
	if messages == nil {
		t.Error("messages is nil")
	}

	tasks, err := m.GetSessionTasks(sessionID)
	if err != nil {
		t.Fatalf("GetSessionTasks failed: %v", err)
	}
	if tasks == nil {
		t.Error("tasks is nil")
	}

	// 5. Stop session
	err = m.StopSession(sessionID)
	if err != nil {
		t.Fatalf("StopSession failed: %v", err)
	}

	if session.Status != SessionStatusStopped {
		t.Errorf("stopped session: expected status %s, got %s", SessionStatusStopped, session.Status)
	}

	// 6. Session should still be in the list
	sessions = m.ListSessions()
	found = false
	for _, s := range sessions {
		if s.ID == sessionID {
			found = true
			break
		}
	}
	if !found {
		t.Error("stopped session not found in list")
	}

	// 7. Delete session
	err = m.DeleteSession(sessionID)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	// 8. Session should be removed from list
	sessions = m.ListSessions()
	found = false
	for _, s := range sessions {
		if s.ID == sessionID {
			found = true
			break
		}
	}
	if found {
		t.Error("deleted session still found in list")
	}

	// 9. GetSession should return error
	_, err = m.GetSession(sessionID)
	if err == nil {
		t.Error("GetSession should return error for deleted session")
	}
}

// TestThreadSafety tests thread safety of manager operations
func TestThreadSafety(t *testing.T) {
	m := NewManager()
	m.SetContext(context.Background())

	var wg sync.WaitGroup
	numGoroutines := 50
	sessionIDs := make(chan string, numGoroutines)

	// Concurrent creates
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			session, err := m.CreateSession("/test/path")
			if err != nil {
				t.Errorf("CreateSession failed: %v", err)
				return
			}
			sessionIDs <- session.ID
		}(i)
	}

	wg.Wait()
	close(sessionIDs)

	// Collect all session IDs
	var ids []string
	for id := range sessionIDs {
		ids = append(ids, id)
	}

	if len(ids) != numGoroutines {
		t.Errorf("expected %d sessions, got %d", numGoroutines, len(ids))
	}

	// Concurrent operations on sessions
	wg = sync.WaitGroup{}
	for _, id := range ids {
		// Start
		wg.Add(1)
		go func(sessionID string) {
			defer wg.Done()
			err := m.StartSession(sessionID)
			if err != nil {
				t.Errorf("StartSession failed: %v", err)
			}
		}(id)

		// GetMessages
		wg.Add(1)
		go func(sessionID string) {
			defer wg.Done()
			_, err := m.GetSessionMessages(sessionID)
			if err != nil {
				t.Errorf("GetSessionMessages failed: %v", err)
			}
		}(id)

		// GetTasks
		wg.Add(1)
		go func(sessionID string) {
			defer wg.Done()
			_, err := m.GetSessionTasks(sessionID)
			if err != nil {
				t.Errorf("GetSessionTasks failed: %v", err)
			}
		}(id)
	}

	// Concurrent lists
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sessions := m.ListSessions()
			if len(sessions) < numGoroutines {
				t.Errorf("expected at least %d sessions, got %d", numGoroutines, len(sessions))
			}
		}()
	}

	wg.Wait()

	// Concurrent deletes
	wg = sync.WaitGroup{}
	for _, id := range ids {
		wg.Add(1)
		go func(sessionID string) {
			defer wg.Done()
			err := m.DeleteSession(sessionID)
			if err != nil {
				t.Errorf("DeleteSession failed: %v", err)
			}
		}(id)
	}

	wg.Wait()

	// Verify all sessions are deleted
	sessions := m.ListSessions()
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions after deletion, got %d", len(sessions))
	}
}
