import { useState } from 'react';
import {
  ChevronDown,
  ChevronRight,
  Folder,
  MessageSquare,
  Clock,
  Play,
  Pause,
  Trash2,
  Circle,
  Star,
  MoreVertical,
  Tag,
} from 'lucide-react';
import type { Project, AgentSession, SessionStatus } from '../../types';
import { FirefighterBadge } from '../firefighter/FirefighterBadge';
import { BoatmanModeBadge } from '../boatmanmode/BoatmanModeBadge';

interface SidebarProps {
  projects: Project[];
  sessions: AgentSession[];
  activeSessionId: string | null;
  activeProjectId: string | null;
  onSessionSelect: (sessionId: string) => void;
  onProjectSelect: (projectId: string) => void;
  onDeleteSession: (sessionId: string) => void;
  onToggleFavorite?: (sessionId: string) => void;
  onAddTag?: (sessionId: string, tag: string) => void;
  onRemoveTag?: (sessionId: string, tag: string) => void;
  isOpen: boolean;
}

function getStatusColor(status: SessionStatus): string {
  switch (status) {
    case 'running':
      return 'text-green-500';
    case 'waiting':
      return 'text-yellow-500';
    case 'error':
      return 'text-red-500';
    case 'stopped':
      return 'text-slate-400';
    default:
      return 'text-slate-500';
  }
}

function getStatusIcon(status: SessionStatus) {
  switch (status) {
    case 'running':
      return <Play className="w-3 h-3" />;
    case 'waiting':
      return <Pause className="w-3 h-3" />;
    default:
      return <Circle className="w-3 h-3" />;
  }
}

