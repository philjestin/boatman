import { Flame } from 'lucide-react';

interface FirefighterBadgeProps {
  className?: string;
  showLabel?: boolean;
}

export function FirefighterBadge({ className = '', showLabel = true }: FirefighterBadgeProps) {
  return (
    <div className={`inline-flex items-center gap-1 px-2 py-0.5 bg-red-500/10 border border-red-500/30 rounded-md ${className}`}>
      <Flame className="w-3 h-3 text-red-500" />
      {showLabel && (
        <span className="text-xs font-medium text-red-500">Firefighter</span>
      )}
    </div>
  );
}
