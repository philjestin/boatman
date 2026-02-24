import { useApi } from '../hooks/useApi';
import { CostChart } from '../components/CostChart';
import type { UsageSummary } from '../types';

export function Costs() {
  const { data: summaries, loading, error } = useApi<UsageSummary[]>('/costs/summary?group=day');

  if (loading) return <p>Loading cost data...</p>;
  if (error) return <p className="text-red-600">Error: {error}</p>;

  const totalCost = summaries?.reduce((sum, s) => sum + s.total_cost_usd, 0) ?? 0;
  const totalRuns = summaries?.reduce((sum, s) => sum + s.total_runs, 0) ?? 0;

  return (
    <div>
      <h2 className="text-xl font-bold mb-4">Cost Analytics</h2>

      <div className="grid grid-cols-3 gap-4 mb-6">
        <div className="bg-white border rounded p-4">
          <p className="text-sm text-gray-500">Total Cost</p>
          <p className="text-2xl font-bold">${totalCost.toFixed(4)}</p>
        </div>
        <div className="bg-white border rounded p-4">
          <p className="text-sm text-gray-500">Total Runs</p>
          <p className="text-2xl font-bold">{totalRuns}</p>
        </div>
        <div className="bg-white border rounded p-4">
          <p className="text-sm text-gray-500">Avg Cost/Run</p>
          <p className="text-2xl font-bold">
            ${totalRuns > 0 ? (totalCost / totalRuns).toFixed(4) : '0.0000'}
          </p>
        </div>
      </div>

      <div className="bg-white border rounded p-4">
        <h3 className="text-sm font-medium text-gray-500 mb-4">Daily Costs</h3>
        {summaries && summaries.length > 0 ? (
          <CostChart data={summaries} />
        ) : (
          <p className="text-gray-500">No cost data available.</p>
        )}
      </div>
    </div>
  );
}
