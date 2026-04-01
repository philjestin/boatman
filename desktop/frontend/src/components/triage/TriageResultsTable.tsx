import { useState } from 'react';
import { ChevronUp, ChevronDown } from 'lucide-react';
import type { TriageClassification, TriageNormalizedTicket, TriageCategory, TriageCluster } from '../../types';
import { TriageCategoryBadge } from './TriageCategoryBadge';

interface TriageResultsTableProps {
  classifications: TriageClassification[];
  tickets: TriageNormalizedTicket[];
  clusters: TriageCluster[];
  onTicketClick: (classification: TriageClassification) => void;
}

type SortField = 'ticketId' | 'category' | 'clarity' | 'blastRadius';
type SortDir = 'asc' | 'desc';

const categoryOrder: Record<TriageCategory, number> = {
  AI_DEFINITE: 0,
  AI_LIKELY: 1,
  HUMAN_REVIEW_REQUIRED: 2,
  HUMAN_ONLY: 3,
};

export function TriageResultsTable({ classifications, tickets, clusters, onTicketClick }: TriageResultsTableProps) {
  const [sortField, setSortField] = useState<SortField>('category');
  const [sortDir, setSortDir] = useState<SortDir>('asc');
  const [filterCategory, setFilterCategory] = useState<TriageCategory | 'all'>('all');

  const ticketMap = new Map(tickets.map((t) => [t.ticketId, t]));
  const clusterMap = new Map<string, string>();
  clusters.forEach((c) => (c.tickets || []).forEach((id) => clusterMap.set(id, c.clusterId)));

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDir(sortDir === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDir('asc');
    }
  };

  const filtered = filterCategory === 'all'
    ? classifications
    : classifications.filter((c) => c.category === filterCategory);

  const sorted = [...filtered].sort((a, b) => {
    let cmp = 0;
    switch (sortField) {
      case 'ticketId':
        cmp = a.ticketId.localeCompare(b.ticketId);
        break;
      case 'category':
        cmp = categoryOrder[a.category] - categoryOrder[b.category];
        break;
      case 'clarity':
        cmp = a.rubric.clarity - b.rubric.clarity;
        break;
      case 'blastRadius':
        cmp = a.rubric.blastRadius - b.rubric.blastRadius;
        break;
    }
    return sortDir === 'asc' ? cmp : -cmp;
  });

  const SortIcon = ({ field }: { field: SortField }) => {
    if (sortField !== field) return null;
    return sortDir === 'asc'
      ? <ChevronUp className="w-3 h-3 inline ml-1" />
      : <ChevronDown className="w-3 h-3 inline ml-1" />;
  };

  if (sorted.length === 0 && filterCategory === 'all') {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="text-center space-y-2 max-w-md">
          <p className="text-slate-400 font-medium">No classifications available</p>
          <p className="text-sm text-slate-500">
            Scoring may have failed for all tickets. Check that your Anthropic API key is configured in Settings and that the key is valid.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col">
      {/* Filter */}
      <div className="flex items-center gap-2 px-4 py-2 border-b border-slate-700">
        <span className="text-xs text-slate-400">Filter:</span>
        {(['all', 'AI_DEFINITE', 'AI_LIKELY', 'HUMAN_REVIEW_REQUIRED', 'HUMAN_ONLY'] as const).map((cat) => (
          <button
            key={cat}
            onClick={() => setFilterCategory(cat)}
            className={`px-2 py-1 text-xs rounded transition-colors ${
              filterCategory === cat
                ? 'bg-slate-600 text-slate-100'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-700'
            }`}
          >
            {cat === 'all' ? 'All' : cat.replace(/_/g, ' ')}
          </button>
        ))}
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-slate-700 text-left">
              <th className="px-4 py-2 text-xs text-slate-400 cursor-pointer hover:text-slate-200" onClick={() => handleSort('ticketId')}>
                Ticket <SortIcon field="ticketId" />
              </th>
              <th className="px-4 py-2 text-xs text-slate-400">Title</th>
              <th className="px-4 py-2 text-xs text-slate-400 cursor-pointer hover:text-slate-200" onClick={() => handleSort('category')}>
                Category <SortIcon field="category" />
              </th>
              <th className="px-4 py-2 text-xs text-slate-400 cursor-pointer hover:text-slate-200" onClick={() => handleSort('clarity')}>
                Clarity <SortIcon field="clarity" />
              </th>
              <th className="px-4 py-2 text-xs text-slate-400 cursor-pointer hover:text-slate-200" onClick={() => handleSort('blastRadius')}>
                Blast <SortIcon field="blastRadius" />
              </th>
              <th className="px-4 py-2 text-xs text-slate-400">Cluster</th>
            </tr>
          </thead>
          <tbody>
            {sorted.map((c) => {
              const ticket = ticketMap.get(c.ticketId);
              const cluster = clusterMap.get(c.ticketId);
              return (
                <tr
                  key={c.ticketId}
                  className="border-b border-slate-700/50 hover:bg-slate-700/30 cursor-pointer transition-colors"
                  onClick={() => onTicketClick(c)}
                >
                  <td className="px-4 py-2 font-mono text-xs text-slate-300">{c.ticketId}</td>
                  <td className="px-4 py-2 text-slate-300 truncate max-w-xs">
                    {ticket?.title || '-'}
                  </td>
                  <td className="px-4 py-2">
                    <TriageCategoryBadge category={c.category} />
                  </td>
                  <td className="px-4 py-2 text-center text-slate-400">{c.rubric.clarity}</td>
                  <td className="px-4 py-2 text-center text-slate-400">{c.rubric.blastRadius}</td>
                  <td className="px-4 py-2 text-xs text-slate-500 truncate max-w-[120px]">
                    {cluster || '-'}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
