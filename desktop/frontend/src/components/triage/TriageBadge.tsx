import { Filter } from 'lucide-react';

interface TriageBadgeProps {
  className?: string;
}

export function TriageBadge({ className = '' }: TriageBadgeProps) {
  return (
    <span
      className={`inline-flex items-center gap-1 px-2 py-0.5 text-xs font-medium bg-amber-500/20 text-amber-300 border border-amber-500/30 rounded ${className}`}
      title="Triage Session"
    >
      <Filter className="w-3 h-3" />
      <span>Triage</span>
    </span>
  );
}
