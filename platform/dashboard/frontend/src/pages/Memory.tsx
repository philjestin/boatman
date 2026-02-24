import { useApi } from '../hooks/useApi';
import { PatternList } from '../components/PatternList';
import type { Pattern } from '../types';

export function Memory() {
  const { data: patterns, loading, error } = useApi<Pattern[]>('/memory/patterns');

  if (loading) return <p>Loading memory...</p>;
  if (error) return <p className="text-red-600">Error: {error}</p>;

  return (
    <div>
      <h2 className="text-xl font-bold mb-4">Shared Memory</h2>

      <div className="mb-6">
        <h3 className="text-sm font-medium text-gray-500 mb-2">
          Learned Patterns ({patterns?.length ?? 0})
        </h3>
        <PatternList patterns={patterns ?? []} />
      </div>
    </div>
  );
}
