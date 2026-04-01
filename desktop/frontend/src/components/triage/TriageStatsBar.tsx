import type { TriageStats } from '../../types';

interface TriageStatsBarProps {
  stats: TriageStats;
}

export function TriageStatsBar({ stats }: TriageStatsBarProps) {
  const cards = [
    { label: 'AI Definite', count: stats.aiDefiniteCount, color: 'text-green-400', bg: 'bg-green-500/10', border: 'border-green-500/20' },
    { label: 'AI Likely', count: stats.aiLikelyCount, color: 'text-blue-400', bg: 'bg-blue-500/10', border: 'border-blue-500/20' },
    { label: 'Human Review', count: stats.humanReviewCount, color: 'text-yellow-400', bg: 'bg-yellow-500/10', border: 'border-yellow-500/20' },
    { label: 'Human Only', count: stats.humanOnlyCount, color: 'text-red-400', bg: 'bg-red-500/10', border: 'border-red-500/20' },
  ];

  return (
    <div className="grid grid-cols-4 gap-3 p-4">
      {cards.map((card) => (
        <div key={card.label} className={`${card.bg} border ${card.border} rounded-lg p-3`}>
          <p className="text-xs text-slate-400">{card.label}</p>
          <p className={`text-2xl font-bold ${card.color}`}>{card.count}</p>
        </div>
      ))}
    </div>
  );
}
