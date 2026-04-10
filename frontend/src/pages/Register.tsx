import { useState, type FormEvent } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import axios from 'axios';
import api from '../api';
import { APP_ROUTES, API_ROUTES } from '../constants/routes';

export const Register = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();
  const { t } = useTranslation();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    try {
      await api.post(API_ROUTES.AUTH_REGISTER, { email, password });
      navigate(APP_ROUTES.AUTH_LOGIN, { state: { registrationSuccess: true } });
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        setError(err.response?.data?.error || 'Registration failed');
      } else {
        setError('An unexpected error occurred');
      }
    }
  };

  return (
    <div className="min-h-screen flex bg-slate-50">
      {/* Left branding panel */}
      <div className="hidden lg:flex flex-col justify-between lg:w-[var(--sidebar-width)] bg-slate-800 flex-shrink-0">
        <div className="px-5 pt-7 pb-5 bg-slate-900/60 border-b-2 border-violet-600">
          <div className="flex flex-col gap-0" style={{ fontFamily: 'var(--font-display)' }}>
            <span className="text-xs font-medium tracking-[0.35em] uppercase text-violet-400 ml-[8px] mb-[-2px]">
              Family
            </span>
            <span className="text-5xl font-bold text-white leading-none" style={{ letterSpacing: '-0.02em' }}>
              Journal
            </span>
          </div>
        </div>
      </div>

      {/* Form panel */}
      <div className="flex-1 flex items-start sm:items-center justify-center px-6 pt-10 pb-8 sm:p-8">
        <div className="w-full max-w-sm">
          <div className="lg:hidden mb-6 flex flex-col gap-0" style={{ fontFamily: 'var(--font-display)' }}>
            <span className="text-xs font-medium tracking-[0.35em] uppercase text-violet-500 ml-[5px] mb-[-1px]">
              Family
            </span>
            <span className="text-3xl font-bold text-stone-900 leading-none" style={{ letterSpacing: '-0.02em' }}>
              Journal
            </span>
          </div>

          <h1 className="text-2xl font-semibold text-stone-900 mb-1">{t('register')}</h1>
          <p className="text-sm text-stone-500 mb-7">
            {t('already_have_account')}{' '}
            <Link to={APP_ROUTES.AUTH_LOGIN} className="text-violet-700 hover:text-violet-600 font-medium transition-colors">
              {t('login')}
            </Link>
          </p>

          {error && (
            <div className="mb-5 rounded-md bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-stone-700 mb-1.5">{t('email')}</label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="block w-full rounded-md border border-stone-200 bg-white px-3.5 py-2.5 text-sm text-stone-900 shadow-sm placeholder:text-stone-400 focus:border-violet-600 focus:outline-none focus:ring-1 focus:ring-violet-600 transition"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-stone-700 mb-1.5">{t('password')}</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="block w-full rounded-md border border-stone-200 bg-white px-3.5 py-2.5 text-sm text-stone-900 shadow-sm placeholder:text-stone-400 focus:border-violet-600 focus:outline-none focus:ring-1 focus:ring-violet-600 transition"
                required
              />
            </div>
            <button
              type="submit"
              className="w-full rounded-md bg-violet-600 px-4 py-2.5 text-sm font-medium text-white hover:bg-violet-500 focus:outline-none focus:ring-2 focus:ring-violet-500 focus:ring-offset-2 transition-colors"
            >
              {t('register')}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};
