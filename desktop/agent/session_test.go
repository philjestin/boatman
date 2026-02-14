package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestNewSession tests session initialization
func TestNewSession(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		projectPath string
	}{
		{
			name:        "basic session creation",
			id:          "test-session-1",
			projectPath: "/path/to/project",
		},
		{
			name:        "session with empty project path",
			id:          "test-session-2",
			projectPath: "",
		},
		{
			name:        "session with long id",
			id:          "test-session-very-long-id-with-many-characters",
			projectPath: "/path/to/another/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession(tt.id, tt.projectPath)

			if session == nil {
				t.Fatal("NewSession returned nil")
			}

			if session.ID != tt.id {
				t.Errorf("Expected ID %s, got %s", tt.id, session.ID)
			}

			if session.ProjectPath != tt.projectPath {
				t.Errorf("Expected ProjectPath %s, got %s", tt.projectPath, session.ProjectPath)
			}

			if session.Status != SessionStatusIdle {
				t.Errorf("Expected Status %s, got %s", SessionStatusIdle, session.Status)
			}

			if session.Messages == nil {
				t.Error("Messages should be initialized")
			}

			if len(session.Messages) != 0 {
				t.Errorf("Expected 0 messages, got %d", len(session.Messages))
			}

			if session.Tasks == nil {
				t.Error("Tasks should be initialized")
			}

			if len(session.Tasks) != 0 {
				t.Errorf("Expected 0 tasks, got %d", len(session.Tasks))
			}

			if session.CreatedAt.IsZero() {
				t.Error("CreatedAt should be set")
			}

			if session.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should be set")
			}

			if session.currentAgentID != "main" {
				t.Errorf("Expected currentAgentID to be 'main', got %s", session.currentAgentID)
			}

			if session.agents == nil {
				t.Error("agents should be initialized")
			}

			if len(session.agents) != 1 {
				t.Errorf("Expected 1 agent, got %d", len(session.agents))
			}

			mainAgent, ok := session.agents["main"]
			if !ok {
				t.Error("Expected main agent to be present")
			}

			if mainAgent.AgentID != "main" {
				t.Errorf("Expected main agent ID to be 'main', got %s", mainAgent.AgentID)
			}

			if mainAgent.AgentType != "main" {
				t.Errorf("Expected main agent type to be 'main', got %s", mainAgent.AgentType)
			}
		})
	}
}

// TestSessionStart tests the Start method
func TestSessionStart(t *testing.T) {
	tests := []struct {
		name  string
		model string
	}{
		{
			name:  "start with sonnet model",
			model: "sonnet",
		},
		{
			name:  "start with opus model",
			model: "opus",
		},
		{
			name:  "start with empty model",
			model: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession("test-session", "/path/to/project")

			err := session.Start(tt.model)
			if err != nil {
				t.Fatalf("Start failed: %v", err)
			}

			if session.Model != tt.model {
				t.Errorf("Expected Model %s, got %s", tt.model, session.Model)
			}

			if session.Status != SessionStatusIdle {
				t.Errorf("Expected Status %s, got %s", SessionStatusIdle, session.Status)
			}

			if session.ctx == nil {
				t.Error("ctx should be set")
			}

			if session.cancel == nil {
				t.Error("cancel should be set")
			}
		})
	}
}

// TestSessionStop tests the Stop method
func TestSessionStop(t *testing.T) {
	t.Run("stop idle session", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.Start("sonnet")

		err := session.Stop()
		if err != nil {
			t.Fatalf("Stop failed: %v", err)
		}

		if session.Status != SessionStatusStopped {
			t.Errorf("Expected Status %s, got %s", SessionStatusStopped, session.Status)
		}
	})

	t.Run("stop without start", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		err := session.Stop()
		if err != nil {
			t.Fatalf("Stop failed: %v", err)
		}

		if session.Status != SessionStatusStopped {
			t.Errorf("Expected Status %s, got %s", SessionStatusStopped, session.Status)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.Start("sonnet")

		originalCtx := session.ctx

		err := session.Stop()
		if err != nil {
			t.Fatalf("Stop failed: %v", err)
		}

		select {
		case <-originalCtx.Done():
			// Context was cancelled, expected
		case <-time.After(100 * time.Millisecond):
			t.Error("Context was not cancelled")
		}
	})
}

// TestSetHandlers tests handler setters
func TestSetHandlers(t *testing.T) {
	t.Run("set message handler", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		called := false
		var receivedMsg Message

		session.SetMessageHandler(func(msg Message) {
			called = true
			receivedMsg = msg
		})

		// Trigger a message
		session.addAssistantMessage("test message")

		if !called {
			t.Error("Message handler was not called")
		}

		if receivedMsg.Content != "test message" {
			t.Errorf("Expected message content 'test message', got %s", receivedMsg.Content)
		}

		if receivedMsg.Role != "assistant" {
			t.Errorf("Expected role 'assistant', got %s", receivedMsg.Role)
		}
	})

	t.Run("set task handler", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		called := false
		var receivedTask Task

		session.SetTaskHandler(func(task Task) {
			called = true
			receivedTask = task
		})

		// Trigger a task event
		event := map[string]any{
			"type": "task_create",
			"task": map[string]any{
				"id":          "task-1",
				"subject":     "Test Task",
				"description": "Test Description",
				"status":      "pending",
			},
		}
		session.handleTaskEvent(event)

		if !called {
			t.Error("Task handler was not called")
		}

		if receivedTask.ID != "task-1" {
			t.Errorf("Expected task ID 'task-1', got %s", receivedTask.ID)
		}

		if receivedTask.Subject != "Test Task" {
			t.Errorf("Expected task subject 'Test Task', got %s", receivedTask.Subject)
		}
	})

	t.Run("set status handler", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		called := false
		var receivedStatus SessionStatus

		session.SetStatusHandler(func(status SessionStatus) {
			called = true
			receivedStatus = status
		})

		session.Start("sonnet")

		if !called {
			t.Error("Status handler was not called")
		}

		if receivedStatus != SessionStatusIdle {
			t.Errorf("Expected status %s, got %s", SessionStatusIdle, receivedStatus)
		}
	})
}

