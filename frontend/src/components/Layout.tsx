import { useState } from 'react';
import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '../store';
import api from '../api';
import { LayoutDashboard, Users, UserCog, User, LogOut, Languages, Menu, X } from 'lucide-react';
import { APP_ROUTES, API_ROUTES } from '../constants/routes';

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  `flex items-center gap-3 px-4 py-2.5 text-sm transition-colors rounded-md mx-2 ${
    isActive
      ? 'bg-violet-600 text-white font-medium'
      : 'text-slate-400 hover:bg-slate-700 hover:text-slate-100'
  }`;

export const Layout = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { user, setUser } = useAuthStore();
  const [drawerOpen, setDrawerOpen] = useState(false);

  const closeDrawer = () => setDrawerOpen(false);

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
    <div className="min-h-screen flex flex-col md:flex-row bg-slate-50">

      {/* Mobile top bar */}
      <div className="md:hidden flex items-center justify-between bg-slate-800 px-5 py-4 flex-shrink-0">
        <div className="flex flex-col gap-0" style={{ fontFamily: 'var(--font-display)' }}>
          <span className="text-xs font-medium tracking-[0.35em] uppercase text-violet-400 ml-[8px] mb-[-2px]">
            Family
          </span>
          <span className="text-5xl font-bold text-white leading-none" style={{ letterSpacing: '-0.02em' }}>
            Journal
          </span>
        </div>
        <button
          onClick={() => setDrawerOpen(true)}
          className="p-2 text-slate-300 hover:text-violet-400 rounded transition-colors"
          aria-label="Open menu"
        >
          <Menu size={24} />
        </button>
      </div>

      {/* Mobile drawer (backdrop + panel) — always in DOM for CSS transitions */}
      <div className="md:hidden">
        {/* Backdrop */}
        <div
          onClick={closeDrawer}
          className={`fixed inset-0 z-40 bg-black/50 transition-opacity duration-300 ${
            drawerOpen ? 'opacity-100 pointer-events-auto' : 'opacity-0 pointer-events-none'
          }`}
          aria-hidden="true"
        />

        {/* Drawer panel */}
        <div
          className={`fixed inset-y-0 left-0 z-50 w-72 bg-slate-800 text-slate-300 flex flex-col shadow-2xl transition-transform duration-300 ease-in-out ${
            drawerOpen ? 'translate-x-0' : '-translate-x-full'
          }`}
        >
          {/* Drawer header */}
          <div className="px-5 pt-7 pb-5 border-b-2 border-violet-600 flex items-center justify-between flex-shrink-0 bg-slate-900/60">
            <div className="flex flex-col gap-0" style={{ fontFamily: 'var(--font-display)' }}>
              <span className="text-xs font-medium tracking-[0.35em] uppercase text-violet-400 ml-[8px] mb-[-2px]">
                Family
              </span>
              <span className="text-5xl font-bold text-white leading-none" style={{ letterSpacing: '-0.02em' }}>
                Journal
              </span>
            </div>
            <button
              onClick={closeDrawer}
              className="p-1.5 text-slate-400 hover:text-violet-400 rounded transition-colors"
              aria-label="Close menu"
            >
              <X size={22} />
            </button>
          </div>

          {/* Drawer nav links */}
          <div className="flex flex-col py-3 gap-0.5 flex-1 overflow-y-auto">
            <NavLink to={APP_ROUTES.ROOT} end className={navLinkClass} onClick={closeDrawer}>
              <LayoutDashboard size={17} />
              <span>{t('timeline')}</span>
            </NavLink>
            <NavLink to={APP_ROUTES.PERSONS} className={navLinkClass} onClick={closeDrawer}>
              <Users size={17} />
              <span>{t('persons')}</span>
            </NavLink>
            <NavLink to={APP_ROUTES.PROFILE} className={navLinkClass} onClick={closeDrawer}>
              <User size={17} />
              <span>{t('profile')}</span>
            </NavLink>
            {user?.role === 'admin' && (
              <NavLink to={APP_ROUTES.ADMIN} className={navLinkClass} onClick={closeDrawer}>
                <UserCog size={17} />
                <span>{t('admin')}</span>
              </NavLink>
            )}
          </div>

          {/* Drawer bottom actions */}
          <div className="flex flex-col border-t border-slate-700 py-2 gap-0.5">
            <button
              onClick={toggleLanguage}
              className="flex items-center gap-3 px-4 py-2.5 text-sm text-slate-400 hover:bg-slate-700 hover:text-slate-100 transition-colors rounded-md mx-2"
            >
              <Languages size={17} />
              <span>{i18n.language.toUpperCase()}</span>
            </button>
            <button
              onClick={handleLogout}
              className="flex items-center gap-3 px-4 py-2.5 text-sm text-slate-400 hover:bg-slate-700 hover:text-slate-100 transition-colors rounded-md mx-2"
            >
              <LogOut size={17} />
              <span>{t('logout')}</span>
            </button>
          </div>
        </div>
      </div>

      {/* Desktop sidebar — hidden on mobile */}
      <nav className="hidden md:flex bg-slate-800 text-slate-300 md:w-[var(--sidebar-width)] flex-shrink-0 flex-col">
        {/* Brand */}
        <div className="px-5 pt-7 pb-5 border-b-2 border-violet-600 flex items-center justify-between flex-shrink-0 bg-slate-900/60">
          <div className="flex flex-col gap-0" style={{ fontFamily: 'var(--font-display)' }}>
            <span className="text-xs font-medium tracking-[0.35em] uppercase text-violet-400 ml-[8px] mb-[-2px]">
              Family
            </span>
            <span className="text-5xl font-bold text-white leading-none" style={{ letterSpacing: '-0.02em' }}>
              Journal
            </span>
          </div>
        </div>

        {/* Nav links */}
        <div className="flex flex-col flex-1 py-3 gap-0.5">
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

        {/* Bottom items */}
        <div className="flex flex-col border-t border-slate-700 py-2 gap-0.5">
          <button
            onClick={toggleLanguage}
            className="flex items-center gap-3 px-4 py-2.5 text-sm text-slate-400 hover:bg-slate-700 hover:text-slate-100 transition-colors rounded-md mx-2"
          >
            <Languages size={17} />
            <span>{i18n.language.toUpperCase()}</span>
          </button>
          <button
            onClick={handleLogout}
            className="flex items-center gap-3 px-4 py-2.5 text-sm text-slate-400 hover:bg-slate-700 hover:text-slate-100 transition-colors rounded-md mx-2"
          >
            <LogOut size={17} />
            <span>{t('logout')}</span>
          </button>
        </div>
      </nav>

      <main className="flex-1 p-4 md:p-8 overflow-y-auto min-h-0">
        <Outlet />
      </main>
    </div>
  );
};
