import { useState, useEffect, useRef } from 'react';
import type { Event } from '../types';

export function useSSE(org?: string, types?: string) {
  const [events, setEvents] = useState<Event[]>([]);
  const [connected, setConnected] = useState(false);
  const esRef = useRef<EventSource | null>(null);

  useEffect(() => {
    const params = new URLSearchParams();
    if (org) params.set('org', org);
    if (types) params.set('types', types);

    const url = `/api/v1/events/stream?${params.toString()}`;
    const es = new EventSource(url);
    esRef.current = es;

    es.onopen = () => setConnected(true);
    es.onerror = () => setConnected(false);

    es.onmessage = (msg) => {
      try {
        const event: Event = JSON.parse(msg.data);
        setEvents(prev => [...prev.slice(-99), event]);
      } catch {
        // ignore parse errors
      }
    };

    return () => {
      es.close();
      esRef.current = null;
    };
  }, [org, types]);

  return { events, connected };
}
