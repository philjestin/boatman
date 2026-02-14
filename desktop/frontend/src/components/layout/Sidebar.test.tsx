import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, within } from '@testing-library/react';
import { Sidebar } from './Sidebar';
import type { Project, AgentSession } from '../../types';

describe('Sidebar', () => {
  const mockOnSessionSelect = vi.fn();
  const mockOnProjectSelect = vi.fn();
  const mockOnDeleteSession = vi.fn();
  const mockOnToggleFavorite = vi.fn();
  const mockOnAddTag = vi.fn();
  const mockOnRemoveTag = vi.fn();

  const mockProjects: Project[] = [
    {
      id: 'proj-1',
      name: 'Project 1',
      path: '/path/to/project1',
      lastOpened: '2024-01-15T10:00:00Z',
      createdAt: '2024-01-01T10:00:00Z',
    },
    {
      id: 'proj-2',
      name: 'Project 2',
      path: '/path/to/project2',
      lastOpened: '2024-01-14T10:00:00Z',
      createdAt: '2024-01-02T10:00:00Z',
    },
  ];

  const mockSessions: AgentSession[] = [
    {
      id: 'session-1',
      projectPath: '/path/to/project1',
      status: 'idle',
      createdAt: '2024-01-15T10:00:00Z',
      messages: [],
      tasks: [],
      tags: ['frontend', 'bug'],
      isFavorite: true,
    },
    {
      id: 'session-2',
      projectPath: '/path/to/project2',
      status: 'running',
      createdAt: '2024-01-15T11:00:00Z',
      messages: [],
      tasks: [],
      tags: [],
      isFavorite: false,
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Visibility', () => {
    it('should not render when isOpen is false', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={false}
        />
      );

      expect(screen.queryByText('Projects')).not.toBeInTheDocument();
    });

    it('should render when isOpen is true', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      expect(screen.getByText('Projects')).toBeInTheDocument();
      expect(screen.getByText('Sessions')).toBeInTheDocument();
    });
  });

  describe('Projects Section', () => {
    it('should render all projects', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      const projectButtons = screen.getAllByText('Project 1');
      expect(projectButtons.length).toBeGreaterThan(0);
      expect(screen.getAllByText('Project 2').length).toBeGreaterThan(0);
    });

    it('should call onProjectSelect when project is clicked', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      const projectButtons = screen.getAllByText('Project 1');
      fireEvent.click(projectButtons[0]);
      expect(mockOnProjectSelect).toHaveBeenCalledWith('proj-1');
    });

    it('should highlight active project', () => {
      const { container } = render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId="proj-1"
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      const activeButtons = container.querySelectorAll('.bg-slate-700');
      expect(activeButtons.length).toBeGreaterThan(0);
    });

    it('should show empty state when no projects', () => {
      render(
        <Sidebar
          projects={[]}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      expect(screen.getByText('No projects yet')).toBeInTheDocument();
    });

    it('should collapse/expand projects section', () => {
      const { container } = render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      const projectsHeader = screen.getByText('Projects');
      const initialProjectCount = screen.getAllByText('Project 1').length;

      fireEvent.click(projectsHeader);

      // After collapse, should have fewer instances (only in Recent Activity)
      const collapsedProjectCount = screen.getAllByText('Project 1').length;
      expect(collapsedProjectCount).toBeLessThan(initialProjectCount);

      fireEvent.click(projectsHeader);

      // After expand, should have original count
      const expandedProjectCount = screen.getAllByText('Project 1').length;
      expect(expandedProjectCount).toBe(initialProjectCount);
    });
  });

  describe('Sessions Section', () => {
    it('should render all sessions', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      expect(screen.getByText('project1')).toBeInTheDocument();
      expect(screen.getByText('project2')).toBeInTheDocument();
    });

    it('should call onSessionSelect when session is clicked', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      fireEvent.click(screen.getByText('project1'));
      expect(mockOnSessionSelect).toHaveBeenCalledWith('session-1');
    });

    it('should highlight active session', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId="session-1"
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      const session1Container = screen.getByText('project1').closest('.bg-slate-700');
      expect(session1Container).toBeInTheDocument();
    });

    it('should show status icons', () => {
      const { container } = render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      const playIcons = container.querySelectorAll('.lucide-play');
      expect(playIcons.length).toBeGreaterThan(0);
    });

    it('should call onDeleteSession when delete button is clicked', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      const deleteButtons = screen.getAllByLabelText('Delete session');
      fireEvent.click(deleteButtons[0]);

      expect(mockOnDeleteSession).toHaveBeenCalledWith('session-1');
      expect(mockOnSessionSelect).not.toHaveBeenCalled();
    });

    it('should show empty state when no sessions', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={[]}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      expect(screen.getByText('No active sessions')).toBeInTheDocument();
    });
  });

  describe('Favorite Functionality', () => {
    it('should show filled star for favorite sessions', () => {
      const { container } = render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          onToggleFavorite={mockOnToggleFavorite}
          isOpen={true}
        />
      );

      const stars = container.querySelectorAll('.fill-yellow-500');
      expect(stars.length).toBe(1);
    });

    it('should call onToggleFavorite when star is clicked on favorite session', () => {
      const { container } = render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          onToggleFavorite={mockOnToggleFavorite}
          isOpen={true}
        />
      );

      const favoriteButton = screen.getByLabelText('Unfavorite session');
      fireEvent.click(favoriteButton);

      expect(mockOnToggleFavorite).toHaveBeenCalledWith('session-1');
      expect(mockOnSessionSelect).not.toHaveBeenCalled();
    });

    it('should show empty star on hover for non-favorite sessions', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          onToggleFavorite={mockOnToggleFavorite}
          isOpen={true}
        />
      );

      const favoriteButton = screen.getByLabelText('Favorite session');
      expect(favoriteButton).toHaveClass('opacity-0');
    });

    it('should call onToggleFavorite when star is clicked on non-favorite session', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          onToggleFavorite={mockOnToggleFavorite}
          isOpen={true}
        />
      );

      const favoriteButton = screen.getByLabelText('Favorite session');
      fireEvent.click(favoriteButton);

      expect(mockOnToggleFavorite).toHaveBeenCalledWith('session-2');
    });

    it('should not show star icons when onToggleFavorite is not provided', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      expect(screen.queryByLabelText('Favorite session')).not.toBeInTheDocument();
      expect(screen.queryByLabelText('Unfavorite session')).not.toBeInTheDocument();
    });
  });

  describe('Tag Display', () => {
    it('should display session tags', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      expect(screen.getByText('frontend')).toBeInTheDocument();
      expect(screen.getByText('bug')).toBeInTheDocument();
    });

    it('should not show tags section when session has no tags', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      // Session 2 has no tags, so we check it doesn't have tag pills
      const session2 = screen.getByText('project2').closest('div');
      const tagPills = session2?.querySelectorAll('.bg-slate-600');
      expect(tagPills?.length || 0).toBe(0);
    });

    it('should call onRemoveTag when tag × is clicked', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          onRemoveTag={mockOnRemoveTag}
          isOpen={true}
        />
      );

      const removeTagButton = screen.getByLabelText('Remove tag frontend');
      fireEvent.click(removeTagButton);

      expect(mockOnRemoveTag).toHaveBeenCalledWith('session-1', 'frontend');
      expect(mockOnSessionSelect).not.toHaveBeenCalled();
    });

    it('should not show × on tags when onRemoveTag is not provided', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      expect(screen.queryByLabelText('Remove tag frontend')).not.toBeInTheDocument();
    });
  });

  describe('Tag Management Menu', () => {
    it('should show menu button when onAddTag is provided', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          onAddTag={mockOnAddTag}
          isOpen={true}
        />
      );

      const menuButtons = screen.getAllByLabelText('Manage tags');
      expect(menuButtons.length).toBe(2);
    });

    it('should open tag menu when menu button is clicked', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          onAddTag={mockOnAddTag}
          isOpen={true}
        />
      );

      const menuButtons = screen.getAllByLabelText('Manage tags');
      fireEvent.click(menuButtons[0]);

      expect(screen.getByText('Add Tag')).toBeInTheDocument();
    });

    it('should show tag input when "Add Tag" is clicked', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          onAddTag={mockOnAddTag}
          isOpen={true}
        />
      );

      const menuButtons = screen.getAllByLabelText('Manage tags');
      fireEvent.click(menuButtons[0]);

      const addTagButton = screen.getByText('Add Tag');
      fireEvent.click(addTagButton);

      expect(screen.getByPlaceholderText('Enter tag...')).toBeInTheDocument();
    });

    it('should call onAddTag when Enter is pressed in tag input', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          onAddTag={mockOnAddTag}
          isOpen={true}
        />
      );

      const menuButtons = screen.getAllByLabelText('Manage tags');
      fireEvent.click(menuButtons[0]);

      const addTagButton = screen.getByText('Add Tag');
      fireEvent.click(addTagButton);

      const input = screen.getByPlaceholderText('Enter tag...');
      fireEvent.change(input, { target: { value: 'new-tag' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(mockOnAddTag).toHaveBeenCalledWith('session-1', 'new-tag');
    });

    it('should not call onAddTag when Enter is pressed with empty input', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          onAddTag={mockOnAddTag}
          isOpen={true}
        />
      );

      const menuButtons = screen.getAllByLabelText('Manage tags');
      fireEvent.click(menuButtons[0]);

      const addTagButton = screen.getByText('Add Tag');
      fireEvent.click(addTagButton);

      const input = screen.getByPlaceholderText('Enter tag...');
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(mockOnAddTag).not.toHaveBeenCalled();
    });

    it('should close tag input when Escape is pressed', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          onAddTag={mockOnAddTag}
          isOpen={true}
        />
      );

      const menuButtons = screen.getAllByLabelText('Manage tags');
      fireEvent.click(menuButtons[0]);

      const addTagButton = screen.getByText('Add Tag');
      fireEvent.click(addTagButton);

      const input = screen.getByPlaceholderText('Enter tag...');
      fireEvent.change(input, { target: { value: 'new-tag' } });
      fireEvent.keyDown(input, { key: 'Escape' });

      expect(screen.queryByPlaceholderText('Enter tag...')).not.toBeInTheDocument();
      expect(mockOnAddTag).not.toHaveBeenCalled();
    });

    it('should close menu after adding tag', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          onAddTag={mockOnAddTag}
          isOpen={true}
        />
      );

      const menuButtons = screen.getAllByLabelText('Manage tags');
      fireEvent.click(menuButtons[0]);

      const addTagButton = screen.getByText('Add Tag');
      fireEvent.click(addTagButton);

      const input = screen.getByPlaceholderText('Enter tag...');
      fireEvent.change(input, { target: { value: 'new-tag' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(screen.queryByPlaceholderText('Enter tag...')).not.toBeInTheDocument();
      expect(screen.queryByText('Add Tag')).not.toBeInTheDocument();
    });

    it('should toggle menu when menu button is clicked twice', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          onAddTag={mockOnAddTag}
          isOpen={true}
        />
      );

      const menuButtons = screen.getAllByLabelText('Manage tags');
      fireEvent.click(menuButtons[0]);
      expect(screen.getByText('Add Tag')).toBeInTheDocument();

      fireEvent.click(menuButtons[0]);
      expect(screen.queryByText('Add Tag')).not.toBeInTheDocument();
    });
  });

  describe('Recent Activity', () => {
    it('should show recent activity section', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      expect(screen.getByText('Recent Activity')).toBeInTheDocument();
    });

    it('should show up to 3 recent projects', () => {
      const manyProjects = [
        ...mockProjects,
        {
          id: 'proj-3',
          name: 'Project 3',
          path: '/path/to/project3',
          lastOpened: '2024-01-13T10:00:00Z',
          createdAt: '2024-01-03T10:00:00Z',
        },
        {
          id: 'proj-4',
          name: 'Project 4',
          path: '/path/to/project4',
          lastOpened: '2024-01-12T10:00:00Z',
          createdAt: '2024-01-04T10:00:00Z',
        },
      ];

      render(
        <Sidebar
          projects={manyProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      const recentSection = screen.getByText('Recent Activity').parentElement;
      const projectButtons = within(recentSection!).getAllByRole('button');
      expect(projectButtons.length).toBe(3);
    });
  });

  describe('Edge Cases', () => {
    it('should handle sessions without tags array', () => {
      const sessionsWithoutTags: AgentSession[] = [
        {
          id: 'session-1',
          projectPath: '/path/to/project1',
          status: 'idle',
          createdAt: '2024-01-15T10:00:00Z',
          messages: [],
          tasks: [],
        },
      ];

      render(
        <Sidebar
          projects={mockProjects}
          sessions={sessionsWithoutTags}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          isOpen={true}
        />
      );

      expect(screen.getByText('project1')).toBeInTheDocument();
    });

    it('should handle sessions without isFavorite property', () => {
      const sessionsWithoutFavorite: AgentSession[] = [
        {
          id: 'session-1',
          projectPath: '/path/to/project1',
          status: 'idle',
          createdAt: '2024-01-15T10:00:00Z',
          messages: [],
          tasks: [],
        },
      ];

      render(
        <Sidebar
          projects={mockProjects}
          sessions={sessionsWithoutFavorite}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          onToggleFavorite={mockOnToggleFavorite}
          isOpen={true}
        />
      );

      expect(screen.getByLabelText('Favorite session')).toBeInTheDocument();
    });

    it('should trim whitespace when adding tags', () => {
      render(
        <Sidebar
          projects={mockProjects}
          sessions={mockSessions}
          activeSessionId={null}
          activeProjectId={null}
          onSessionSelect={mockOnSessionSelect}
          onProjectSelect={mockOnProjectSelect}
          onDeleteSession={mockOnDeleteSession}
          onAddTag={mockOnAddTag}
          isOpen={true}
        />
      );

      const menuButtons = screen.getAllByLabelText('Manage tags');
      fireEvent.click(menuButtons[0]);

      const addTagButton = screen.getByText('Add Tag');
      fireEvent.click(addTagButton);

      const input = screen.getByPlaceholderText('Enter tag...');
      fireEvent.change(input, { target: { value: '  spaced-tag  ' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(mockOnAddTag).toHaveBeenCalledWith('session-1', 'spaced-tag');
    });
  });
});
