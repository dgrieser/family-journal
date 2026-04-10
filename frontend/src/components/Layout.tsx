import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '../store';
import api from '../api';
import { LayoutDashboard, Users, UserCog, User, LogOut, Languages } from 'lucide-react';
import { APP_ROUTES, API_ROUTES } from '../constants/routes';

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  `flex items-center gap-3 px-4 py-2.5 text-sm transition-colors rounded-md mx-2 ${
    isActive
      ? 'bg-amber-700 text-white font-medium'
      : 'text-stone-400 hover:bg-stone-800 hover:text-stone-100'
  }`;

export const Layout = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { user, setUser } = useAuthStore();

  const handleLogout = async () => {
    try {
      await api.post(API_ROUTES.AUTH_LOGOUT);
      setUser(null);
      navigate(APP_ROUTES.AUTH_LOGIN);
    } catch (err) {
      console.error('Logout failed', err);
    }
  };

  const toggleLanguage = () => {
    i18n.changeLanguage(i18n.language === 'de' ? 'en' : 'de');
  };

  return (
    <div className="min-h-screen flex flex-col md:flex-row bg-stone-50">
      {/* Sidebar / Topbar */}
      <nav className="bg-stone-900 text-stone-300 w-full md:w-56 flex-shrink-0 flex md:flex-col">
        {/* Brand */}
        <div className="px-5 pt-7 pb-5 md:border-b-2 md:border-amber-600 flex items-center justify-between flex-shrink-0 bg-stone-950/60">
          <div className="flex flex-col gap-1" style={{ fontFamily: 'var(--font-display)' }}>
            <span className="text-xs font-medium tracking-[0.35em] uppercase text-amber-500">
              Family
            </span>
            <span className="text-5xl font-bold text-white leading-none" style={{ letterSpacing: '-0.02em' }}>
              Journal
            </span>
          </div>
          <button
            onClick={toggleLanguage}
            className="md:hidden p-1.5 hover:text-amber-400 rounded transition-colors"
            aria-label="Toggle language"
          >
            <Languages size={18} />
          </button>
        </div>

        {/* Nav links */}
        <div className="flex md:flex-col flex-1 overflow-x-auto md:overflow-visible md:py-3 gap-0.5 items-center md:items-stretch">
          <NavLink to={APP_ROUTES.ROOT} end className={navLinkClass}>
            <LayoutDashboard size={17} />
            <span>{t('timeline')}</span>
          </NavLink>
          <NavLink to={APP_ROUTES.PERSONS} className={navLinkClass}>
            <Users size={17} />
            <span>{t('persons')}</span>
          </NavLink>
          <NavLink to={APP_ROUTES.PROFILE} className={navLinkClass}>
            <User size={17} />
            <span>{t('profile')}</span>
          </NavLink>
          {user?.role === 'admin' && (
            <NavLink to={APP_ROUTES.ADMIN} className={navLinkClass}>
              <UserCog size={17} />
              <span>{t('admin')}</span>
            </NavLink>
          )}
        </div>

        {/* Bottom items - desktop */}
        <div className="hidden md:flex flex-col border-t border-stone-800 py-2 gap-0.5">
          <button
            onClick={toggleLanguage}
            className="flex items-center gap-3 px-4 py-2.5 text-sm text-stone-400 hover:bg-stone-800 hover:text-stone-100 transition-colors rounded-md mx-2"
          >
            <Languages size={17} />
            <span>{i18n.language.toUpperCase()}</span>
          </button>
          <button
            onClick={handleLogout}
            className="flex items-center gap-3 px-4 py-2.5 text-sm text-stone-400 hover:bg-stone-800 hover:text-stone-100 transition-colors rounded-md mx-2"
          >
            <LogOut size={17} />
            <span>{t('logout')}</span>
          </button>
        </div>
      </nav>

      <main className="flex-1 p-4 md:p-8 overflow-y-auto min-h-0">
        <Outlet />
      </main>

      {/* Mobile logout bar */}
      <div className="md:hidden bg-stone-900 text-stone-400 border-t border-stone-800 p-2 flex justify-around">
        <button
          onClick={handleLogout}
          className="flex flex-col items-center gap-0.5 p-2 hover:text-stone-100 transition-colors"
        >
          <LogOut size={18} />
          <span className="text-xs">{t('logout')}</span>
        </button>
      </div>
    </div>
  );
};
