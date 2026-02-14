import { useEffect, useCallback } from 'react';
import { useStore } from '../store';
import type { Project } from '../types';

// Import Wails bindings
import {
  OpenProject,
  RemoveProject,
  ListProjects,
  GetRecentProjects,
  SelectFolder,
  GetWorkspaceInfo,
  GetGitStatus,
} from '../../wailsjs/go/main/App';

export function useProject() {
  const {
    projects,
    activeProjectId,
    setProjects,
    addProject,
    removeProject,
    setActiveProject,
    setLoading,
    setError,
  } = useStore();

  // Load projects on mount
  useEffect(() => {
    const loadProjects = async () => {
      try {
        setLoading('projects', true);
        const projectList = await ListProjects();
        setProjects(projectList);
      } catch (err) {
        setError('Failed to load projects');
      } finally {
        setLoading('projects', false);
      }
    };

    loadProjects();
  }, [setProjects, setLoading, setError]);

  // Open a project by path
  const openProject = useCallback(async (path: string): Promise<Project | null> => {
    try {
      setLoading('projects', true);
      const project = await OpenProject(path);
      addProject(project);
      setActiveProject(project.id);
      return project;
    } catch (err) {
      setError('Failed to open project');
      return null;
    } finally {
      setLoading('projects', false);
    }
  }, [addProject, setActiveProject, setLoading, setError]);

  // Open folder dialog and select a project
  const selectAndOpenProject = useCallback(async (): Promise<Project | null> => {
    try {
      const path = await SelectFolder();
      if (!path) return null;
      return await openProject(path);
    } catch (err) {
      setError('Failed to select folder');
      return null;
    }
  }, [openProject, setError]);

  // Remove a project
  const deleteProject = useCallback(async (projectId: string) => {
    try {
      await RemoveProject(projectId);
      removeProject(projectId);
    } catch (err) {
      setError('Failed to remove project');
    }
  }, [removeProject, setError]);

  // Select a project
  const selectProject = useCallback((projectId: string) => {
    setActiveProject(projectId);
  }, [setActiveProject]);

  // Get workspace info for a project
  const getWorkspaceInfo = useCallback(async (path: string) => {
    try {
      return await GetWorkspaceInfo(path);
    } catch (err) {
      console.error('Failed to get workspace info:', err);
      return null;
    }
  }, []);

  // Get git status for a project
  const getGitStatus = useCallback(async (path: string) => {
    try {
      return await GetGitStatus(path);
    } catch (err) {
      console.error('Failed to get git status:', err);
      return null;
    }
  }, []);

  // Get active project
  const activeProject = projects.find((p) => p.id === activeProjectId) ?? null;

  // Get recent projects
  const recentProjects = projects.slice(0, 5);

  return {
    projects,
    activeProject,
    activeProjectId,
    recentProjects,
    openProject,
    selectAndOpenProject,
    deleteProject,
    selectProject,
    getWorkspaceInfo,
    getGitStatus,
  };
}
