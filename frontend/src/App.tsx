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
  const { user, setUser, initialized, setInitialized } = useAuthStore();

  useEffect(() => {
    const checkAuth = async () => {
      try {
        const response = await api.get('/me');
        setUser(response.data);
      } catch (err) {
        setUser(null);
      } finally {
        setInitialized(true);
      }
    };
    checkAuth();
  }, [setUser, setInitialized]);

  const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
    if (!initialized) {
      return (
        <div className="min-h-screen flex items-center justify-center">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-indigo-500"></div>
        </div>
      );
    }
    if (user === null) {
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
