import { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { Timeline } from './pages/Timeline';
import { Persons } from './pages/Persons';
import { Admin } from './pages/Admin';
import { Profile } from './pages/Profile';
import { Layout } from './components/Layout';
import { useAuthStore } from './store';
import api from './api';
import './i18n';

function App() {
  const { user, setUser } = useAuthStore();

  useEffect(() => {
    const checkAuth = async () => {
      try {
        const response = await api.get('/me');
        setUser(response.data);
      } catch (err) {
        setUser(null);
      }
    };
    checkAuth();
  }, [setUser]);

  const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
    if (user === null) {
      // We might want a loading state here while checkAuth is running
      return <Navigate to="/login" />;
    }
    return <>{children}</>;
  };

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />

        <Route path="/" element={<ProtectedRoute><Layout /></ProtectedRoute>}>
          <Route index element={<Timeline />} />
          <Route path="persons" element={<Persons />} />
          <Route path="profile" element={<Profile />} />
          <Route path="admin" element={user?.role === 'admin' ? <Admin /> : <Navigate to="/" />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
