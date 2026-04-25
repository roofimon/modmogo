'use client';

import { useState, useEffect } from 'react';

export function useAutocomplete<T>(fetchFn: (query: string) => Promise<T[]>, minLength = 1) {
  const [results, setResults] = useState<T[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const search = async (query: string) => {
    if (query.length < minLength) {
      setResults([]);
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const data = await fetchFn(query);
      setResults(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Search failed');
      setResults([]);
    } finally {
      setLoading(false);
    }
  };

  return { results, loading, error, search, setResults };
}