export function Sidebar({
  projects,
  sessions,
  activeSessionId,
  activeProjectId,
  onSessionSelect,
  onProjectSelect,
  onDeleteSession,
  onToggleFavorite,
  onAddTag,
  onRemoveTag,
  isOpen,
}: SidebarProps) {
  const [projectsExpanded, setProjectsExpanded] = useState(true);
  const [sessionsExpanded, setSessionsExpanded] = useState(true);
  const [menuOpenSession, setMenuOpenSession] = useState<string | null>(null);
  const [tagInputSession, setTagInputSession] = useState<string | null>(null);
  const [newTagValue, setNewTagValue] = useState('');

  if (!isOpen) {
    return null;
  }

  return (
    <aside className="w-64 bg-slate-900 border-r border-slate-700 flex flex-col h-full overflow-hidden">
      {/* Projects Section */}
      <div className="flex-shrink-0">
        <button
          onClick={() => setProjectsExpanded(!projectsExpanded)}
          className="w-full flex items-center gap-2 px-3 py-2 text-xs font-medium text-slate-400 uppercase tracking-wider hover:bg-slate-800"
        >
          {projectsExpanded ? (
            <ChevronDown className="w-3 h-3" />
          ) : (
            <ChevronRight className="w-3 h-3" />
          )}
          Projects
        </button>
        {projectsExpanded && (
          <div className="px-2 pb-2">
            {projects.length === 0 ? (
              <p className="px-2 py-1 text-xs text-slate-500">No projects yet</p>
            ) : (
              projects.map((project) => (
                <button
                  key={project.id}
                  onClick={() => onProjectSelect(project.id)}
                  className={`w-full flex items-center gap-2 px-2 py-1.5 text-sm rounded-md transition-colors ${
                    activeProjectId === project.id
                      ? 'bg-slate-700 text-slate-100'
                      : 'text-slate-300 hover:bg-slate-800 hover:text-slate-200'
                  }`}
                >
                  <Folder className="w-4 h-4 flex-shrink-0" />
                  <span className="truncate">{project.name}</span>
                </button>
              ))
            )}
          </div>
        )}
      </div>

      {/* Sessions Section */}
      <div className="flex-1 overflow-y-auto">
        <button
          onClick={() => setSessionsExpanded(!sessionsExpanded)}
          className="w-full flex items-center gap-2 px-3 py-2 text-xs font-medium text-slate-400 uppercase tracking-wider hover:bg-slate-800"
        >
          {sessionsExpanded ? (
            <ChevronDown className="w-3 h-3" />
          ) : (
            <ChevronRight className="w-3 h-3" />
          )}
          Sessions
        </button>
        {sessionsExpanded && (
          <div className="px-2 pb-2">
            {sessions.length === 0 ? (
              <p className="px-2 py-1 text-xs text-slate-500">No active sessions</p>
            ) : (
              sessions.map((session) => (
                <div key={session.id} className="relative">
                  <div
                    className={`group flex items-start gap-2 px-2 py-1.5 rounded-md transition-colors cursor-pointer ${
                      activeSessionId === session.id
                        ? 'bg-slate-700'
                        : 'hover:bg-slate-800'
                    }`}
                    onClick={() => onSessionSelect(session.id)}
                  >
                    <span className={`${getStatusColor(session.status)} mt-0.5`}>
                      {getStatusIcon(session.status)}
                    </span>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-1">
                        <p
                          className={`text-sm truncate flex-1 ${
                            activeSessionId === session.id
                              ? 'text-slate-100'
                              : 'text-slate-300'
                          }`}
                        >
                          {session.projectPath.split('/').pop() || 'Session'}
                        </p>
                        {session.isFavorite && onToggleFavorite && (
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              onToggleFavorite(session.id);
                            }}
                            className="p-0.5 flex-shrink-0"
                            aria-label="Unfavorite session"
                          >
                            <Star className="w-3 h-3 text-yellow-500 fill-yellow-500" />
                          </button>
                        )}
                        {!session.isFavorite && onToggleFavorite && (
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              onToggleFavorite(session.id);
                            }}
                            className="p-0.5 flex-shrink-0 opacity-0 group-hover:opacity-100"
                            aria-label="Favorite session"
                          >
                            <Star className="w-3 h-3 text-slate-400 hover:text-yellow-500" />
                          </button>
                        )}
                      </div>
                      <div className="flex items-center gap-2 mt-0.5">
                        <p className="text-xs text-slate-500 flex items-center gap-1">
                          <Clock className="w-3 h-3" />
                          {new Date(session.createdAt).toLocaleTimeString()}
                        </p>
                        {session.mode === 'firefighter' && (
                          <FirefighterBadge showLabel={false} className="flex-shrink-0" />
                        )}
                        {session.mode === 'boatmanmode' && (
                          <BoatmanModeBadge className="flex-shrink-0" />
                        )}
                      </div>
                      {session.tags && session.tags.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          {session.tags.map((tag) => (
                            <span
                              key={tag}
                              className="inline-flex items-center gap-0.5 px-1.5 py-0.5 text-xs bg-slate-600 text-slate-300 rounded"
                              onClick={(e) => e.stopPropagation()}
                            >
                              {tag}
                              {onRemoveTag && (
                                <button
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    onRemoveTag(session.id, tag);
                                  }}
                                  className="hover:text-red-400"
                                  aria-label={`Remove tag ${tag}`}
                                >
                                  ×
                                </button>
                              )}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 flex-shrink-0">
                      {onAddTag && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            setMenuOpenSession(session.id === menuOpenSession ? null : session.id);
                          }}
                          className="p-1 hover:bg-slate-600 rounded transition-all"
                          aria-label="Manage tags"
                        >
                          <MoreVertical className="w-3 h-3 text-slate-400" />
                        </button>
                      )}
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          onDeleteSession(session.id);
                        }}
                        className="p-1 hover:bg-slate-600 rounded transition-all"
                        aria-label="Delete session"
                      >
                        <Trash2 className="w-3 h-3 text-slate-400 hover:text-red-500" />
                      </button>
                    </div>
                  </div>
                  {/* Tag Menu */}
                  {menuOpenSession === session.id && onAddTag && (
                    <div className="absolute right-2 top-full mt-1 z-10 bg-slate-800 border border-slate-700 rounded-lg shadow-lg p-2 w-48">
                      {tagInputSession === session.id ? (
                        <div className="flex gap-1">
                          <input
                            type="text"
                            value={newTagValue}
                            onChange={(e) => setNewTagValue(e.target.value)}
                            onKeyDown={(e) => {
                              if (e.key === 'Enter' && newTagValue.trim()) {
                                onAddTag(session.id, newTagValue.trim());
                                setNewTagValue('');
                                setTagInputSession(null);
                                setMenuOpenSession(null);
                              } else if (e.key === 'Escape') {
                                setNewTagValue('');
                                setTagInputSession(null);
                              }
                            }}
                            placeholder="Enter tag..."
                            className="flex-1 px-2 py-1 text-xs bg-slate-900 border border-slate-600 rounded text-slate-100 outline-none focus:border-blue-500"
                            autoFocus
                            onClick={(e) => e.stopPropagation()}
                          />
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              setNewTagValue('');
                              setTagInputSession(null);
                            }}
                            className="px-2 py-1 text-xs text-slate-400 hover:text-slate-200"
                          >
                            ×
                          </button>
                        </div>
                      ) : (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            setTagInputSession(session.id);
                          }}
                          className="w-full flex items-center gap-2 px-2 py-1.5 text-xs text-slate-300 hover:bg-slate-700 rounded transition-colors"
                        >
                          <Tag className="w-3 h-3" />
                          Add Tag
                        </button>
                      )}
                    </div>
                  )}
                </div>
              ))
            )}
          </div>
        )}
      </div>

      {/* Recent Activity */}
      <div className="flex-shrink-0 border-t border-slate-700 p-3">
        <p className="text-xs text-slate-500 mb-2">Recent Activity</p>
        {projects.slice(0, 3).map((project) => (
          <button
            key={project.id}
            onClick={() => onProjectSelect(project.id)}
            className="w-full flex items-center gap-2 px-2 py-1 text-xs text-slate-400 hover:text-slate-200 hover:bg-slate-800 rounded transition-colors"
          >
            <MessageSquare className="w-3 h-3" />
            <span className="truncate">{project.name}</span>
          </button>
        ))}
      </div>
    </aside>
  );
}
