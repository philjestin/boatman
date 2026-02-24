import { useApi } from '../hooks/useApi';
import type { Run } from '../types';

export function Runs() {
  const { data: runs, loading, error } = useApi<Run[]>('/runs');

  if (loading) return <p>Loading runs...</p>;
  if (error) return <p className="text-red-600">Error: {error}</p>;

  return (
    <div>
      <h2 className="text-xl font-bold mb-4">Agent Runs</h2>
      {!runs || runs.length === 0 ? (
        <p className="text-gray-500">No runs recorded yet.</p>
      ) : (
        <table className="w-full bg-white border rounded">
          <thead className="bg-gray-50">
            <tr>
              <th className="text-left px-4 py-2 text-sm font-medium text-gray-500">ID</th>
              <th className="text-left px-4 py-2 text-sm font-medium text-gray-500">Status</th>
              <th className="text-left px-4 py-2 text-sm font-medium text-gray-500">Prompt</th>
              <th className="text-right px-4 py-2 text-sm font-medium text-gray-500">Cost</th>
              <th className="text-right px-4 py-2 text-sm font-medium text-gray-500">Iterations</th>
              <th className="text-right px-4 py-2 text-sm font-medium text-gray-500">Files</th>
              <th className="text-left px-4 py-2 text-sm font-medium text-gray-500">Created</th>
            </tr>
          </thead>
          <tbody>
            {runs.map(run => (
              <tr key={run.id} className="border-t hover:bg-gray-50">
                <td className="px-4 py-2 text-sm font-mono">{run.id.slice(0, 8)}</td>
                <td className="px-4 py-2">
                  <StatusBadge status={run.status} />
                </td>
                <td className="px-4 py-2 text-sm max-w-xs truncate">{run.prompt}</td>
                <td className="px-4 py-2 text-sm text-right">${run.total_cost_usd.toFixed(4)}</td>
                <td className="px-4 py-2 text-sm text-right">{run.iterations}</td>
                <td className="px-4 py-2 text-sm text-right">{run.files_changed?.length ?? 0}</td>
                <td className="px-4 py-2 text-sm text-gray-500">
                  {new Date(run.created_at).toLocaleString()}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    passed: 'bg-green-100 text-green-800',
    running: 'bg-blue-100 text-blue-800',
    pending: 'bg-yellow-100 text-yellow-800',
    failed: 'bg-red-100 text-red-800',
    error: 'bg-red-100 text-red-800',
    canceled: 'bg-gray-100 text-gray-800',
  };
  return (
    <span className={`inline-block px-2 py-0.5 rounded text-xs ${colors[status] ?? 'bg-gray-100'}`}>
      {status}
    </span>
  );
}
