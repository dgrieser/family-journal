import type { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { APP_ROUTES } from '../constants/routes';
import { useAuthStore } from '../store';
import { FullScreenLoader } from './FullScreenLoader';

type AdminRouteProps = {
  children: ReactNode;
};

export const AdminRoute = ({ children }: AdminRouteProps) => {
  const { user, initialized } = useAuthStore();

  if (!initialized) {
    return <FullScreenLoader />;
  }

  if (user === null) {
    return <Navigate to={APP_ROUTES.AUTH_LOGIN} />;
  }

  if (user.role !== 'admin') {
    return <Navigate to={APP_ROUTES.ROOT} />;
  }

  return <>{children}</>;
};
