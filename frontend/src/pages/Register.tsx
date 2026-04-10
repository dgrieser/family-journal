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
    <div className="min-h-screen flex bg-stone-50">
      {/* Left branding panel */}
      <div className="hidden lg:flex flex-col justify-between w-80 xl:w-96 bg-stone-900 px-10 py-12 flex-shrink-0">
        <span
          style={{ fontFamily: 'var(--font-display)' }}
          className="text-2xl font-semibold text-white tracking-wide"
        >
          FamilyJournal
        </span>
        <p
          style={{ fontFamily: 'var(--font-display)' }}
          className="text-stone-400 text-lg italic leading-relaxed"
        >
          "Preserve the moments that matter."
        </p>
      </div>

      {/* Form panel */}
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="w-full max-w-sm">
          <div className="lg:hidden mb-8">
            <span
              style={{ fontFamily: 'var(--font-display)' }}
              className="text-2xl font-semibold text-stone-900 tracking-wide"
            >
              FamilyJournal
            </span>
          </div>

          <h1 className="text-2xl font-semibold text-stone-900 mb-1">{t('register')}</h1>
          <p className="text-sm text-stone-500 mb-7">
            {t('already_have_account')}{' '}
            <Link to={APP_ROUTES.AUTH_LOGIN} className="text-amber-700 hover:text-amber-600 font-medium transition-colors">
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
                className="block w-full rounded-md border border-stone-200 bg-white px-3.5 py-2.5 text-sm text-stone-900 shadow-sm placeholder:text-stone-400 focus:border-amber-600 focus:outline-none focus:ring-1 focus:ring-amber-600 transition"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-stone-700 mb-1.5">{t('password')}</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="block w-full rounded-md border border-stone-200 bg-white px-3.5 py-2.5 text-sm text-stone-900 shadow-sm placeholder:text-stone-400 focus:border-amber-600 focus:outline-none focus:ring-1 focus:ring-amber-600 transition"
                required
              />
            </div>
            <button
              type="submit"
              className="w-full rounded-md bg-amber-700 px-4 py-2.5 text-sm font-medium text-white hover:bg-amber-600 focus:outline-none focus:ring-2 focus:ring-amber-500 focus:ring-offset-2 transition-colors"
            >
              {t('register')}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};
