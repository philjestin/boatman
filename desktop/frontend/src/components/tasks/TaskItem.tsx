import { CheckCircle2, Circle, Loader2, ChevronRight } from 'lucide-react';
import type { Task } from '../../types';

interface TaskItemProps {
  task: Task;
  onClick?: () => void;
}

function getStatusIcon(status: Task['status']) {
  switch (status) {
    case 'completed':
      return <CheckCircle2 className="w-4 h-4 text-green-500" />;
    case 'in_progress':
      return <Loader2 className="w-4 h-4 text-blue-500 animate-spin" />;
    default:
      return <Circle className="w-4 h-4 text-slate-500" />;
  }
}

function getStatusLabel(status: Task['status']) {
  switch (status) {
    case 'completed':
      return 'Completed';
    case 'in_progress':
      return 'In Progress';
    default:
      return 'Pending';
  }
}

function getStatusStyles(status: Task['status']) {
  switch (status) {
    case 'completed':
      return 'bg-green-500/10 text-green-500';
    case 'in_progress':
      return 'bg-blue-500/10 text-blue-500';
    default:
      return 'bg-slate-700 text-slate-400';
  }
}

export function TaskItem({ task, onClick }: TaskItemProps) {
  return (
    <div
      onClick={onClick}
      className={`flex items-start gap-3 p-3 rounded-lg border transition-colors ${
        onClick ? 'cursor-pointer hover:bg-slate-800' : ''
      } ${
        task.status === 'completed'
          ? 'border-slate-800 bg-slate-900/50'
          : 'border-slate-700 bg-slate-800'
      }`}
    >
      <div className="flex-shrink-0 mt-0.5">{getStatusIcon(task.status)}</div>
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between gap-2">
          <h4
            className={`text-sm font-medium ${
              task.status === 'completed' ? 'text-slate-400 line-through' : 'text-slate-100'
            }`}
          >
            {task.subject}
          </h4>
          <span
            className={`flex-shrink-0 text-xs px-2 py-0.5 rounded-full ${getStatusStyles(
              task.status
            )}`}
          >
            {getStatusLabel(task.status)}
          </span>
        </div>
        {task.description && (
          <p className="mt-1 text-xs text-slate-400 line-clamp-2">{task.description}</p>
        )}
      </div>
      {onClick && (
        <ChevronRight className="flex-shrink-0 w-4 h-4 text-slate-500 mt-0.5" />
      )}
    </div>
  );
}
