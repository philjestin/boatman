import { PlayCircle } from 'lucide-react';

interface BoatmanModeBadgeProps {
  className?: string;
}

export function BoatmanModeBadge({ className = '' }: BoatmanModeBadgeProps) {
  return (
    <span
      className={`inline-flex items-center gap-1 px-2 py-0.5 text-xs font-medium bg-purple-500/20 text-purple-300 border border-purple-500/30 rounded ${className}`}
      title="Boatman Mode Session"
    >
      <PlayCircle className="w-3 h-3" />
      <span>Boatman</span>
    </span>
  );
}