// TestSessionSendMessage tests message sending (without actual process)
func TestSessionSendMessage(t *testing.T) {
	t.Run("send message to idle session", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.Start("sonnet")

		var messageCalled bool
		var statusCalled bool
		var receivedMessage Message
		var receivedStatus SessionStatus

		session.SetMessageHandler(func(msg Message) {
			messageCalled = true
			receivedMessage = msg
		})

		session.SetStatusHandler(func(status SessionStatus) {
			statusCalled = true
			receivedStatus = status
		})

		authConfig := AuthConfig{
			Method:       "anthropic-api",
			APIKey:       "test-key",
			ApprovalMode: "full-auto",
		}

		// Note: This will fail to actually run claude, but should add the user message
		err := session.SendMessage("Hello, Claude!", authConfig)
		if err != nil {
			t.Fatalf("SendMessage failed: %v", err)
		}

		// Give it a moment to process
		time.Sleep(10 * time.Millisecond)

		if !messageCalled {
			t.Error("Message handler was not called")
		}

		if receivedMessage.Role != "user" {
			t.Errorf("Expected role 'user', got %s", receivedMessage.Role)
		}

		if receivedMessage.Content != "Hello, Claude!" {
			t.Errorf("Expected content 'Hello, Claude!', got %s", receivedMessage.Content)
		}

		messages := session.GetMessages()
		if len(messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(messages))
		}

		if !statusCalled {
			t.Error("Status handler was not called")
		}

		if receivedStatus != SessionStatusRunning {
			t.Errorf("Expected status %s, got %s", SessionStatusRunning, receivedStatus)
		}
	})

	t.Run("send message to stopped session", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.Start("sonnet")
		session.Stop()

		authConfig := AuthConfig{
			Method: "anthropic-api",
			APIKey: "test-key",
		}

		err := session.SendMessage("Hello, Claude!", authConfig)
		if err == nil {
			t.Error("Expected error when sending to stopped session")
		}

		if !strings.Contains(err.Error(), "session not available") {
			t.Errorf("Expected 'session not available' error, got: %v", err)
		}
	})

	t.Run("send message to error session", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.Start("sonnet")
		session.mu.Lock()
		session.Status = SessionStatusError
		session.mu.Unlock()

		authConfig := AuthConfig{
			Method: "anthropic-api",
			APIKey: "test-key",
		}

		err := session.SendMessage("Hello, Claude!", authConfig)
		if err == nil {
			t.Error("Expected error when sending to error session")
		}
	})
}

