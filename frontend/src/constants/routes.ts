export const APP_ROUTES = {
  ROOT: '/',
  PERSONS: '/persons',
  PROFILE: '/profile',
  ADMIN: '/admin',
  AUTH_LOGIN: '/auth/login',
  AUTH_REGISTER: '/auth/register',
} as const;

export const APP_ROUTE_SEGMENTS = {
  PERSONS: 'persons',
  PROFILE: 'profile',
  ADMIN: 'admin',
} as const;

export const API_ROUTES = {
  AUTH_LOGIN: '/auth/login',
  AUTH_REGISTER: '/auth/register',
  AUTH_LOGOUT: '/auth/logout',
  AUTH_PROFILE: '/auth/profile',
} as const;
