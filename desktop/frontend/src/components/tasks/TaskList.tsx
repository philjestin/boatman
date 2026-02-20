import { TaskItem } from './TaskItem';
import { ClipboardList } from 'lucide-react';
import type { Task } from '../../types';

interface TaskListProps {
  tasks: Task[];
  onTaskClick?: (task: Task) => void;
}

export function TaskList({ tasks, onTaskClick }: TaskListProps) {
  const pendingTasks = tasks.filter((t) => t.status === 'pending');
  const inProgressTasks = tasks.filter((t) => t.status === 'in_progress');
  const completedTasks = tasks.filter((t) => t.status === 'completed');

  // Only count tasks with recognized statuses (excludes deleted/hidden tasks)
  const visibleTotal = pendingTasks.length + inProgressTasks.length + completedTasks.length;

  const progress =
    visibleTotal > 0
      ? Math.round((completedTasks.length / visibleTotal) * 100)
      : 0;

  if (visibleTotal === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-8 text-center">
        <ClipboardList className="w-12 h-12 text-slate-600 mb-3" />
        <h3 className="text-sm font-medium text-slate-300 mb-1">No tasks yet</h3>
        <p className="text-xs text-slate-500">
          Tasks will appear here as the agent works
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Progress bar */}
      <div className="space-y-2">
        <div className="flex items-center justify-between text-xs">
          <span className="text-slate-400">Progress</span>
          <span className="text-slate-300">
            {completedTasks.length} / {visibleTotal} completed
          </span>
        </div>
        <div className="h-2 bg-slate-800 rounded-full overflow-hidden">
          <div
            className="h-full bg-green-500 transition-all duration-300"
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>

      {/* In Progress */}
      {inProgressTasks.length > 0 && (
        <div>
          <h3 className="text-xs font-medium text-slate-400 uppercase tracking-wider mb-2">
            In Progress ({inProgressTasks.length})
          </h3>
          <div className="space-y-2">
            {inProgressTasks.map((task) => (
              <TaskItem
                key={task.id}
                task={task}
                onClick={onTaskClick ? () => onTaskClick(task) : undefined}
              />
            ))}
          </div>
        </div>
      )}

      {/* Pending */}
      {pendingTasks.length > 0 && (
        <div>
          <h3 className="text-xs font-medium text-slate-400 uppercase tracking-wider mb-2">
            Pending ({pendingTasks.length})
          </h3>
          <div className="space-y-2">
            {pendingTasks.map((task) => (
              <TaskItem
                key={task.id}
                task={task}
                onClick={onTaskClick ? () => onTaskClick(task) : undefined}
              />
            ))}
          </div>
        </div>
      )}

      {/* Completed */}
      {completedTasks.length > 0 && (
        <div>
          <h3 className="text-xs font-medium text-slate-400 uppercase tracking-wider mb-2">
            Completed ({completedTasks.length})
          </h3>
          <div className="space-y-2">
            {completedTasks.map((task) => (
              <TaskItem
                key={task.id}
                task={task}
                onClick={onTaskClick ? () => onTaskClick(task) : undefined}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