// TestParseStreamLine tests JSON stream parsing
func TestParseStreamLine(t *testing.T) {
	tests := []struct {
		name           string
		line           string
		expectMessages int
		expectStatus   SessionStatus
		checkFunc      func(*testing.T, *Session)
	}{
		{
			name:           "empty line",
			line:           "",
			expectMessages: 0,
		},
		{
			name:           "whitespace only",
			line:           "   ",
			expectMessages: 0,
		},
		{
			name: "system event with conversation ID",
			line: `{"type":"system","conversation_id":"conv-123"}`,
			checkFunc: func(t *testing.T, s *Session) {
				if s.conversationID != "conv-123" {
					t.Errorf("Expected conversationID 'conv-123', got %s", s.conversationID)
				}
			},
		},
		{
			name: "system event with session ID",
			line: `{"type":"system","session_id":"session-456"}`,
			checkFunc: func(t *testing.T, s *Session) {
				if s.conversationID != "session-456" {
					t.Errorf("Expected conversationID 'session-456', got %s", s.conversationID)
				}
			},
		},
		{
			name:           "content_block_delta",
			line:           `{"type":"content_block_delta","delta":{"text":"Hello"}}`,
			expectMessages: 0, // Deltas don't create messages immediately
		},
		{
			name:           "content_block_stop",
			line:           `{"type":"content_block_stop"}`,
			expectMessages: 0, // No message if buffer is empty
		},
		{
			name: "message_start with usage",
			line: `{"type":"message_start","message":{"usage":{"input_tokens":100,"output_tokens":50}}}`,
			checkFunc: func(t *testing.T, s *Session) {
				// Usage is tracked but not necessarily a message
			},
		},
		{
			name: "message_delta with usage",
			line: `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":100,"output_tokens":50}}`,
			checkFunc: func(t *testing.T, s *Session) {
				// Final usage creates a system message
				messages := s.GetMessages()
				found := false
				for _, msg := range messages {
					if msg.Role == "system" && strings.Contains(msg.Content, "Token usage") {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected token usage message")
				}
			},
		},
		{
			name:           "message_stop",
			line:           `{"type":"message_stop"}`,
			expectMessages: 0,
		},
		{
			name: "result with session_id",
			line: `{"type":"result","result":{"session_id":"result-session-789"}}`,
			checkFunc: func(t *testing.T, s *Session) {
				if s.conversationID != "result-session-789" {
					t.Errorf("Expected conversationID 'result-session-789', got %s", s.conversationID)
				}
			},
		},
		{
			name: "tool_use event",
			line: `{"type":"tool_use","name":"Read","id":"tool-1","input":{"file_path":"/test/file.go"}}`,
			checkFunc: func(t *testing.T, s *Session) {
				messages := s.GetMessages()
				if len(messages) == 0 {
					t.Fatal("Expected at least one message")
				}
				lastMsg := messages[len(messages)-1]
				if lastMsg.Metadata == nil || lastMsg.Metadata.ToolUse == nil {
					t.Error("Expected tool use metadata")
				}
				if lastMsg.Metadata.ToolUse.ToolName != "Read" {
					t.Errorf("Expected tool name 'Read', got %s", lastMsg.Metadata.ToolUse.ToolName)
				}
			},
		},
		{
			name: "tool_result event",
			line: `{"type":"tool_result","tool_use_id":"tool-1","content":"file contents","is_error":false}`,
			checkFunc: func(t *testing.T, s *Session) {
				messages := s.GetMessages()
				if len(messages) == 0 {
					t.Fatal("Expected at least one message")
				}
				lastMsg := messages[len(messages)-1]
				if lastMsg.Metadata == nil || lastMsg.Metadata.ToolResult == nil {
					t.Error("Expected tool result metadata")
				}
				if lastMsg.Metadata.ToolResult.ToolID != "tool-1" {
					t.Errorf("Expected tool ID 'tool-1', got %s", lastMsg.Metadata.ToolResult.ToolID)
				}
			},
		},
		{
			name:         "input_request event",
			line:         `{"type":"input_request"}`,
			expectStatus: SessionStatusWaiting,
		},
		{
			name: "task_create event",
			line: `{"type":"task_create","task":{"id":"task-1","subject":"Test Task","description":"Test","status":"pending"}}`,
			checkFunc: func(t *testing.T, s *Session) {
				tasks := s.GetTasks()
				if len(tasks) != 1 {
					t.Fatalf("Expected 1 task, got %d", len(tasks))
				}
				if tasks[0].ID != "task-1" {
					t.Errorf("Expected task ID 'task-1', got %s", tasks[0].ID)
				}
			},
		},
		{
			name:         "error event",
			line:         `{"type":"error","error":{"message":"something went wrong"}}`,
			expectStatus: SessionStatusError,
			checkFunc: func(t *testing.T, s *Session) {
				messages := s.GetMessages()
				found := false
				for _, msg := range messages {
					if strings.Contains(msg.Content, "something went wrong") {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected error message")
				}
			},
		},
		{
			name: "non-JSON line",
			line: "This is not JSON",
			checkFunc: func(t *testing.T, s *Session) {
				messages := s.GetMessages()
				found := false
				for _, msg := range messages {
					if strings.Contains(msg.Content, "This is not JSON") {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected system message for non-JSON line")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession("test-session", "/path/to/project")
			session.Start("sonnet")

			var responseBuilder strings.Builder
			var currentMessageID string
			session.parseStreamLine(tt.line, &responseBuilder, &currentMessageID)

			if tt.expectMessages > 0 {
				messages := session.GetMessages()
				if len(messages) != tt.expectMessages {
					t.Errorf("Expected %d messages, got %d", tt.expectMessages, len(messages))
				}
			}

			if tt.expectStatus != "" {
				if session.Status != tt.expectStatus {
					t.Errorf("Expected status %s, got %s", tt.expectStatus, session.Status)
				}
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, session)
			}
		})
	}
}

// TestFormatToolUseDescription tests tool use formatting
func TestFormatToolUseDescription(t *testing.T) {
	tests := []struct {
		name        string
		toolName    string
		input       any
		expectedMsg string
	}{
		{
			name:        "Read tool",
			toolName:    "Read",
			input:       map[string]any{"file_path": "/test/file.go"},
			expectedMsg: "📖 Reading file: /test/file.go",
		},
		{
			name:        "Write tool",
			toolName:    "Write",
			input:       map[string]any{"file_path": "/test/output.txt"},
			expectedMsg: "✏️  Writing file: /test/output.txt",
		},
		{
			name:        "Edit tool",
			toolName:    "Edit",
			input:       map[string]any{"file_path": "/test/edit.txt"},
			expectedMsg: "✏️  Editing file: /test/edit.txt",
		},
		{
			name:        "Bash tool",
			toolName:    "Bash",
			input:       map[string]any{"command": "ls -la"},
			expectedMsg: "💻 Running: ls -la",
		},
		{
			name:        "Bash tool with long command",
			toolName:    "Bash",
			input:       map[string]any{"command": strings.Repeat("a", 100)},
			expectedMsg: "💻 Running: " + strings.Repeat("a", 60) + "...",
		},
		{
			name:        "Glob tool",
			toolName:    "Glob",
			input:       map[string]any{"pattern": "**/*.go"},
			expectedMsg: "🔍 Searching files: **/*.go",
		},
		{
			name:        "Grep tool",
			toolName:    "Grep",
			input:       map[string]any{"pattern": "TODO"},
			expectedMsg: "🔍 Searching content: TODO",
		},
		{
			name:        "Task tool",
			toolName:    "Task",
			input:       map[string]any{"description": "Analyze codebase"},
			expectedMsg: "🤖 Starting agent: Analyze codebase",
		},
		{
			name:        "WebSearch tool",
			toolName:    "WebSearch",
			input:       map[string]any{"query": "golang testing"},
			expectedMsg: "🌐 Searching web: golang testing",
		},
		{
			name:        "WebFetch tool",
			toolName:    "WebFetch",
			input:       map[string]any{"url": "https://example.com"},
			expectedMsg: "🌐 Fetching: https://example.com",
		},
		{
			name:        "Unknown tool",
			toolName:    "UnknownTool",
			input:       map[string]any{"some": "data"},
			expectedMsg: "🔧 Using tool: UnknownTool",
		},
		{
			name:        "Tool with invalid input",
			toolName:    "Read",
			input:       "not a map",
			expectedMsg: "🔧 Using tool: Read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession("test-session", "/path/to/project")
			result := session.formatToolUseDescription(tt.toolName, tt.input)

			if result != tt.expectedMsg {
				t.Errorf("Expected %s, got %s", tt.expectedMsg, result)
			}
		})
	}
}

// TestHandleToolUse tests tool use handling
func TestHandleToolUse(t *testing.T) {
	t.Run("handle read tool use", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		var receivedMsg Message

		session.SetMessageHandler(func(msg Message) {
			receivedMsg = msg
		})

		event := map[string]any{
			"type": "tool_use",
			"name": "Read",
			"id":   "tool-123",
			"input": map[string]any{
				"file_path": "/test/file.go",
			},
		}

		session.handleToolUse(event)

		if receivedMsg.Role != "assistant" {
			t.Errorf("Expected role 'assistant', got %s", receivedMsg.Role)
		}

		if receivedMsg.Metadata == nil || receivedMsg.Metadata.ToolUse == nil {
			t.Fatal("Expected tool use metadata")
		}

		if receivedMsg.Metadata.ToolUse.ToolName != "Read" {
			t.Errorf("Expected tool name 'Read', got %s", receivedMsg.Metadata.ToolUse.ToolName)
		}

		if receivedMsg.Metadata.ToolUse.ToolID != "tool-123" {
			t.Errorf("Expected tool ID 'tool-123', got %s", receivedMsg.Metadata.ToolUse.ToolID)
		}
	})

	t.Run("handle Task tool use (spawns agent)", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		event := map[string]any{
			"type": "tool_use",
			"name": "Task",
			"id":   "tool-456",
			"input": map[string]any{
				"description":    "Analyze codebase",
				"subagent_type":  "explore",
			},
		}

		initialAgentCount := len(session.agents)
		session.handleToolUse(event)
		newAgentCount := len(session.agents)

		if newAgentCount != initialAgentCount+1 {
			t.Errorf("Expected %d agents, got %d", initialAgentCount+1, newAgentCount)
		}
	})
}

