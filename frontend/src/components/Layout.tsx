import { NavLink } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import LanguageSwitcher from './LanguageSwitcher';
import { useAuthStore } from '../stores/authStore';

interface Props {
  children: React.ReactNode;
}

const Layout = ({ children }: Props) => {
  const { t } = useTranslation();
  const { user, logout } = useAuthStore();

  return (
    <div className="min-h-screen flex flex-col">
      <header className="bg-white shadow-sm">
        <div className="max-w-5xl mx-auto flex items-center justify-between px-4 py-3">
          <h1 className="text-lg font-semibold">{t('appName')}</h1>
          <div className="flex items-center gap-3">
            <LanguageSwitcher />
            {user && (
              <button className="text-sm text-slate-600" onClick={() => void logout()}>
                {t('nav.logout')}
              </button>
            )}
          </div>
        </div>
        {user && (
          <nav className="border-t">
            <div className="max-w-5xl mx-auto flex gap-4 px-4 py-2 text-sm">
              <NavLink to="/" className="text-slate-600" end>
                {t('nav.timeline')}
              </NavLink>
              <NavLink to="/posts/new" className="text-slate-600">
                {t('nav.newPost')}
              </NavLink>
              <NavLink to="/persons" className="text-slate-600">
                {t('nav.persons')}
              </NavLink>
              <NavLink to="/profile" className="text-slate-600">
                {t('nav.profile')}
              </NavLink>
              {user.role === 'admin' && (
                <NavLink to="/admin" className="text-slate-600">
                  {t('nav.admin')}
                </NavLink>
              )}
            </div>
          </nav>
        )}
      </header>
      <main className="flex-1">{children}</main>
    </div>
  );
};

export default Layout;
