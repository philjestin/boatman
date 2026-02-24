import { useState } from 'react';
import type { Policy } from '../types';
import { apiPut } from '../hooks/useApi';

interface PolicyEditorProps {
  policy: Policy | null;
  onSave: () => void;
}

export function PolicyEditor({ policy, onSave }: PolicyEditorProps) {
  const [maxIterations, setMaxIterations] = useState(policy?.max_iterations ?? 0);
  const [maxCost, setMaxCost] = useState(policy?.max_cost_per_run ?? 0);
  const [requireTests, setRequireTests] = useState(policy?.require_tests ?? false);
  const [requireReview, setRequireReview] = useState(policy?.require_review ?? false);
  const [saving, setSaving] = useState(false);

  const handleSave = async () => {
    setSaving(true);
    try {
      await apiPut('/policies', {
        id: policy?.id ?? 'new-policy',
        max_iterations: maxIterations,
        max_cost_per_run: maxCost,
        require_tests: requireTests,
        require_review: requireReview,
      });
      onSave();
    } catch (err) {
      console.error('Save policy failed:', err);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="bg-white border rounded p-4 space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700">Max Iterations</label>
        <input
          type="number"
          value={maxIterations}
          onChange={e => setMaxIterations(Number(e.target.value))}
          className="mt-1 block w-32 border rounded px-2 py-1 text-sm"
        />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700">Max Cost Per Run ($)</label>
        <input
          type="number"
          step="0.01"
          value={maxCost}
          onChange={e => setMaxCost(Number(e.target.value))}
          className="mt-1 block w-32 border rounded px-2 py-1 text-sm"
        />
      </div>
      <div className="flex gap-4">
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={requireTests}
            onChange={e => setRequireTests(e.target.checked)}
          />
          Require Tests
        </label>
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={requireReview}
            onChange={e => setRequireReview(e.target.checked)}
          />
          Require Review
        </label>
      </div>
      <button
        onClick={handleSave}
        disabled={saving}
        className="px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
      >
        {saving ? 'Saving...' : 'Save Policy'}
      </button>
    </div>
  );
}