// TestHandleToolResult tests tool result handling
func TestHandleToolResult(t *testing.T) {
	tests := []struct {
		name      string
		event     map[string]any
		checkFunc func(*testing.T, Message)
	}{
		{
			name: "successful tool result with string content",
			event: map[string]any{
				"type":        "tool_result",
				"tool_use_id": "tool-123",
				"content":     "File contents here",
				"is_error":    false,
			},
			checkFunc: func(t *testing.T, msg Message) {
				if msg.Role != "system" {
					t.Errorf("Expected role 'system', got %s", msg.Role)
				}
				if !strings.Contains(msg.Content, "✅") {
					t.Error("Expected success emoji in content")
				}
				if msg.Metadata == nil || msg.Metadata.ToolResult == nil {
					t.Fatal("Expected tool result metadata")
				}
				if msg.Metadata.ToolResult.IsError {
					t.Error("Expected IsError to be false")
				}
			},
		},
		{
			name: "error tool result",
			event: map[string]any{
				"type":        "tool_result",
				"tool_use_id": "tool-456",
				"content":     "File not found",
				"is_error":    true,
			},
			checkFunc: func(t *testing.T, msg Message) {
				if !strings.Contains(msg.Content, "❌") {
					t.Error("Expected error emoji in content")
				}
				if msg.Metadata == nil || msg.Metadata.ToolResult == nil {
					t.Fatal("Expected tool result metadata")
				}
				if !msg.Metadata.ToolResult.IsError {
					t.Error("Expected IsError to be true")
				}
			},
		},
		{
			name: "tool result with array content",
			event: map[string]any{
				"type":        "tool_result",
				"tool_use_id": "tool-789",
				"content": []any{
					map[string]any{"text": "First block"},
					map[string]any{"text": "Second block"},
				},
				"is_error": false,
			},
			checkFunc: func(t *testing.T, msg Message) {
				if msg.Metadata == nil || msg.Metadata.ToolResult == nil {
					t.Fatal("Expected tool result metadata")
				}
				content := msg.Metadata.ToolResult.Content
				if !strings.Contains(content, "First block") || !strings.Contains(content, "Second block") {
					t.Errorf("Expected combined content, got: %s", content)
				}
			},
		},
		{
			name: "tool result with long content (truncation)",
			event: map[string]any{
				"type":        "tool_result",
				"tool_use_id": "tool-long",
				"content":     strings.Repeat("a", 600),
				"is_error":    false,
			},
			checkFunc: func(t *testing.T, msg Message) {
				if !strings.Contains(msg.Content, "(truncated)") {
					t.Error("Expected content to be truncated")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession("test-session", "/path/to/project")
			var receivedMsg Message

			session.SetMessageHandler(func(msg Message) {
				receivedMsg = msg
			})

			session.handleToolResult(tt.event)

			if tt.checkFunc != nil {
				tt.checkFunc(t, receivedMsg)
			}
		})
	}
}

// TestHandleTaskEvent tests task tracking
func TestHandleTaskEvent(t *testing.T) {
	t.Run("create new task", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		var receivedTask Task
		taskCalled := false

		session.SetTaskHandler(func(task Task) {
			taskCalled = true
			receivedTask = task
		})

		event := map[string]any{
			"type": "task_create",
			"task": map[string]any{
				"id":          "task-1",
				"subject":     "Implement feature",
				"description": "Add new functionality",
				"status":      "pending",
			},
		}

		session.handleTaskEvent(event)

		if !taskCalled {
			t.Error("Task handler was not called")
		}

		if receivedTask.ID != "task-1" {
			t.Errorf("Expected task ID 'task-1', got %s", receivedTask.ID)
		}

		if receivedTask.Subject != "Implement feature" {
			t.Errorf("Expected subject 'Implement feature', got %s", receivedTask.Subject)
		}

		if receivedTask.Status != "pending" {
			t.Errorf("Expected status 'pending', got %s", receivedTask.Status)
		}

		tasks := session.GetTasks()
		if len(tasks) != 1 {
			t.Errorf("Expected 1 task, got %d", len(tasks))
		}
	})

	t.Run("update existing task", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		// Create initial task
		event1 := map[string]any{
			"type": "task_create",
			"task": map[string]any{
				"id":          "task-1",
				"subject":     "Implement feature",
				"description": "Add new functionality",
				"status":      "pending",
			},
		}
		session.handleTaskEvent(event1)

		// Update task
		event2 := map[string]any{
			"type": "task_update",
			"task": map[string]any{
				"id":          "task-1",
				"subject":     "Implement feature",
				"description": "Add new functionality",
				"status":      "in_progress",
			},
		}
		session.handleTaskEvent(event2)

		tasks := session.GetTasks()
		if len(tasks) != 1 {
			t.Errorf("Expected 1 task, got %d", len(tasks))
		}

		if tasks[0].Status != "in_progress" {
			t.Errorf("Expected status 'in_progress', got %s", tasks[0].Status)
		}
	})

	t.Run("task without ID gets generated ID", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		event := map[string]any{
			"type": "task_create",
			"task": map[string]any{
				"subject":     "No ID task",
				"description": "Task without ID",
			},
		}

		session.handleTaskEvent(event)

		tasks := session.GetTasks()
		if len(tasks) != 1 {
			t.Fatalf("Expected 1 task, got %d", len(tasks))
		}

		if tasks[0].ID == "" {
			t.Error("Expected generated task ID")
		}

		if !strings.HasPrefix(tasks[0].ID, "task-") {
			t.Errorf("Expected task ID to start with 'task-', got %s", tasks[0].ID)
		}
	})

	t.Run("task without status defaults to pending", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		event := map[string]any{
			"type": "task_create",
			"task": map[string]any{
				"id":          "task-2",
				"subject":     "No status task",
				"description": "Task without status",
			},
		}

		session.handleTaskEvent(event)

		tasks := session.GetTasks()
		if len(tasks) != 1 {
			t.Fatalf("Expected 1 task, got %d", len(tasks))
		}

		if tasks[0].Status != "pending" {
			t.Errorf("Expected status 'pending', got %s", tasks[0].Status)
		}
	})
}

