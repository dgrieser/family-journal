import type { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { APP_ROUTES } from '../constants/routes';
import { useAuthStore } from '../store';
import { FullScreenLoader } from './FullScreenLoader';

type ProtectedRouteProps = {
  children: ReactNode;
};

export const ProtectedRoute = ({ children }: ProtectedRouteProps) => {
  const { user, initialized } = useAuthStore();

  if (!initialized) {
    return <FullScreenLoader />;
  }

  if (user === null) {
    return <Navigate to={APP_ROUTES.AUTH_LOGIN} />;
  }

  return <>{children}</>;
};
