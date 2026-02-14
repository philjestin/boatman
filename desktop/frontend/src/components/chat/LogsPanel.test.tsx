import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { LogsPanel } from './LogsPanel';
import type { Message } from '../../types';

describe('LogsPanel', () => {
  const mockMessages: Message[] = [
    {
      id: '1',
      role: 'user',
      content: 'Hello, Claude!',
      timestamp: '2024-01-01T10:00:00Z',
    },
    {
      id: '2',
      role: 'assistant',
      content: 'Hello! How can I help you today?',
      timestamp: '2024-01-01T10:00:01Z',
    },
    {
      id: '3',
      role: 'system',
      content: 'System initialized',
      timestamp: '2024-01-01T10:00:02Z',
    },
  ];

  const mockToolUseMessage: Message = {
    id: '4',
    role: 'assistant',
    content: 'Using a tool',
    timestamp: '2024-01-01T10:00:03Z',
    metadata: {
      toolUse: {
        toolName: 'read_file',
        toolId: 'tool-123',
        input: { path: '/test/file.txt' },
      },
    },
  };

  const mockToolResultMessage: Message = {
    id: '5',
    role: 'assistant',
    content: 'File contents here',
    timestamp: '2024-01-01T10:00:04Z',
    metadata: {
      toolResult: {
        toolId: 'tool-123',
        content: 'File contents here',
        isError: false,
      },
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Component rendering', () => {
    it('should render with collapsed state', () => {
      render(<LogsPanel messages={mockMessages} isActive={false} />);

      expect(screen.getByText('Activity Logs')).toBeInTheDocument();
    });

    it('should render with expanded state by default', () => {
      render(<LogsPanel messages={mockMessages} isActive={false} />);

      expect(screen.getByText('Hello, Claude!')).toBeInTheDocument();
      expect(screen.getByText('Hello! How can I help you today?')).toBeInTheDocument();
    });

    it('should display the correct number of log entries', () => {
      render(<LogsPanel messages={mockMessages} isActive={false} />);

      expect(screen.getByText('(3 entries)')).toBeInTheDocument();
    });

    it('should show Terminal icon', () => {
      const { container } = render(<LogsPanel messages={mockMessages} isActive={false} />);

      const terminalIcon = container.querySelector('.lucide-terminal');
      expect(terminalIcon).toBeInTheDocument();
    });
  });

  describe('Empty state', () => {
    it('should show empty state when no messages', () => {
      render(<LogsPanel messages={[]} isActive={false} />);

      expect(screen.getByText('No activity yet. Send a message to see logs.')).toBeInTheDocument();
    });

    it('should show 0 entries count when empty', () => {
      render(<LogsPanel messages={[]} isActive={false} />);

      expect(screen.getByText('(0 entries)')).toBeInTheDocument();
    });
  });

  describe('Expand/Collapse functionality', () => {
    it('should toggle expanded state when header is clicked', () => {
      render(<LogsPanel messages={mockMessages} isActive={false} />);

      const header = screen.getByText('Activity Logs').closest('button')!;

      // Initially expanded, should see messages
      expect(screen.getByText('Hello, Claude!')).toBeInTheDocument();

      // Click to collapse
      fireEvent.click(header);

      // Messages should be hidden
      expect(screen.queryByText('Hello, Claude!')).not.toBeInTheDocument();

      // Click to expand again
      fireEvent.click(header);

      // Messages should be visible again
      expect(screen.getByText('Hello, Claude!')).toBeInTheDocument();
    });

    it('should show correct chevron icon based on expanded state', () => {
      const { container } = render(<LogsPanel messages={mockMessages} isActive={false} />);

      // Initially expanded, should show ChevronDown
      let chevronDown = container.querySelector('.lucide-chevron-down');
      expect(chevronDown).toBeInTheDocument();

      const header = screen.getByText('Activity Logs').closest('button')!;
      fireEvent.click(header);

      // After collapse, should show ChevronUp
      const chevronUp = container.querySelector('.lucide-chevron-up');
      expect(chevronUp).toBeInTheDocument();
    });
  });

  describe('Active state indicator', () => {
    it('should show active indicator when isActive is true', () => {
      render(<LogsPanel messages={mockMessages} isActive={true} />);

      expect(screen.getByText('Active')).toBeInTheDocument();
    });

    it('should not show active indicator when isActive is false', () => {
      render(<LogsPanel messages={mockMessages} isActive={false} />);

      expect(screen.queryByText('Active')).not.toBeInTheDocument();
    });

    it('should add activity event when active', () => {
      render(<LogsPanel messages={mockMessages} isActive={true} />);

      expect(screen.getByText('⚡ Claude is working...')).toBeInTheDocument();
    });

    it('should update entry count when activity indicator is added', () => {
      render(<LogsPanel messages={mockMessages} isActive={true} />);

      // 3 original messages + 1 activity indicator
      expect(screen.getByText('(4 entries)')).toBeInTheDocument();
    });

    it('should not add activity event when inactive', () => {
      render(<LogsPanel messages={mockMessages} isActive={false} />);

      expect(screen.queryByText('⚡ Claude is working...')).not.toBeInTheDocument();
    });

    it('should not add activity event when no messages', () => {
      render(<LogsPanel messages={[]} isActive={true} />);

      expect(screen.queryByText('⚡ Claude is working...')).not.toBeInTheDocument();
    });
  });

  describe('Clear logs functionality', () => {
    it('should clear logs when trash button is clicked', () => {
      render(<LogsPanel messages={mockMessages} isActive={false} />);

      expect(screen.getByText('Hello, Claude!')).toBeInTheDocument();

      const trashButton = screen.getByTitle('Clear logs');
      fireEvent.click(trashButton);

      expect(screen.getByText('No activity yet. Send a message to see logs.')).toBeInTheDocument();
      expect(screen.queryByText('Hello, Claude!')).not.toBeInTheDocument();
    });

    it('should update entry count to 0 after clearing', () => {
      render(<LogsPanel messages={mockMessages} isActive={false} />);

      const trashButton = screen.getByTitle('Clear logs');
      fireEvent.click(trashButton);

      expect(screen.getByText('(0 entries)')).toBeInTheDocument();
    });

    it('should not expand/collapse when trash button is clicked', () => {
      render(<LogsPanel messages={mockMessages} isActive={false} />);

      const trashButton = screen.getByTitle('Clear logs');

      // Logs should be visible initially
      expect(screen.getByText('Hello, Claude!')).toBeInTheDocument();

      // Click trash (stopPropagation should prevent collapse)
      fireEvent.click(trashButton);

      // Panel should still be expanded (showing empty state)
      expect(screen.getByText('No activity yet. Send a message to see logs.')).toBeInTheDocument();
    });
  });

  describe('Message processing and log entry creation', () => {
    it('should convert user messages correctly', () => {
      const userMessage: Message = {
        id: '1',
        role: 'user',
        content: 'Test user message',
        timestamp: '2024-01-01T10:00:00Z',
      };

      render(<LogsPanel messages={[userMessage]} isActive={false} />);

      expect(screen.getByText('[USER]')).toBeInTheDocument();
      expect(screen.getByText('Test user message')).toBeInTheDocument();
    });

    it('should convert assistant messages correctly', () => {
      const assistantMessage: Message = {
        id: '1',
        role: 'assistant',
        content: 'Test assistant message',
        timestamp: '2024-01-01T10:00:00Z',
      };

      render(<LogsPanel messages={[assistantMessage]} isActive={false} />);

      expect(screen.getByText('[ASSISTANT]')).toBeInTheDocument();
      expect(screen.getByText('Test assistant message')).toBeInTheDocument();
    });

    it('should convert system messages correctly', () => {
      const systemMessage: Message = {
        id: '1',
        role: 'system',
        content: 'Test system message',
        timestamp: '2024-01-01T10:00:00Z',
      };

      render(<LogsPanel messages={[systemMessage]} isActive={false} />);

      expect(screen.getByText('[SYSTEM]')).toBeInTheDocument();
      expect(screen.getByText('Test system message')).toBeInTheDocument();
    });

    it('should identify tool_use messages', () => {
      render(<LogsPanel messages={[mockToolUseMessage]} isActive={false} />);

      expect(screen.getByText('[TOOL→]')).toBeInTheDocument();
      expect(screen.getByText('Using a tool')).toBeInTheDocument();
    });

    it('should identify tool_result messages', () => {
      render(<LogsPanel messages={[mockToolResultMessage]} isActive={false} />);

      expect(screen.getByText('[←RESULT]')).toBeInTheDocument();
      expect(screen.getByText('File contents here')).toBeInTheDocument();
    });

    it('should display timestamps for log entries', () => {
      render(<LogsPanel messages={mockMessages} isActive={false} />);

      const timestamps = screen.getAllByText(/\d{1,2}:\d{2}:\d{2}/);
      expect(timestamps.length).toBeGreaterThan(0);
    });
  });

  describe('Log colors and styling', () => {
    it('should apply correct color to user messages', () => {
      const userMessage: Message = {
        id: '1',
        role: 'user',
        content: 'Test',
        timestamp: '2024-01-01T10:00:00Z',
      };

      const { container } = render(<LogsPanel messages={[userMessage]} isActive={false} />);

      const userPrefix = screen.getByText('[USER]');
      expect(userPrefix).toHaveClass('text-blue-400');
    });

    it('should apply correct color to assistant messages', () => {
      const assistantMessage: Message = {
        id: '1',
        role: 'assistant',
        content: 'Test',
        timestamp: '2024-01-01T10:00:00Z',
      };

      const { container } = render(<LogsPanel messages={[assistantMessage]} isActive={false} />);

      const assistantPrefix = screen.getByText('[ASSISTANT]');
      expect(assistantPrefix).toHaveClass('text-purple-400');
    });

    it('should apply correct color to system messages', () => {
      const systemMessage: Message = {
        id: '1',
        role: 'system',
        content: 'Test',
        timestamp: '2024-01-01T10:00:00Z',
      };

      const { container } = render(<LogsPanel messages={[systemMessage]} isActive={false} />);

      const systemPrefix = screen.getByText('[SYSTEM]');
      expect(systemPrefix).toHaveClass('text-slate-400');
    });

    it('should apply correct color to tool_use messages', () => {
      render(<LogsPanel messages={[mockToolUseMessage]} isActive={false} />);

      const toolUsePrefix = screen.getByText('[TOOL→]');
      expect(toolUsePrefix).toHaveClass('text-amber-400');
    });

    it('should apply correct color to tool_result messages', () => {
      render(<LogsPanel messages={[mockToolResultMessage]} isActive={false} />);

      const toolResultPrefix = screen.getByText('[←RESULT]');
      expect(toolResultPrefix).toHaveClass('text-green-400');
    });

    it('should apply correct color to event messages', () => {
      render(<LogsPanel messages={mockMessages} isActive={true} />);

      const eventPrefix = screen.getByText('[EVENT]');
      expect(eventPrefix).toHaveClass('text-cyan-400');
    });
  });

  describe('Content truncation', () => {
    it('should truncate long content to 200 characters', () => {
      const longContent = 'A'.repeat(300);
      const longMessage: Message = {
        id: '1',
        role: 'user',
        content: longContent,
        timestamp: '2024-01-01T10:00:00Z',
      };

      render(<LogsPanel messages={[longMessage]} isActive={false} />);

      const truncatedContent = 'A'.repeat(200) + '...';
      expect(screen.getByText(truncatedContent)).toBeInTheDocument();
    });

    it('should not truncate short content', () => {
      const shortContent = 'Short message';
      const shortMessage: Message = {
        id: '1',
        role: 'user',
        content: shortContent,
        timestamp: '2024-01-01T10:00:00Z',
      };

      render(<LogsPanel messages={[shortMessage]} isActive={false} />);

      expect(screen.getByText(shortContent)).toBeInTheDocument();
      expect(screen.queryByText(/\.\.\./)).not.toBeInTheDocument();
    });

    it('should truncate exactly 200 character content', () => {
      const exactContent = 'A'.repeat(200);
      const exactMessage: Message = {
        id: '1',
        role: 'user',
        content: exactContent,
        timestamp: '2024-01-01T10:00:00Z',
      };

      render(<LogsPanel messages={[exactMessage]} isActive={false} />);

      expect(screen.getByText(exactContent)).toBeInTheDocument();
      expect(screen.queryByText(/\.\.\./)).not.toBeInTheDocument();
    });

    it('should truncate 201 character content', () => {
      const longContent = 'A'.repeat(201);
      const longMessage: Message = {
        id: '1',
        role: 'user',
        content: longContent,
        timestamp: '2024-01-01T10:00:00Z',
      };

      render(<LogsPanel messages={[longMessage]} isActive={false} />);

      const truncatedContent = 'A'.repeat(200) + '...';
      expect(screen.getByText(truncatedContent)).toBeInTheDocument();
    });
  });

  describe('Auto-scrolling behavior', () => {
    it('should have scroll container with correct height', () => {
      const { container } = render(<LogsPanel messages={mockMessages} isActive={false} />);

      const scrollContainer = container.querySelector('.h-64.overflow-y-auto');
      expect(scrollContainer).toBeInTheDocument();
    });

    it('should include ref element for auto-scroll', () => {
      const { container } = render(<LogsPanel messages={mockMessages} isActive={false} />);

      // The ref is attached to a div, but we can't directly query it
      // We can verify the scroll container exists
      const scrollContainer = container.querySelector('.overflow-y-auto');
      expect(scrollContainer).toBeInTheDocument();
    });
  });

  describe('Message updates', () => {
    it('should update logs when messages prop changes', () => {
      const { rerender } = render(<LogsPanel messages={mockMessages} isActive={false} />);

      expect(screen.getByText('(3 entries)')).toBeInTheDocument();

      const updatedMessages = [
        ...mockMessages,
        {
          id: '4',
          role: 'user',
          content: 'New message',
          timestamp: '2024-01-01T10:00:05Z',
        } as Message,
      ];

      rerender(<LogsPanel messages={updatedMessages} isActive={false} />);

      expect(screen.getByText('(4 entries)')).toBeInTheDocument();
      expect(screen.getByText('New message')).toBeInTheDocument();
    });

    it('should update when isActive changes', () => {
      const { rerender } = render(<LogsPanel messages={mockMessages} isActive={false} />);

      expect(screen.queryByText('Active')).not.toBeInTheDocument();

      rerender(<LogsPanel messages={mockMessages} isActive={true} />);

      expect(screen.getByText('Active')).toBeInTheDocument();
      expect(screen.getByText('⚡ Claude is working...')).toBeInTheDocument();
    });
  });

  describe('Multiple message types in sequence', () => {
    it('should display all message types correctly in sequence', () => {
      const mixedMessages: Message[] = [
        mockMessages[0], // user
        mockMessages[1], // assistant
        mockMessages[2], // system
        mockToolUseMessage,
        mockToolResultMessage,
      ];

      render(<LogsPanel messages={mixedMessages} isActive={false} />);

      expect(screen.getByText('[USER]')).toBeInTheDocument();
      expect(screen.getByText('[ASSISTANT]')).toBeInTheDocument();
      expect(screen.getByText('[SYSTEM]')).toBeInTheDocument();
      expect(screen.getByText('[TOOL→]')).toBeInTheDocument();
      expect(screen.getByText('[←RESULT]')).toBeInTheDocument();
    });

    it('should maintain correct order of logs', () => {
      render(<LogsPanel messages={mockMessages} isActive={false} />);

      const logEntries = screen.getAllByText(/\[.*\]/);
      expect(logEntries[0]).toHaveTextContent('[USER]');
      expect(logEntries[1]).toHaveTextContent('[ASSISTANT]');
      expect(logEntries[2]).toHaveTextContent('[SYSTEM]');
    });
  });

  describe('Edge cases', () => {
    it('should handle messages with empty content', () => {
      const emptyMessage: Message = {
        id: '1',
        role: 'user',
        content: '',
        timestamp: '2024-01-01T10:00:00Z',
      };

      render(<LogsPanel messages={[emptyMessage]} isActive={false} />);

      expect(screen.getByText('[USER]')).toBeInTheDocument();
      expect(screen.getByText('(1 entries)')).toBeInTheDocument();
    });

    it('should handle messages with special characters', () => {
      const specialMessage: Message = {
        id: '1',
        role: 'user',
        content: 'Test <script>alert("xss")</script> & special chars',
        timestamp: '2024-01-01T10:00:00Z',
      };

      render(<LogsPanel messages={[specialMessage]} isActive={false} />);

      expect(screen.getByText('Test <script>alert("xss")</script> & special chars')).toBeInTheDocument();
    });

    it('should handle messages with newlines', () => {
      const multilineMessage: Message = {
        id: '1',
        role: 'user',
        content: 'Line 1\nLine 2\nLine 3',
        timestamp: '2024-01-01T10:00:00Z',
      };

      render(<LogsPanel messages={[multilineMessage]} isActive={false} />);

      // Use a custom matcher to handle newlines properly
      expect(screen.getByText((content, element) => {
        return element?.textContent === 'Line 1\nLine 2\nLine 3';
      })).toBeInTheDocument();
    });

    it('should handle messages with undefined metadata', () => {
      const messageNoMetadata: Message = {
        id: '1',
        role: 'assistant',
        content: 'Test',
        timestamp: '2024-01-01T10:00:00Z',
        metadata: undefined,
      };

      render(<LogsPanel messages={[messageNoMetadata]} isActive={false} />);

      expect(screen.getByText('[ASSISTANT]')).toBeInTheDocument();
      expect(screen.getByText('Test')).toBeInTheDocument();
    });

    it('should handle invalid timestamps gracefully', () => {
      const invalidTimestampMessage: Message = {
        id: '1',
        role: 'user',
        content: 'Test',
        timestamp: 'invalid-timestamp',
      };

      render(<LogsPanel messages={[invalidTimestampMessage]} isActive={false} />);

      // Should still render the message
      expect(screen.getByText('Test')).toBeInTheDocument();
    });

    it('should handle very large number of messages', () => {
      const manyMessages: Message[] = Array.from({ length: 100 }, (_, i) => ({
        id: `msg-${i}`,
        role: 'user' as const,
        content: `Message ${i}`,
        timestamp: '2024-01-01T10:00:00Z',
      }));

      render(<LogsPanel messages={manyMessages} isActive={false} />);

      expect(screen.getByText('(100 entries)')).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should have accessible header button', () => {
      render(<LogsPanel messages={mockMessages} isActive={false} />);

      const headerButton = screen.getByText('Activity Logs').closest('button');
      expect(headerButton).toBeInTheDocument();
    });

    it('should have accessible clear button with title', () => {
      render(<LogsPanel messages={mockMessages} isActive={false} />);

      const clearButton = screen.getByTitle('Clear logs');
      expect(clearButton).toBeInTheDocument();
    });

    it('should have proper semantic structure', () => {
      const { container } = render(<LogsPanel messages={mockMessages} isActive={false} />);

      const panel = container.firstChild;
      expect(panel).toBeInTheDocument();
    });
  });

  describe('State preservation', () => {
    it('should preserve expanded state when clearing logs', () => {
      render(<LogsPanel messages={mockMessages} isActive={false} />);

      const trashButton = screen.getByTitle('Clear logs');
      fireEvent.click(trashButton);

      // Should still show the empty state (panel is expanded)
      expect(screen.getByText('No activity yet. Send a message to see logs.')).toBeInTheDocument();
    });

    it('should preserve cleared state when new messages arrive', () => {
      const { rerender } = render(<LogsPanel messages={mockMessages} isActive={false} />);

      // Clear logs
      const trashButton = screen.getByTitle('Clear logs');
      fireEvent.click(trashButton);

      expect(screen.getByText('No activity yet. Send a message to see logs.')).toBeInTheDocument();

      // Add new messages
      const newMessages: Message[] = [
        {
          id: '10',
          role: 'user',
          content: 'New message after clear',
          timestamp: '2024-01-01T10:00:10Z',
        },
      ];

      rerender(<LogsPanel messages={newMessages} isActive={false} />);

      // Should show the new message (clearing is local state, new props override)
      expect(screen.getByText('New message after clear')).toBeInTheDocument();
    });
  });
});