// TestHandleUsageInfo tests token usage tracking
func TestHandleUsageInfo(t *testing.T) {
	t.Run("handle initial usage info", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		usage := map[string]any{
			"input_tokens":  float64(1000),
			"output_tokens": float64(500),
		}

		session.handleUsageInfo(usage, false)

		// Non-final usage should not create a message
		messages := session.GetMessages()
		if len(messages) != 0 {
			t.Errorf("Expected 0 messages for non-final usage, got %d", len(messages))
		}
	})

	t.Run("handle final usage info", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		usage := map[string]any{
			"input_tokens":  float64(1000),
			"output_tokens": float64(500),
		}

		session.handleUsageInfo(usage, true)

		messages := session.GetMessages()
		if len(messages) != 1 {
			t.Fatalf("Expected 1 message for final usage, got %d", len(messages))
		}

		msg := messages[0]
		if msg.Role != "system" {
			t.Errorf("Expected role 'system', got %s", msg.Role)
		}

		if !strings.Contains(msg.Content, "Token usage") {
			t.Error("Expected message to contain 'Token usage'")
		}

		if !strings.Contains(msg.Content, "1000 input") {
			t.Error("Expected message to contain input token count")
		}

		if !strings.Contains(msg.Content, "500 output") {
			t.Error("Expected message to contain output token count")
		}

		if msg.Metadata == nil || msg.Metadata.CostInfo == nil {
			t.Fatal("Expected cost info metadata")
		}

		costInfo := msg.Metadata.CostInfo
		if costInfo.InputTokens != 1000 {
			t.Errorf("Expected 1000 input tokens, got %d", costInfo.InputTokens)
		}

		if costInfo.OutputTokens != 500 {
			t.Errorf("Expected 500 output tokens, got %d", costInfo.OutputTokens)
		}

		// Check cost calculation (3/1M for input, 15/1M for output)
		expectedCost := (1000.0 * 3.0 / 1_000_000) + (500.0 * 15.0 / 1_000_000)
		if costInfo.TotalCost != expectedCost {
			t.Errorf("Expected cost %.6f, got %.6f", expectedCost, costInfo.TotalCost)
		}
	})

	t.Run("handle zero usage", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		usage := map[string]any{
			"input_tokens":  float64(0),
			"output_tokens": float64(0),
		}

		session.handleUsageInfo(usage, true)

		messages := session.GetMessages()
		if len(messages) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(messages))
		}

		msg := messages[0]
		if msg.Metadata.CostInfo.TotalCost != 0 {
			t.Errorf("Expected cost 0, got %.6f", msg.Metadata.CostInfo.TotalCost)
		}
	})
}

