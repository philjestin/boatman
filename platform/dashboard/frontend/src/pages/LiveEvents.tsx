import { useSSE } from '../hooks/useSSE';

export function LiveEvents() {
  const { events, connected } = useSSE();

  return (
    <div>
      <div className="flex items-center gap-2 mb-4">
        <h2 className="text-xl font-bold">Live Events</h2>
        <span className={`inline-block w-2 h-2 rounded-full ${connected ? 'bg-green-500' : 'bg-red-500'}`} />
        <span className="text-sm text-gray-500">{connected ? 'Connected' : 'Disconnected'}</span>
      </div>

      {events.length === 0 ? (
        <p className="text-gray-500">Waiting for events...</p>
      ) : (
        <div className="space-y-1 font-mono text-sm">
          {[...events].reverse().map(event => (
            <div key={event.id} className="bg-white border rounded px-3 py-2 flex gap-4">
              <span className="text-gray-400 w-20 shrink-0">
                {new Date(event.created_at).toLocaleTimeString()}
              </span>
              <span className="text-blue-600 w-32 shrink-0">{event.type}</span>
              <span className="text-gray-700 truncate">
                {event.message || event.name || JSON.stringify(event.data)}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
