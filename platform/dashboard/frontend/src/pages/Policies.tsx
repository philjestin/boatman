import { useApi } from '../hooks/useApi';
import { PolicyEditor } from '../components/PolicyEditor';
import type { Policy } from '../types';

export function Policies() {
  const { data: policy, loading, error, refetch } = useApi<Policy>('/policies');
  const { data: effective } = useApi<Policy>('/policies/effective');

  if (loading) return <p>Loading policies...</p>;
  if (error) return <p className="text-red-600">Error: {error}</p>;

  return (
    <div>
      <h2 className="text-xl font-bold mb-4">Policy Management</h2>

      <div className="grid grid-cols-2 gap-6">
        <div>
          <h3 className="text-sm font-medium text-gray-500 mb-2">Edit Policy</h3>
          <PolicyEditor policy={policy} onSave={refetch} />
        </div>

        <div>
          <h3 className="text-sm font-medium text-gray-500 mb-2">Effective Policy (Merged)</h3>
          {effective ? (
            <div className="bg-white border rounded p-4 space-y-2 text-sm">
              <p><strong>Max Iterations:</strong> {effective.max_iterations || 'unlimited'}</p>
              <p><strong>Max Cost/Run:</strong> {effective.max_cost_per_run ? `$${effective.max_cost_per_run}` : 'unlimited'}</p>
              <p><strong>Max Files Changed:</strong> {effective.max_files_changed || 'unlimited'}</p>
              <p><strong>Require Tests:</strong> {effective.require_tests ? 'Yes' : 'No'}</p>
              <p><strong>Require Review:</strong> {effective.require_review ? 'Yes' : 'No'}</p>
              {effective.allowed_models && effective.allowed_models.length > 0 && (
                <p><strong>Allowed Models:</strong> {effective.allowed_models.join(', ')}</p>
              )}
            </div>
          ) : (
            <p className="text-gray-500">No effective policy.</p>
          )}
        </div>
      </div>
    </div>
  );
}