// TestGetMessagesAndTasks tests getter methods
func TestGetMessagesAndTasks(t *testing.T) {
	t.Run("get messages returns copy", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.addAssistantMessage("Test message 1")
		session.addAssistantMessage("Test message 2")

		messages1 := session.GetMessages()
		messages2 := session.GetMessages()

		if len(messages1) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(messages1))
		}

		// Modify the returned slice
		messages1[0].Content = "Modified"

		// Original should be unchanged
		messages3 := session.GetMessages()
		if messages3[0].Content == "Modified" {
			t.Error("GetMessages should return a copy, not original")
		}

		// Different calls should return different slices
		if &messages1[0] == &messages2[0] {
			t.Error("GetMessages should return different slice instances")
		}
	})

	t.Run("get tasks returns copy", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		event1 := map[string]any{
			"type": "task_create",
			"task": map[string]any{
				"id":      "task-1",
				"subject": "Task 1",
			},
		}
		event2 := map[string]any{
			"type": "task_create",
			"task": map[string]any{
				"id":      "task-2",
				"subject": "Task 2",
			},
		}

		session.handleTaskEvent(event1)
		session.handleTaskEvent(event2)

		tasks1 := session.GetTasks()
		tasks2 := session.GetTasks()

		if len(tasks1) != 2 {
			t.Errorf("Expected 2 tasks, got %d", len(tasks1))
		}

		// Modify the returned slice
		tasks1[0].Subject = "Modified"

		// Original should be unchanged
		tasks3 := session.GetTasks()
		if tasks3[0].Subject == "Modified" {
			t.Error("GetTasks should return a copy, not original")
		}

		// Different calls should return different slices
		if &tasks1[0] == &tasks2[0] {
			t.Error("GetTasks should return different slice instances")
		}
	})
}

// TestMessageMetadata tests message metadata handling
func TestMessageMetadata(t *testing.T) {
	t.Run("assistant message has agent metadata", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.addAssistantMessage("Test message")

		messages := session.GetMessages()
		if len(messages) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(messages))
		}

		msg := messages[0]
		if msg.Metadata == nil {
			t.Fatal("Expected metadata")
		}

		if msg.Metadata.Agent == nil {
			t.Fatal("Expected agent metadata")
		}

		if msg.Metadata.Agent.AgentID != "main" {
			t.Errorf("Expected agent ID 'main', got %s", msg.Metadata.Agent.AgentID)
		}

		if msg.Metadata.Agent.AgentType != "main" {
			t.Errorf("Expected agent type 'main', got %s", msg.Metadata.Agent.AgentType)
		}
	})

	t.Run("system message has agent metadata", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.addSystemMessage("System message")

		messages := session.GetMessages()
		if len(messages) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(messages))
		}

		msg := messages[0]
		if msg.Metadata == nil || msg.Metadata.Agent == nil {
			t.Fatal("Expected agent metadata")
		}
	})
}

// TestConcurrency tests thread safety
func TestConcurrency(t *testing.T) {
	t.Run("concurrent message additions", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.Start("sonnet")

		var wg sync.WaitGroup
		numGoroutines := 10
		messagesPerGoroutine := 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < messagesPerGoroutine; j++ {
					session.addAssistantMessage(fmt.Sprintf("Message %d-%d", id, j))
				}
			}(i)
		}

		wg.Wait()

		messages := session.GetMessages()
		expectedCount := numGoroutines * messagesPerGoroutine
		if len(messages) != expectedCount {
			t.Errorf("Expected %d messages, got %d", expectedCount, len(messages))
		}
	})

	t.Run("concurrent handler sets and message additions", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		var wg sync.WaitGroup
		messageCount := 0
		var countMu sync.Mutex

		// Set handlers concurrently
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				session.SetMessageHandler(func(msg Message) {
					countMu.Lock()
					messageCount++
					countMu.Unlock()
				})
			}()
		}

		// Add messages concurrently
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				session.addAssistantMessage(fmt.Sprintf("Message %d", id))
			}(i)
		}

		wg.Wait()

		// Should have received all messages (handler might be set multiple times)
		if messageCount != 5 {
			t.Errorf("Expected 5 messages to be handled, got %d", messageCount)
		}
	})

	t.Run("concurrent task operations", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		var wg sync.WaitGroup
		numGoroutines := 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				event := map[string]any{
					"type": "task_create",
					"task": map[string]any{
						"id":      fmt.Sprintf("task-%d", id),
						"subject": fmt.Sprintf("Task %d", id),
						"status":  "pending",
					},
				}
				session.handleTaskEvent(event)
			}(i)
		}

		wg.Wait()

		tasks := session.GetTasks()
		if len(tasks) != numGoroutines {
			t.Errorf("Expected %d tasks, got %d", numGoroutines, len(tasks))
		}
	})

	t.Run("concurrent start and stop", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		var wg sync.WaitGroup

		// Start multiple times concurrently
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				session.Start("sonnet")
			}()
		}

		wg.Wait()

		// Stop multiple times concurrently
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				session.Stop()
			}()
		}

		wg.Wait()

		if session.Status != SessionStatusStopped {
			t.Errorf("Expected status %s, got %s", SessionStatusStopped, session.Status)
		}
	})
}

