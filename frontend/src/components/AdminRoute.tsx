import type { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { APP_ROUTES } from '../constants/routes';
import { useAuthStore } from '../store';

type AdminRouteProps = {
  children: ReactNode;
};

export const AdminRoute = ({ children }: AdminRouteProps) => {
  const { user } = useAuthStore();

  if (user?.role !== 'admin') {
    return <Navigate to={APP_ROUTES.ROOT} />;
  }

  return <>{children}</>;
};
