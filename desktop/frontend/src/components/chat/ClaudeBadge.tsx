import { Terminal } from 'lucide-react';

interface ClaudeBadgeProps {
  className?: string;
}

export function ClaudeBadge({ className = '' }: ClaudeBadgeProps) {
  return (
    <span className={`inline-flex items-center gap-0.5 px-1 py-0.5 text-xs bg-blue-900/40 text-blue-400 rounded ${className}`}>
      <Terminal className="w-2.5 h-2.5" />
    </span>
  );
}
