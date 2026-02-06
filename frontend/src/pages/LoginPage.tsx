import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';

const LoginPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const login = useAuthStore((state) => state.login);
  const [email, setEmail] = useState(() => searchParams.get('email') ?? '');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const showRegisterSuccess = useMemo(() => searchParams.get('registered') === '1', [searchParams]);

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError('');
    try {
      await login(email, password);
      navigate('/');
    } catch (err) {
      setError(String(err));
    }
  };

  return (
    <div className="max-w-md mx-auto p-6">
      <h2 className="text-2xl font-semibold mb-4">{t('auth.login')}</h2>
      {showRegisterSuccess && (
        <p className="mb-4 text-sm text-emerald-700 bg-emerald-50 border border-emerald-200 rounded px-3 py-2">
          {t('auth.registerSuccess')}
        </p>
      )}
      <form onSubmit={handleSubmit} className="space-y-4">
        <input
          className="w-full border rounded px-3 py-2"
          placeholder={t('auth.email')}
          value={email}
          onChange={(event) => setEmail(event.target.value)}
        />
        <input
          className="w-full border rounded px-3 py-2"
          type="password"
          placeholder={t('auth.password')}
          value={password}
          onChange={(event) => setPassword(event.target.value)}
        />
        {error && <p className="text-sm text-red-600">{error}</p>}
        <button className="w-full bg-slate-900 text-white py-2 rounded" type="submit">
          {t('auth.signIn')}
        </button>
      </form>
      <p className="mt-4 text-sm">
        {t('auth.noAccount')} <Link to="/register" className="text-slate-600">{t('auth.register')}</Link>
      </p>
    </div>
  );
};

export default LoginPage;
