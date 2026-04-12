import axios from 'axios';

export function extractError(err: unknown, fallback: string): string {
  if (axios.isAxiosError(err)) {
    return (err.response?.data as { error?: string })?.error || fallback;
  }
  return fallback;
}
