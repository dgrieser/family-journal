import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';

const RegisterPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const register = useAuthStore((state) => state.register);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError('');
    try {
      await register(email, password);
      const params = new URLSearchParams({
        registered: '1',
        email
      });
      navigate(`/login?${params.toString()}`);
    } catch (err) {
      setError(String(err));
    }
  };

  return (
    <div className="max-w-md mx-auto p-6">
      <h2 className="text-2xl font-semibold mb-4">{t('auth.register')}</h2>
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
          {t('auth.createAccount')}
        </button>
      </form>
      <p className="mt-4 text-sm">
        {t('auth.haveAccount')} <Link to="/login" className="text-slate-600">{t('auth.login')}</Link>
      </p>
    </div>
  );
};

export default RegisterPage;