// TestEdgeCases tests edge cases and error handling
func TestEdgeCases(t *testing.T) {
	t.Run("add empty assistant message", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.addAssistantMessage("")

		messages := session.GetMessages()
		if len(messages) != 0 {
			t.Errorf("Expected 0 messages for empty content, got %d", len(messages))
		}
	})

	t.Run("add whitespace-only assistant message", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.addAssistantMessage("   \n\t  ")

		messages := session.GetMessages()
		if len(messages) != 0 {
			t.Errorf("Expected 0 messages for whitespace-only content, got %d", len(messages))
		}
	})

	t.Run("parse malformed JSON", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		var responseBuilder strings.Builder
		var currentMessageID string

		session.parseStreamLine(`{"type":"incomplete"`, &responseBuilder, &currentMessageID)

		// Should not crash, might add a system message
	})

	t.Run("handle task event with missing task data", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		event := map[string]any{
			"type": "task_create",
			// No task field
		}

		// Should not crash
		session.handleTaskEvent(event)

		tasks := session.GetTasks()
		if len(tasks) != 0 {
			t.Errorf("Expected 0 tasks, got %d", len(tasks))
		}
	})

	t.Run("handle tool use with missing fields", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		event := map[string]any{
			"type": "tool_use",
			// Missing name, id, input
		}

		// Should not crash
		session.handleToolUse(event)

		messages := session.GetMessages()
		if len(messages) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(messages))
		}

		if messages[0].Metadata == nil || messages[0].Metadata.ToolUse == nil {
			t.Fatal("Expected tool use metadata")
		}

		if messages[0].Metadata.ToolUse.ToolName != "" {
			t.Error("Expected empty tool name")
		}
	})

	t.Run("handle tool result with missing fields", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		event := map[string]any{
			"type": "tool_result",
			// Missing tool_use_id, content
		}

		// Should not crash
		session.handleToolResult(event)

		messages := session.GetMessages()
		if len(messages) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(messages))
		}
	})

	t.Run("handle usage with invalid types", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		usage := map[string]any{
			"input_tokens":  "not a number",
			"output_tokens": "also not a number",
		}

		// Should not crash, should default to 0
		session.handleUsageInfo(usage, true)

		messages := session.GetMessages()
		if len(messages) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(messages))
		}

		if messages[0].Metadata == nil || messages[0].Metadata.CostInfo == nil {
			t.Fatal("Expected cost info")
		}

		if messages[0].Metadata.CostInfo.InputTokens != 0 {
			t.Errorf("Expected 0 input tokens, got %d", messages[0].Metadata.CostInfo.InputTokens)
		}
	})

	t.Run("approve and reject not supported", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		err := session.Approve("action-1")
		if err == nil {
			t.Error("Expected error for Approve")
		}

		err = session.Reject("action-1")
		if err == nil {
			t.Error("Expected error for Reject")
		}
	})
}

// TestHandleError tests error handling
func TestHandleError(t *testing.T) {
	t.Run("handle error sets status and creates message", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.Start("sonnet")

		var statusCalled bool
		var receivedStatus SessionStatus

		session.SetStatusHandler(func(status SessionStatus) {
			statusCalled = true
			receivedStatus = status
		})

		err := fmt.Errorf("test error occurred")
		session.handleError(err)

		if !statusCalled {
			t.Error("Status handler was not called")
		}

		if receivedStatus != SessionStatusError {
			t.Errorf("Expected status %s, got %s", SessionStatusError, receivedStatus)
		}

		messages := session.GetMessages()
		found := false
		for _, msg := range messages {
			if strings.Contains(msg.Content, "test error occurred") {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected error message to be added")
		}
	})
}

