import type { Pattern } from '../types';

interface PatternListProps {
  patterns: Pattern[];
}

export function PatternList({ patterns }: PatternListProps) {
  if (patterns.length === 0) {
    return <p className="text-gray-500">No patterns learned yet.</p>;
  }

  return (
    <div className="space-y-2">
      {patterns.map(p => (
        <div key={p.id} className="bg-white border rounded p-3">
          <div className="flex justify-between items-start">
            <div>
              <span className="inline-block px-2 py-0.5 bg-blue-100 text-blue-800 text-xs rounded mr-2">
                {p.type}
              </span>
              <span className="text-sm font-medium">{p.description}</span>
            </div>
            <span className="text-xs text-gray-500">
              weight: {p.weight.toFixed(2)}
            </span>
          </div>
          {p.example && (
            <p className="text-xs text-gray-600 mt-1 font-mono">{p.example}</p>
          )}
        </div>
      ))}
    </div>
  );
}
