import axios from 'axios';

const api = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
});

api.interceptors.request.use((config) => {
  const csrfToken = document.cookie
    .split('; ')
    .find((row) => row.startsWith('csrf_='))
    ?.split('=')[1];

  if (csrfToken) {
    config.headers['X-Csrf-Token'] = csrfToken;
  }
  return config;
});

export default api;