// TestStreamJSONSequences tests complete streaming sequences
func TestStreamJSONSequences(t *testing.T) {
	t.Run("complete message sequence", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.Start("sonnet")

		var responseBuilder strings.Builder

		// Simulate a complete message flow
		lines := []string{
			`{"type":"system","conversation_id":"conv-abc"}`,
			`{"type":"message_start","message":{"usage":{"input_tokens":100,"output_tokens":0}}}`,
			`{"type":"content_block_delta","delta":{"text":"Hello, "}}`,
			`{"type":"content_block_delta","delta":{"text":"world!"}}`,
			`{"type":"content_block_stop"}`,
			`{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":100,"output_tokens":50}}`,
			`{"type":"message_stop"}`,
		}

		var currentMessageID string
		for _, line := range lines {
			session.parseStreamLine(line, &responseBuilder, &currentMessageID)
		}

		// Check conversation ID was set
		if session.conversationID != "conv-abc" {
			t.Errorf("Expected conversationID 'conv-abc', got %s", session.conversationID)
		}

		// Check messages were created
		messages := session.GetMessages()

		// Should have assistant message and usage message
		assistantFound := false
		usageFound := false

		for _, msg := range messages {
			if msg.Role == "assistant" && msg.Content == "Hello, world!" {
				assistantFound = true
			}
			if msg.Role == "system" && strings.Contains(msg.Content, "Token usage") {
				usageFound = true
			}
		}

		if !assistantFound {
			t.Error("Expected assistant message with 'Hello, world!'")
		}

		if !usageFound {
			t.Error("Expected usage message")
		}
	})

	t.Run("tool use sequence", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.Start("sonnet")

		var responseBuilder strings.Builder

		lines := []string{
			`{"type":"content_block_delta","delta":{"text":"Let me read that file."}}`,
			`{"type":"content_block_stop"}`,
			`{"type":"tool_use","name":"Read","id":"tool-1","input":{"file_path":"/test/file.go"}}`,
			`{"type":"tool_result","tool_use_id":"tool-1","content":"package main","is_error":false}`,
		}

		var currentMessageID string
		for _, line := range lines {
			session.parseStreamLine(line, &responseBuilder, &currentMessageID)
		}

		messages := session.GetMessages()

		// Should have text message, tool use message, and tool result message
		if len(messages) < 3 {
			t.Fatalf("Expected at least 3 messages, got %d", len(messages))
		}

		// Find tool use message
		toolUseFound := false
		toolResultFound := false

		for _, msg := range messages {
			if msg.Metadata != nil {
				if msg.Metadata.ToolUse != nil && msg.Metadata.ToolUse.ToolName == "Read" {
					toolUseFound = true
				}
				if msg.Metadata.ToolResult != nil && msg.Metadata.ToolResult.ToolID == "tool-1" {
					toolResultFound = true
				}
			}
		}

		if !toolUseFound {
			t.Error("Expected tool use message")
		}

		if !toolResultFound {
			t.Error("Expected tool result message")
		}
	})
}

// TestAgentTracking tests sub-agent tracking
func TestAgentTracking(t *testing.T) {
	t.Run("task tool spawns new agent", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		input := map[string]any{
			"description":   "Analyze code structure",
			"subagent_type": "explore",
		}

		initialCount := len(session.agents)
		session.handleTaskSpawn(input)
		newCount := len(session.agents)

		if newCount != initialCount+1 {
			t.Errorf("Expected %d agents, got %d", initialCount+1, newCount)
		}

		// Find the new agent
		var newAgent *AgentInfo
		for id, agent := range session.agents {
			if id != "main" {
				newAgent = agent
				break
			}
		}

		if newAgent == nil {
			t.Fatal("Expected to find new agent")
		}

		if newAgent.AgentType != "explore" {
			t.Errorf("Expected agent type 'explore', got %s", newAgent.AgentType)
		}

		if newAgent.ParentAgentID != "main" {
			t.Errorf("Expected parent agent ID 'main', got %s", newAgent.ParentAgentID)
		}

		if newAgent.Description != "Analyze code structure" {
			t.Errorf("Expected description 'Analyze code structure', got %s", newAgent.Description)
		}
	})

	t.Run("task spawn with invalid input", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")

		// Invalid input (not a map)
		session.handleTaskSpawn("not a map")

		// Should not crash, should not add agent
		if len(session.agents) != 1 {
			t.Errorf("Expected 1 agent, got %d", len(session.agents))
		}
	})
}

// TestJSONMarshaling tests that structs can be marshaled to JSON
func TestJSONMarshaling(t *testing.T) {
	t.Run("marshal message", func(t *testing.T) {
		msg := Message{
			ID:        "msg-1",
			Role:      "assistant",
			Content:   "Test content",
			Timestamp: time.Now(),
			Metadata: &MessageMetadata{
				ToolUse: &ToolUse{
					ToolName: "Read",
					ToolID:   "tool-1",
					Input:    json.RawMessage(`{"file_path":"/test"}`),
				},
			},
		}

		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Failed to marshal message: %v", err)
		}

		var decoded Message
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}

		if decoded.ID != msg.ID {
			t.Errorf("Expected ID %s, got %s", msg.ID, decoded.ID)
		}
	})

	t.Run("marshal task", func(t *testing.T) {
		task := Task{
			ID:          "task-1",
			Subject:     "Test Task",
			Description: "Test Description",
			Status:      "pending",
		}

		data, err := json.Marshal(task)
		if err != nil {
			t.Fatalf("Failed to marshal task: %v", err)
		}

		var decoded Task
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Fatalf("Failed to unmarshal task: %v", err)
		}

		if decoded.ID != task.ID {
			t.Errorf("Expected ID %s, got %s", task.ID, decoded.ID)
		}
	})

	t.Run("marshal session", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.Start("sonnet")

		data, err := json.Marshal(session)
		if err != nil {
			t.Fatalf("Failed to marshal session: %v", err)
		}

		var decoded Session
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Fatalf("Failed to unmarshal session: %v", err)
		}

		if decoded.ID != session.ID {
			t.Errorf("Expected ID %s, got %s", session.ID, decoded.ID)
		}
	})
}

// TestContextCancellation tests context cancellation behavior
func TestContextCancellation(t *testing.T) {
	t.Run("context is cancelled on stop", func(t *testing.T) {
		session := NewSession("test-session", "/path/to/project")
		session.Start("sonnet")

		ctx := session.ctx
		if ctx == nil {
			t.Fatal("Context should be set after Start")
		}

		session.Stop()

		select {
		case <-ctx.Done():
			// Expected - context was cancelled
			if ctx.Err() != context.Canceled {
				t.Errorf("Expected context.Canceled error, got %v", ctx.Err())
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Context was not cancelled within timeout")
		}
	})
}
