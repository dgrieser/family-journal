import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { apiFetch } from '../api/client';

const ProfilePage = () => {
  const { t } = useTranslation();
  const [email, setEmail] = useState('');
  const [status, setStatus] = useState('');

  useEffect(() => {
    void apiFetch('/auth/profile').then((user) => setEmail(user.email));
  }, []);

  const handleUpdate = async (event: React.FormEvent) => {
    event.preventDefault();
    await apiFetch('/auth/profile', {
      method: 'PUT',
      body: JSON.stringify({ email })
    });
    setStatus(t('profile.update'));
  };

  return (
    <div className="max-w-md mx-auto p-4 space-y-4">
      <h2 className="text-2xl font-semibold">{t('profile.title')}</h2>
      <form onSubmit={handleUpdate} className="space-y-3">
        <input
          className="w-full border rounded px-3 py-2"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
        />
        <button className="bg-slate-900 text-white px-4 py-2 rounded" type="submit">
          {t('profile.update')}
        </button>
      </form>
      {status && <p className="text-sm text-slate-500">{status}</p>}
    </div>
  );
};

export default ProfilePage;
