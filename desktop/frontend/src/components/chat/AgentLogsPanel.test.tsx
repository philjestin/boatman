import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AgentLogsPanel } from './AgentLogsPanel';
import type { Message, AgentInfo } from '../../types';

describe('AgentLogsPanel', () => {
  const mainAgentInfo: AgentInfo = {
    agentId: 'main',
    agentType: 'main',
  };

  const taskAgentInfo: AgentInfo = {
    agentId: 'agent-task-1',
    agentType: 'task',
    parentAgentId: 'main',
    description: 'Task agent for handling complex operations',
  };

  const exploreAgentInfo: AgentInfo = {
    agentId: 'agent-explore-1',
    agentType: 'Explore',
    parentAgentId: 'main',
    description: 'Exploration agent for analyzing codebase',
  };

  const mockMessagesMainOnly: Message[] = [
    {
      id: '1',
      role: 'user',
      content: 'Hello, Claude!',
      timestamp: '2024-01-01T10:00:00Z',
      metadata: { agent: mainAgentInfo },
    },
    {
      id: '2',
      role: 'assistant',
      content: 'Hello! How can I help you today?',
      timestamp: '2024-01-01T10:00:01Z',
      metadata: { agent: mainAgentInfo },
    },
  ];

  const mockMessagesMultipleAgents: Message[] = [
    {
      id: '1',
      role: 'user',
      content: 'Main agent message',
      timestamp: '2024-01-01T10:00:00Z',
      metadata: { agent: mainAgentInfo },
    },
    {
      id: '2',
      role: 'assistant',
      content: 'Task agent message',
      timestamp: '2024-01-01T10:00:01Z',
      metadata: { agent: taskAgentInfo },
    },
    {
      id: '3',
      role: 'assistant',
      content: 'Explore agent message',
      timestamp: '2024-01-01T10:00:02Z',
      metadata: { agent: exploreAgentInfo },
    },
  ];

  const mockToolUseMessage: Message = {
    id: '4',
    role: 'assistant',
    content: 'Reading file',
    timestamp: '2024-01-01T10:00:03Z',
    metadata: {
      agent: mainAgentInfo,
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
      agent: mainAgentInfo,
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
      render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      expect(screen.getByText('Agent Activity')).toBeInTheDocument();
    });

    it('should render with expanded state by default', () => {
      render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      expect(screen.getByText('Hello, Claude!')).toBeInTheDocument();
      expect(screen.getByText('Hello! How can I help you today?')).toBeInTheDocument();
    });

    it('should display the correct number of agents', () => {
      render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      expect(screen.getByText('(3 agents)')).toBeInTheDocument();
    });

    it('should display singular "agent" when only one agent', () => {
      render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      expect(screen.getByText('(1 agent)')).toBeInTheDocument();
    });

    it('should show Terminal icon', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      const terminalIcon = container.querySelector('.lucide-terminal');
      expect(terminalIcon).toBeInTheDocument();
    });
  });

  describe('Empty state', () => {
    it('should show empty state when no messages', () => {
      render(<AgentLogsPanel messages={[]} isActive={false} />);

      expect(screen.getByText('No activity yet. Send a message to see logs.')).toBeInTheDocument();
    });

    it('should show 0 agents when empty', () => {
      render(<AgentLogsPanel messages={[]} isActive={false} />);

      expect(screen.getByText('(0 agents)')).toBeInTheDocument();
    });
  });

  describe('Expand/Collapse functionality', () => {
    it('should toggle expanded state when header is clicked', () => {
      render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      const header = screen.getByText('Agent Activity').closest('button')!;

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
      const { container } = render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      // Initially expanded, should show ChevronDown
      let chevronDown = container.querySelector('.lucide-chevron-down');
      expect(chevronDown).toBeInTheDocument();

      const header = screen.getByText('Agent Activity').closest('button')!;
      fireEvent.click(header);

      // After collapse, should show ChevronUp
      const chevronUp = container.querySelector('.lucide-chevron-up');
      expect(chevronUp).toBeInTheDocument();
    });
  });

  describe('Active state indicator', () => {
    it('should show active indicator when isActive is true', () => {
      render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={true} />);

      expect(screen.getByText('Active')).toBeInTheDocument();
    });

    it('should not show active indicator when isActive is false', () => {
      render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      expect(screen.queryByText('Active')).not.toBeInTheDocument();
    });

    it('should add activity event to all agents when active', () => {
      render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={true} />);

      expect(screen.getByText('⚡ Working...')).toBeInTheDocument();
    });

    it('should not add activity event when inactive', () => {
      render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      expect(screen.queryByText('⚡ Working...')).not.toBeInTheDocument();
    });
  });

  describe('Clear logs functionality', () => {
    it('should clear all agents when trash button is clicked', () => {
      render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      expect(screen.getByText('Hello, Claude!')).toBeInTheDocument();

      const trashButton = screen.getByTitle('Clear logs');
      fireEvent.click(trashButton);

      expect(screen.getByText('No activity yet. Send a message to see logs.')).toBeInTheDocument();
      expect(screen.queryByText('Hello, Claude!')).not.toBeInTheDocument();
    });

    it('should update agent count to 0 after clearing', () => {
      render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      const trashButton = screen.getByTitle('Clear logs');
      fireEvent.click(trashButton);

      expect(screen.getByText('(0 agents)')).toBeInTheDocument();
    });

    it('should not expand/collapse when trash button is clicked', () => {
      render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      const trashButton = screen.getByTitle('Clear logs');

      // Logs should be visible initially
      expect(screen.getByText('Hello, Claude!')).toBeInTheDocument();

      // Click trash (stopPropagation should prevent collapse)
      fireEvent.click(trashButton);

      // Panel should still be expanded (showing empty state)
      expect(screen.getByText('No activity yet. Send a message to see logs.')).toBeInTheDocument();
    });
  });

  describe('Agent grouping and tabs', () => {
    it('should group messages by agent', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      // Should show tabs for multiple agents - look for the tab container
      const tabs = container.querySelectorAll('.flex.gap-1 button');
      expect(tabs.length).toBe(3); // main, task, explore

      // Check tab contents
      const tabTexts = Array.from(tabs).map(tab => tab.textContent);
      expect(tabTexts.some(text => text?.includes('Main'))).toBe(true);
      expect(tabTexts.some(text => text?.includes('task'))).toBe(true);
      expect(tabTexts.some(text => text?.includes('Explore'))).toBe(true);
    });

    it('should display main agent tab with star emoji when multiple agents', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const tabs = container.querySelectorAll('.flex.gap-1 button');
      const mainTab = Array.from(tabs).find(tab => tab.textContent?.includes('Main'));
      expect(mainTab).toBeTruthy();
      expect(mainTab?.textContent).toContain('⭐');
    });

    it('should display sub-agent tabs with robot emoji', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const tabs = container.querySelectorAll('.flex.gap-1 button');
      const taskTab = Array.from(tabs).find(tab => tab.textContent?.includes('task'));
      const exploreTab = Array.from(tabs).find(tab => tab.textContent?.includes('Explore'));

      expect(taskTab).toBeTruthy();
      expect(taskTab?.textContent).toContain('🤖');
      expect(exploreTab).toBeTruthy();
      expect(exploreTab?.textContent).toContain('🤖');
    });

    it('should show log count for each agent in tabs', () => {
      render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      // Each agent has 1 message
      const tabs = screen.getAllByText(/\(1\)/);
      expect(tabs.length).toBeGreaterThan(0);
    });

    it('should not show tabs when only one agent', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      // Should not render the tab container when there's only one agent
      const tabs = container.querySelectorAll('.flex.gap-1.px-2.py-2 button');
      expect(tabs.length).toBe(0);
    });

    it('should show tabs when multiple agents present', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const tabs = container.querySelectorAll('.flex.gap-1 button');
      expect(tabs.length).toBeGreaterThan(1);
    });
  });

  describe('Agent switching', () => {
    it('should start with main agent selected by default', () => {
      render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      // Main agent's message should be visible
      expect(screen.getByText('Main agent message')).toBeInTheDocument();
    });

    it('should switch to task agent when tab is clicked', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const tabs = container.querySelectorAll('.flex.gap-1 button');
      const taskTab = Array.from(tabs).find(tab => tab.textContent?.includes('task')) as HTMLButtonElement;
      fireEvent.click(taskTab);

      // Task agent's message should be visible
      expect(screen.getByText('Task agent message')).toBeInTheDocument();
    });

    it('should switch to explore agent when tab is clicked', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const tabs = container.querySelectorAll('.flex.gap-1 button');
      const exploreTab = Array.from(tabs).find(tab => tab.textContent?.includes('Explore')) as HTMLButtonElement;
      fireEvent.click(exploreTab);

      // Explore agent's message should be visible
      expect(screen.getByText('Explore agent message')).toBeInTheDocument();
    });

    it('should highlight active tab', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const tabs = container.querySelectorAll('.flex.gap-1 button');
      const mainTab = Array.from(tabs).find(tab => tab.textContent?.includes('Main')) as HTMLButtonElement;
      expect(mainTab).toHaveClass('border-blue-500');

      const taskTab = Array.from(tabs).find(tab => tab.textContent?.includes('task')) as HTMLButtonElement;
      fireEvent.click(taskTab);

      expect(taskTab).toHaveClass('border-purple-500');
    });

    it('should update displayed logs when switching agents', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      // Initially showing main agent
      expect(screen.getByText('Main agent message')).toBeInTheDocument();
      expect(screen.queryByText('Task agent message')).not.toBeInTheDocument();

      // Switch to task agent
      const tabs = container.querySelectorAll('.flex.gap-1 button');
      const taskTab = Array.from(tabs).find(tab => tab.textContent?.includes('task')) as HTMLButtonElement;
      fireEvent.click(taskTab);

      // Now showing task agent, main agent hidden
      expect(screen.queryByText('Main agent message')).not.toBeInTheDocument();
      expect(screen.getByText('Task agent message')).toBeInTheDocument();
    });
  });

  describe('Agent hierarchy display', () => {
    it('should show hierarchy toggle button when multiple agents', () => {
      render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const hierarchyButton = screen.getByTitle('Show hierarchy');
      expect(hierarchyButton).toBeInTheDocument();
    });

    it('should not show hierarchy toggle button with single agent', () => {
      render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      const hierarchyButton = screen.queryByTitle('Show hierarchy');
      expect(hierarchyButton).not.toBeInTheDocument();
    });

    it('should toggle hierarchy view when button is clicked', () => {
      render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const hierarchyButton = screen.getByTitle('Show hierarchy');

      // Initially not showing hierarchy
      expect(screen.queryByText('Main Agent')).not.toBeInTheDocument();

      // Click to show
      fireEvent.click(hierarchyButton);

      // Should show hierarchy
      expect(screen.getByText('Main Agent')).toBeInTheDocument();

      // Click to hide
      fireEvent.click(hierarchyButton);

      // Should hide hierarchy
      expect(screen.queryByText('Main Agent')).not.toBeInTheDocument();
    });

    it('should display main agent in hierarchy', () => {
      render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const hierarchyButton = screen.getByTitle('Show hierarchy');
      fireEvent.click(hierarchyButton);

      expect(screen.getByText('Main Agent')).toBeInTheDocument();
    });

    it('should display sub-agents in hierarchy', () => {
      render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const hierarchyButton = screen.getByTitle('Show hierarchy');
      fireEvent.click(hierarchyButton);

      expect(screen.getByText('task')).toBeInTheDocument();
      expect(screen.getByText('Explore')).toBeInTheDocument();
    });

    it('should show agent descriptions in hierarchy when available', () => {
      render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const hierarchyButton = screen.getByTitle('Show hierarchy');
      fireEvent.click(hierarchyButton);

      // Description should be truncated to 40 characters with ellipsis
      expect(screen.getByText((content) => {
        return content.includes('Task agent for handling complex') && content.includes('...');
      })).toBeInTheDocument();
    });

    it('should show event count for each agent in hierarchy', () => {
      render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const hierarchyButton = screen.getByTitle('Show hierarchy');
      fireEvent.click(hierarchyButton);

      const eventCounts = screen.getAllByText(/\(\d+ events?\)/);
      expect(eventCounts.length).toBeGreaterThan(0);
    });

    it('should not collapse panel when hierarchy button is clicked', () => {
      render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const hierarchyButton = screen.getByTitle('Show hierarchy');

      // Messages should be visible
      expect(screen.getByText('Main agent message')).toBeInTheDocument();

      // Click hierarchy button
      fireEvent.click(hierarchyButton);

      // Messages should still be visible
      expect(screen.getByText('Main agent message')).toBeInTheDocument();
    });

    it('should show GitBranch icon in hierarchy', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const hierarchyButton = screen.getByTitle('Show hierarchy');
      fireEvent.click(hierarchyButton);

      // Should show main agent text and GitBranch icon (lucide-git-branch class)
      expect(screen.getByText('Main Agent')).toBeInTheDocument();
      const gitBranchIcon = container.querySelector('.lucide-git-branch');
      expect(gitBranchIcon).toBeInTheDocument();
    });
  });

  describe('Message processing and log entry creation', () => {
    it('should convert user messages correctly', () => {
      const userMessage: Message = {
        id: '1',
        role: 'user',
        content: 'Test user message',
        timestamp: '2024-01-01T10:00:00Z',
        metadata: { agent: mainAgentInfo },
      };

      render(<AgentLogsPanel messages={[userMessage]} isActive={false} />);

      expect(screen.getByText('[USER]')).toBeInTheDocument();
      expect(screen.getByText('Test user message')).toBeInTheDocument();
    });

    it('should convert assistant messages correctly', () => {
      const assistantMessage: Message = {
        id: '1',
        role: 'assistant',
        content: 'Test assistant message',
        timestamp: '2024-01-01T10:00:00Z',
        metadata: { agent: mainAgentInfo },
      };

      render(<AgentLogsPanel messages={[assistantMessage]} isActive={false} />);

      expect(screen.getByText('[ASSISTANT]')).toBeInTheDocument();
      expect(screen.getByText('Test assistant message')).toBeInTheDocument();
    });

    it('should convert system messages correctly', () => {
      const systemMessage: Message = {
        id: '1',
        role: 'system',
        content: 'Test system message',
        timestamp: '2024-01-01T10:00:00Z',
        metadata: { agent: mainAgentInfo },
      };

      render(<AgentLogsPanel messages={[systemMessage]} isActive={false} />);

      expect(screen.getByText('[SYSTEM]')).toBeInTheDocument();
      expect(screen.getByText('Test system message')).toBeInTheDocument();
    });

    it('should identify tool_use messages', () => {
      render(<AgentLogsPanel messages={[mockToolUseMessage]} isActive={false} />);

      expect(screen.getByText('[TOOL→]')).toBeInTheDocument();
      expect(screen.getByText('Reading file')).toBeInTheDocument();
    });

    it('should identify tool_result messages', () => {
      render(<AgentLogsPanel messages={[mockToolResultMessage]} isActive={false} />);

      expect(screen.getByText('[←RESULT]')).toBeInTheDocument();
      expect(screen.getByText('File contents here')).toBeInTheDocument();
    });

    it('should default to main agent when no agent metadata', () => {
      const messageNoAgent: Message = {
        id: '1',
        role: 'user',
        content: 'Message without agent',
        timestamp: '2024-01-01T10:00:00Z',
      };

      render(<AgentLogsPanel messages={[messageNoAgent]} isActive={false} />);

      expect(screen.getByText('(1 agent)')).toBeInTheDocument();
      expect(screen.getByText('Message without agent')).toBeInTheDocument();
    });

    it('should associate logs with correct agent', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      // Main agent should show main message
      expect(screen.getByText('Main agent message')).toBeInTheDocument();

      // Switch to task agent
      const tabs = container.querySelectorAll('.flex.gap-1 button');
      const taskTab = Array.from(tabs).find(tab => tab.textContent?.includes('task')) as HTMLButtonElement;
      fireEvent.click(taskTab);

      // Task agent should show task message
      expect(screen.getByText('Task agent message')).toBeInTheDocument();
    });
  });

  describe('Log colors and styling', () => {
    it('should apply correct color to user messages', () => {
      const userMessage: Message = {
        id: '1',
        role: 'user',
        content: 'Test',
        timestamp: '2024-01-01T10:00:00Z',
        metadata: { agent: mainAgentInfo },
      };

      render(<AgentLogsPanel messages={[userMessage]} isActive={false} />);

      const userPrefix = screen.getByText('[USER]');
      expect(userPrefix).toHaveClass('text-blue-400');
    });

    it('should apply correct color to assistant messages', () => {
      const assistantMessage: Message = {
        id: '1',
        role: 'assistant',
        content: 'Test',
        timestamp: '2024-01-01T10:00:00Z',
        metadata: { agent: mainAgentInfo },
      };

      render(<AgentLogsPanel messages={[assistantMessage]} isActive={false} />);

      const assistantPrefix = screen.getByText('[ASSISTANT]');
      expect(assistantPrefix).toHaveClass('text-purple-400');
    });

    it('should apply correct color to system messages', () => {
      const systemMessage: Message = {
        id: '1',
        role: 'system',
        content: 'Test',
        timestamp: '2024-01-01T10:00:00Z',
        metadata: { agent: mainAgentInfo },
      };

      render(<AgentLogsPanel messages={[systemMessage]} isActive={false} />);

      const systemPrefix = screen.getByText('[SYSTEM]');
      expect(systemPrefix).toHaveClass('text-slate-400');
    });

    it('should apply correct color to tool_use messages', () => {
      render(<AgentLogsPanel messages={[mockToolUseMessage]} isActive={false} />);

      const toolUsePrefix = screen.getByText('[TOOL→]');
      expect(toolUsePrefix).toHaveClass('text-amber-400');
    });

    it('should apply correct color to tool_result messages', () => {
      render(<AgentLogsPanel messages={[mockToolResultMessage]} isActive={false} />);

      const toolResultPrefix = screen.getByText('[←RESULT]');
      expect(toolResultPrefix).toHaveClass('text-green-400');
    });

    it('should apply main agent color to tab', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const tabs = container.querySelectorAll('.flex.gap-1 button');
      const mainTab = Array.from(tabs).find(tab => tab.textContent?.includes('Main')) as HTMLButtonElement;
      expect(mainTab).toHaveClass('border-blue-500');
    });

    it('should apply task agent color to tab', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const tabs = container.querySelectorAll('.flex.gap-1 button');
      const taskTab = Array.from(tabs).find(tab => tab.textContent?.includes('task')) as HTMLButtonElement;
      fireEvent.click(taskTab);

      expect(taskTab).toHaveClass('border-purple-500');
    });

    it('should apply explore agent color to tab', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const tabs = container.querySelectorAll('.flex.gap-1 button');
      const exploreTab = Array.from(tabs).find(tab => tab.textContent?.includes('Explore')) as HTMLButtonElement;
      fireEvent.click(exploreTab);

      expect(exploreTab).toHaveClass('border-green-500');
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
        metadata: { agent: mainAgentInfo },
      };

      render(<AgentLogsPanel messages={[longMessage]} isActive={false} />);

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
        metadata: { agent: mainAgentInfo },
      };

      render(<AgentLogsPanel messages={[shortMessage]} isActive={false} />);

      expect(screen.getByText(shortContent)).toBeInTheDocument();
      expect(screen.queryByText(/\.\.\./)).not.toBeInTheDocument();
    });

    it('should truncate agent descriptions in hierarchy to 40 characters', () => {
      const longDescAgent: AgentInfo = {
        agentId: 'agent-long',
        agentType: 'task',
        description: 'This is a very long description that should be truncated to exactly forty characters for display',
      };

      const messageWithLongDesc: Message = {
        id: '1',
        role: 'assistant',
        content: 'Test',
        timestamp: '2024-01-01T10:00:00Z',
        metadata: { agent: longDescAgent },
      };

      render(<AgentLogsPanel messages={[...mockMessagesMainOnly, messageWithLongDesc]} isActive={false} />);

      const hierarchyButton = screen.getByTitle('Show hierarchy');
      fireEvent.click(hierarchyButton);

      // Check for truncated description (should have "..." at the end)
      expect(screen.getByText((content) => {
        return content.includes('This is a very long description') && content.includes('...');
      })).toBeInTheDocument();
    });
  });

  describe('Auto-scrolling behavior', () => {
    it('should have scroll container with correct height', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      const scrollContainer = container.querySelector('.h-64.overflow-y-auto');
      expect(scrollContainer).toBeInTheDocument();
    });

    it('should include ref element for auto-scroll', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      const scrollContainer = container.querySelector('.overflow-y-auto');
      expect(scrollContainer).toBeInTheDocument();
    });
  });

  describe('Message updates', () => {
    it('should update logs when messages prop changes', () => {
      const { rerender } = render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      expect(screen.getByText('(1 agent)')).toBeInTheDocument();

      const updatedMessages = [
        ...mockMessagesMainOnly,
        {
          id: '4',
          role: 'user',
          content: 'New message',
          timestamp: '2024-01-01T10:00:05Z',
          metadata: { agent: mainAgentInfo },
        } as Message,
      ];

      rerender(<AgentLogsPanel messages={updatedMessages} isActive={false} />);

      expect(screen.getByText('New message')).toBeInTheDocument();
    });

    it('should update when isActive changes', () => {
      const { rerender } = render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      expect(screen.queryByText('Active')).not.toBeInTheDocument();

      rerender(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={true} />);

      expect(screen.getByText('Active')).toBeInTheDocument();
      expect(screen.getByText('⚡ Working...')).toBeInTheDocument();
    });

    it('should add new agents when new messages arrive', () => {
      const { rerender } = render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      expect(screen.getByText('(1 agent)')).toBeInTheDocument();

      rerender(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      expect(screen.getByText('(3 agents)')).toBeInTheDocument();
    });

    it('should auto-select first agent if current selection becomes invalid', () => {
      const { rerender } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      // Switch to task agent
      const taskTab = screen.getByText(/🤖 task/).closest('button')!;
      fireEvent.click(taskTab);

      expect(screen.getByText('Task agent message')).toBeInTheDocument();

      // Remove all agents by clearing messages, then add only main
      rerender(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      // Should auto-select main agent
      expect(screen.getByText('Hello, Claude!')).toBeInTheDocument();
    });
  });

  describe('Edge cases', () => {
    it('should handle messages with empty content', () => {
      const emptyMessage: Message = {
        id: '1',
        role: 'user',
        content: '',
        timestamp: '2024-01-01T10:00:00Z',
        metadata: { agent: mainAgentInfo },
      };

      render(<AgentLogsPanel messages={[emptyMessage]} isActive={false} />);

      expect(screen.getByText('[USER]')).toBeInTheDocument();
      expect(screen.getByText('(1 agent)')).toBeInTheDocument();
    });

    it('should handle messages with special characters', () => {
      const specialMessage: Message = {
        id: '1',
        role: 'user',
        content: 'Test <script>alert("xss")</script> & special chars',
        timestamp: '2024-01-01T10:00:00Z',
        metadata: { agent: mainAgentInfo },
      };

      render(<AgentLogsPanel messages={[specialMessage]} isActive={false} />);

      expect(screen.getByText('Test <script>alert("xss")</script> & special chars')).toBeInTheDocument();
    });

    it('should handle agents without descriptions', () => {
      const noDescAgent: AgentInfo = {
        agentId: 'agent-no-desc',
        agentType: 'task',
      };

      const messageNoDesc: Message = {
        id: '1',
        role: 'assistant',
        content: 'Test',
        timestamp: '2024-01-01T10:00:00Z',
        metadata: { agent: noDescAgent },
      };

      render(<AgentLogsPanel messages={[...mockMessagesMainOnly, messageNoDesc]} isActive={false} />);

      const hierarchyButton = screen.getByTitle('Show hierarchy');
      fireEvent.click(hierarchyButton);

      // Should show agent type without description
      expect(screen.getByText('task')).toBeInTheDocument();
    });

    it('should handle undefined agent metadata', () => {
      const messageNoMetadata: Message = {
        id: '1',
        role: 'assistant',
        content: 'Test',
        timestamp: '2024-01-01T10:00:00Z',
        metadata: undefined,
      };

      render(<AgentLogsPanel messages={[messageNoMetadata]} isActive={false} />);

      // Should default to main agent
      expect(screen.getByText('(1 agent)')).toBeInTheDocument();
      expect(screen.getByText('Test')).toBeInTheDocument();
    });

    it('should handle invalid timestamps gracefully', () => {
      const invalidTimestampMessage: Message = {
        id: '1',
        role: 'user',
        content: 'Test',
        timestamp: 'invalid-timestamp',
        metadata: { agent: mainAgentInfo },
      };

      render(<AgentLogsPanel messages={[invalidTimestampMessage]} isActive={false} />);

      // Should still render the message
      expect(screen.getByText('Test')).toBeInTheDocument();
    });

    it('should handle very large number of agents', () => {
      const manyAgents: Message[] = Array.from({ length: 20 }, (_, i) => ({
        id: `msg-${i}`,
        role: 'user' as const,
        content: `Message ${i}`,
        timestamp: '2024-01-01T10:00:00Z',
        metadata: {
          agent: {
            agentId: `agent-${i}`,
            agentType: `type-${i}`,
          },
        },
      }));

      render(<AgentLogsPanel messages={manyAgents} isActive={false} />);

      expect(screen.getByText('(20 agents)')).toBeInTheDocument();
    });

    it('should handle agent type with unknown color mapping', () => {
      const unknownTypeAgent: AgentInfo = {
        agentId: 'agent-unknown',
        agentType: 'unknown-type',
      };

      const messageUnknownType: Message = {
        id: '1',
        role: 'assistant',
        content: 'Test',
        timestamp: '2024-01-01T10:00:00Z',
        metadata: { agent: unknownTypeAgent },
      };

      render(<AgentLogsPanel messages={[messageUnknownType]} isActive={false} />);

      expect(screen.getByText('Test')).toBeInTheDocument();
    });

    it('should handle Plan agent type with correct color', () => {
      const planAgent: AgentInfo = {
        agentId: 'agent-plan',
        agentType: 'Plan',
      };

      const messagePlan: Message = {
        id: '1',
        role: 'assistant',
        content: 'Planning',
        timestamp: '2024-01-01T10:00:00Z',
        metadata: { agent: planAgent },
      };

      const { container } = render(<AgentLogsPanel messages={[...mockMessagesMainOnly, messagePlan]} isActive={false} />);

      const tabs = container.querySelectorAll('.flex.gap-1 button');
      const planTab = Array.from(tabs).find(tab => tab.textContent?.includes('Plan')) as HTMLButtonElement;
      fireEvent.click(planTab);

      expect(planTab).toHaveClass('border-amber-500');
    });

    it('should handle general-purpose agent type with correct color', () => {
      const gpAgent: AgentInfo = {
        agentId: 'agent-gp',
        agentType: 'general-purpose',
      };

      const messageGP: Message = {
        id: '1',
        role: 'assistant',
        content: 'General task',
        timestamp: '2024-01-01T10:00:00Z',
        metadata: { agent: gpAgent },
      };

      const { container } = render(<AgentLogsPanel messages={[...mockMessagesMainOnly, messageGP]} isActive={false} />);

      const tabs = container.querySelectorAll('.flex.gap-1 button');
      const gpTab = Array.from(tabs).find(tab => tab.textContent?.includes('general-purpose')) as HTMLButtonElement;
      fireEvent.click(gpTab);

      expect(gpTab).toHaveClass('border-cyan-500');
    });
  });

  describe('Accessibility', () => {
    it('should have accessible header button', () => {
      render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      const headerButton = screen.getByText('Agent Activity').closest('button');
      expect(headerButton).toBeInTheDocument();
    });

    it('should have accessible clear button with title', () => {
      render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      const clearButton = screen.getByTitle('Clear logs');
      expect(clearButton).toBeInTheDocument();
    });

    it('should have accessible hierarchy button with title', () => {
      render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const hierarchyButton = screen.getByTitle('Show hierarchy');
      expect(hierarchyButton).toBeInTheDocument();
    });

    it('should have accessible agent tab buttons', () => {
      render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      const tabs = screen.getAllByRole('button');
      expect(tabs.length).toBeGreaterThan(0);
    });
  });

  describe('State preservation', () => {
    it('should preserve expanded state when clearing logs', () => {
      render(<AgentLogsPanel messages={mockMessagesMainOnly} isActive={false} />);

      const trashButton = screen.getByTitle('Clear logs');
      fireEvent.click(trashButton);

      // Should still show the empty state (panel is expanded)
      expect(screen.getByText('No activity yet. Send a message to see logs.')).toBeInTheDocument();
    });

    it('should preserve hierarchy state when switching agents', () => {
      const { container } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      // Show hierarchy
      const hierarchyButton = screen.getByTitle('Show hierarchy');
      fireEvent.click(hierarchyButton);

      expect(screen.getByText('Main Agent')).toBeInTheDocument();

      // Switch agent
      const tabs = container.querySelectorAll('.flex.gap-1 button');
      const taskTab = Array.from(tabs).find(tab => tab.textContent?.includes('task')) as HTMLButtonElement;
      fireEvent.click(taskTab);

      // Hierarchy should still be visible
      expect(screen.getByText('Main Agent')).toBeInTheDocument();
    });

    it('should preserve active agent when new messages arrive for same agent', () => {
      const { rerender } = render(<AgentLogsPanel messages={mockMessagesMultipleAgents} isActive={false} />);

      // Switch to task agent
      const taskTab = screen.getByText(/🤖 task/).closest('button')!;
      fireEvent.click(taskTab);

      expect(screen.getByText('Task agent message')).toBeInTheDocument();

      // Add more messages for task agent
      const updatedMessages = [
        ...mockMessagesMultipleAgents,
        {
          id: '10',
          role: 'assistant',
          content: 'Another task message',
          timestamp: '2024-01-01T10:00:10Z',
          metadata: { agent: taskAgentInfo },
        } as Message,
      ];

      rerender(<AgentLogsPanel messages={updatedMessages} isActive={false} />);

      // Should still be on task agent tab and see the new message
      expect(screen.getByText('Another task message')).toBeInTheDocument();
    });
  });

  describe('Tab overflow handling', () => {
    it('should have horizontal scroll for many agent tabs', () => {
      const manyAgents: Message[] = Array.from({ length: 10 }, (_, i) => ({
        id: `msg-${i}`,
        role: 'user' as const,
        content: `Message ${i}`,
        timestamp: '2024-01-01T10:00:00Z',
        metadata: {
          agent: {
            agentId: `agent-${i}`,
            agentType: `type${i}`,
          },
        },
      }));

      const { container } = render(<AgentLogsPanel messages={manyAgents} isActive={false} />);

      const tabContainer = container.querySelector('.overflow-x-auto');
      expect(tabContainer).toBeInTheDocument();
    });
  });
});
