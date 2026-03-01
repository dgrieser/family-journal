import { useEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { Timeline } from './pages/Timeline';
import { Persons } from './pages/Persons';
import { Admin } from './pages/Admin';
import { Profile } from './pages/Profile';
import { Layout } from './components/Layout';
import { ProtectedRoute } from './components/ProtectedRoute';
import { AdminRoute } from './components/AdminRoute';
import { useAuthStore } from './store';
import api from './api';
import './i18n';
import { APP_ROUTES, APP_ROUTE_SEGMENTS, API_ROUTES } from './constants/routes';

function App() {
  const { setUser } = useAuthStore();

  useEffect(() => {
    const checkAuth = async () => {
      try {
        const response = await api.get(API_ROUTES.AUTH_PROFILE);
        setUser(response.data);
      } catch (err) {
        console.error('Auth check failed:', err);
        setUser(null);
      }
    };
    checkAuth();
  }, [setUser]);

  return (
    <BrowserRouter>
      <Routes>
        <Route path={APP_ROUTES.AUTH_LOGIN} element={<Login />} />
        <Route path={APP_ROUTES.AUTH_REGISTER} element={<Register />} />

        <Route path={APP_ROUTES.ROOT} element={<ProtectedRoute><Layout /></ProtectedRoute>}>
          <Route index element={<Timeline />} />
          <Route path={APP_ROUTE_SEGMENTS.PERSONS} element={<Persons />} />
          <Route path={APP_ROUTE_SEGMENTS.PROFILE} element={<Profile />} />
          <Route path={APP_ROUTE_SEGMENTS.ADMIN} element={<ProtectedRoute><AdminRoute><Admin /></AdminRoute></ProtectedRoute>} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
