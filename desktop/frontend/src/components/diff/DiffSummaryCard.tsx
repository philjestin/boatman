import { FilePlus, FileMinus, FileEdit, AlertTriangle, CheckCircle2, AlertCircle } from 'lucide-react';
import type { DiffSummary } from '../../utils/diffUtils';

interface DiffSummaryCardProps {
  summary: DiffSummary;
}

export function DiffSummaryCard({ summary }: DiffSummaryCardProps) {
  const getRiskColor = (level: 'low' | 'medium' | 'high') => {
    switch (level) {
      case 'low':
        return 'text-green-500';
      case 'medium':
        return 'text-yellow-500';
      case 'high':
        return 'text-red-500';
    }
  };

  const getRiskIcon = (level: 'low' | 'medium' | 'high') => {
    switch (level) {
      case 'low':
        return <CheckCircle2 className="w-5 h-5" />;
      case 'medium':
        return <AlertCircle className="w-5 h-5" />;
      case 'high':
        return <AlertTriangle className="w-5 h-5" />;
    }
  };

  const getRiskLabel = (level: 'low' | 'medium' | 'high') => {
    switch (level) {
      case 'low':
        return 'Low Risk';
      case 'medium':
        return 'Medium Risk';
      case 'high':
        return 'High Risk';
    }
  };

  return (
    <div className="bg-slate-900 border border-slate-700 rounded-lg p-4">
      <h3 className="text-sm font-medium text-slate-200 mb-3">Change Summary</h3>

      <div className="space-y-3">
        {/* File counts */}
        <div className="grid grid-cols-3 gap-3">
          <div className="flex items-center gap-2">
            <FilePlus className="w-4 h-4 text-green-500" />
            <div>
              <div className="text-xs text-slate-400">Added</div>
              <div className="text-sm font-medium text-slate-200">
                {summary.filesAdded}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <FileEdit className="w-4 h-4 text-yellow-500" />
            <div>
              <div className="text-xs text-slate-400">Modified</div>
              <div className="text-sm font-medium text-slate-200">
                {summary.filesModified}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <FileMinus className="w-4 h-4 text-red-500" />
            <div>
              <div className="text-xs text-slate-400">Deleted</div>
              <div className="text-sm font-medium text-slate-200">
                {summary.filesDeleted}
              </div>
            </div>
          </div>
        </div>

        {/* Line counts */}
        <div className="flex items-center justify-between pt-2 border-t border-slate-700">
          <div>
            <div className="text-xs text-slate-400">Lines Changed</div>
            <div className="text-sm text-slate-200 mt-1">
              <span className="text-green-500">+{summary.linesAdded}</span>
              {' / '}
              <span className="text-red-500">-{summary.linesDeleted}</span>
            </div>
          </div>

          <div className={`flex items-center gap-1.5 ${getRiskColor(summary.riskLevel)}`}>
            {getRiskIcon(summary.riskLevel)}
            <span className="text-sm font-medium">
              {getRiskLabel(summary.riskLevel)}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
