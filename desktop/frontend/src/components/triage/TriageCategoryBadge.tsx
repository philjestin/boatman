import type { TriageCategory } from '../../types';

interface TriageCategoryBadgeProps {
  category: TriageCategory;
  className?: string;
}

const categoryStyles: Record<TriageCategory, { bg: string; text: string; border: string }> = {
  AI_DEFINITE: { bg: 'bg-green-500/20', text: 'text-green-300', border: 'border-green-500/30' },
  AI_LIKELY: { bg: 'bg-blue-500/20', text: 'text-blue-300', border: 'border-blue-500/30' },
  HUMAN_REVIEW_REQUIRED: { bg: 'bg-yellow-500/20', text: 'text-yellow-300', border: 'border-yellow-500/30' },
  HUMAN_ONLY: { bg: 'bg-red-500/20', text: 'text-red-300', border: 'border-red-500/30' },
};

const categoryLabels: Record<TriageCategory, string> = {
  AI_DEFINITE: 'AI Definite',
  AI_LIKELY: 'AI Likely',
  HUMAN_REVIEW_REQUIRED: 'Human Review',
  HUMAN_ONLY: 'Human Only',
};

export function TriageCategoryBadge({ category, className = '' }: TriageCategoryBadgeProps) {
  const style = categoryStyles[category];
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 text-xs font-medium ${style.bg} ${style.text} border ${style.border} rounded ${className}`}
    >
      {categoryLabels[category]}
    </span>
  );
}
