const API_BASE = '/api/v1';

const getCsrfToken = () => {
  const match = document.cookie.match(/(?:^|;\s*)(?:csrf_|csrf_token)=([^;]+)/);
  return match ? decodeURIComponent(match[1]) : '';
};

const parseErrorMessage = async (response: Response): Promise<string> => {
  const contentType = response.headers.get('Content-Type') || '';

  let text = '';
  try {
    text = await response.text();
  } catch {
    return response.statusText || 'Request failed';
  }

  if (contentType.includes('application/json')) {
    try {
      const data = JSON.parse(text) as { error?: unknown };
      if (typeof data.error === 'string' && data.error.trim() !== '') {
        return data.error;
      }
    } catch {
      // Not valid JSON. Fallback to text/status parsing below.
    }
  }

  if (text.trim() !== '') {
    return text;
  }

  return response.statusText || 'Request failed';
};

export const apiFetch = async (url: string, options: RequestInit = {}) => {
  const headers = new Headers(options.headers || {});
  if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }
  const csrfToken = getCsrfToken();
  if (csrfToken) {
    headers.set('X-CSRF-Token', csrfToken);
  }
  const response = await fetch(`${API_BASE}${url}`, {
    credentials: 'include',
    ...options,
    headers
  });
  if (!response.ok) {
    const message = await parseErrorMessage(response);
    throw new Error(message);
  }
  if (response.status === 204) {
    return null;
  }
  return response.json();
};
