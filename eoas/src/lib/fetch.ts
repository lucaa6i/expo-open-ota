import fetchRetry from 'fetch-retry';
import originalFetch, { RequestInit, Response } from 'node-fetch';

import Log from './log';
const fetch = fetchRetry(originalFetch);

export async function fetchWithRetries(url: string, options: RequestInit): Promise<Response> {
  return await fetch(url, {
    ...options,
    retryDelay(attempt) {
      return Math.pow(2, attempt) * 500;
    },
    retryOn: (attempt, error) => {
      if (attempt > 3) {
        return false;
      }
      if (error) {
        Log.warn(`Retry ${attempt} after network error:`, error.message);
        return true;
      }
      return false;
    },
  });
}
