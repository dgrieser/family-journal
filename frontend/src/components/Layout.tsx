import { Outlet, Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '../store';
import api from '../api';
import { LayoutDashboard, Users, UserCog, User, LogOut, Languages } from 'lucide-react';

export const Layout = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { user, setUser } = useAuthStore();

  const handleLogout = async () => {
    try {
      await api.post('/auth/logout');
      setUser(null);
      navigate('/auth/login');
    } catch (err) {
      console.error('Logout failed', err);
    }
  };

  const toggleLanguage = () => {
    i18n.changeLanguage(i18n.language === 'de' ? 'en' : 'de');
  };

  return (
    <div className="min-h-screen flex flex-col md:flex-row bg-gray-50">
      {/* Sidebar / Topbar */}
      <nav className="bg-indigo-700 text-white w-full md:w-64 flex-shrink-0">
        <div className="p-4 flex items-center justify-between">
          <span className="text-xl font-bold">{t('app_name')}</span>
          <button onClick={toggleLanguage} className="md:hidden p-1 hover:bg-indigo-600 rounded">
            <Languages size={20} />
          </button>
        </div>

        <div className="flex md:flex-col overflow-x-auto md:overflow-y-auto">
          <Link to="/" className="flex items-center space-x-2 p-4 hover:bg-indigo-600">
            <LayoutDashboard size={20} />
            <span>{t('timeline')}</span>
          </Link>
          <Link to="/persons" className="flex items-center space-x-2 p-4 hover:bg-indigo-600">
            <Users size={20} />
            <span>{t('persons')}</span>
          </Link>
          <Link to="/profile" className="flex items-center space-x-2 p-4 hover:bg-indigo-600">
            <User size={20} />
            <span>{t('profile')}</span>
          </Link>
          {user?.role === 'admin' && (
            <Link to="/admin" className="flex items-center space-x-2 p-4 hover:bg-indigo-600">
              <UserCog size={20} />
              <span>{t('admin')}</span>
            </Link>
          )}

          <div className="mt-auto hidden md:block">
             <button onClick={toggleLanguage} className="flex items-center space-x-2 p-4 hover:bg-indigo-600 w-full text-left">
                <Languages size={20} />
                <span>{i18n.language.toUpperCase()}</span>
             </button>
             <button onClick={handleLogout} className="flex items-center space-x-2 p-4 hover:bg-indigo-600 w-full text-left">
                <LogOut size={20} />
                <span>{t('logout')}</span>
             </button>
          </div>
        </div>
      </nav>

      <main className="flex-1 p-4 md:p-8 overflow-y-auto">
        <Outlet />
      </main>

      {/* Mobile Logout Button */}
      <div className="md:hidden bg-white border-t p-2 flex justify-around">
          <button onClick={handleLogout} className="flex flex-col items-center text-gray-600 p-2">
            <LogOut size={20} />
            <span className="text-xs">{t('logout')}</span>
          </button>
      </div>
    </div>
  );
};
