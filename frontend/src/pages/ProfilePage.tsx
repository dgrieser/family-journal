import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { apiFetch } from '../api/client';

const ProfilePage = () => {
  const { t } = useTranslation();
  const [email, setEmail] = useState('');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [profileStatus, setProfileStatus] = useState('');
  const [passwordStatus, setPasswordStatus] = useState('');
  const [profileError, setProfileError] = useState('');
  const [passwordError, setPasswordError] = useState('');

  useEffect(() => {
    void apiFetch('/auth/profile').then((user) => setEmail(user.email));
  }, []);

  const handleUpdateProfile = async (event: React.FormEvent) => {
    event.preventDefault();
    setProfileError('');
    setProfileStatus('');
    try {
      await apiFetch('/auth/profile', {
        method: 'PUT',
        body: JSON.stringify({ email })
      });
      setProfileStatus(t('profile.updateSuccess'));
    } catch (err) {
      setProfileError(err instanceof Error ? err.message : String(err));
    }
  };

  const handleChangePassword = async (event: React.FormEvent) => {
    event.preventDefault();
    setPasswordError('');
    setPasswordStatus('');
    try {
      await apiFetch('/auth/profile', {
        method: 'PUT',
        body: JSON.stringify({
          currentPassword,
          newPassword
        })
      });
      setCurrentPassword('');
      setNewPassword('');
      setPasswordStatus(t('profile.passwordSuccess'));
    } catch (err) {
      setPasswordError(err instanceof Error ? err.message : String(err));
    }
  };

  return (
    <div className="max-w-md mx-auto p-4 space-y-8">
      <h2 className="text-2xl font-semibold">{t('profile.title')}</h2>

      <form onSubmit={handleUpdateProfile} className="space-y-3">
        <input
          className="w-full border rounded px-3 py-2"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
        />
        <button className="bg-slate-900 text-white px-4 py-2 rounded" type="submit">
          {t('profile.update')}
        </button>
      </form>
      {profileError && <p className="text-sm text-red-600">{profileError}</p>}
      {profileStatus && <p className="text-sm text-slate-500">{profileStatus}</p>}

      <section className="space-y-3">
        <h3 className="text-xl font-semibold">{t('profile.passwordTitle')}</h3>
        <form onSubmit={handleChangePassword} className="space-y-3">
          <input
            className="w-full border rounded px-3 py-2"
            type="password"
            placeholder={t('profile.currentPassword')}
            value={currentPassword}
            onChange={(event) => setCurrentPassword(event.target.value)}
          />
          <input
            className="w-full border rounded px-3 py-2"
            type="password"
            placeholder={t('profile.newPassword')}
            value={newPassword}
            onChange={(event) => setNewPassword(event.target.value)}
          />
          <button className="bg-slate-900 text-white px-4 py-2 rounded" type="submit">
            {t('profile.changePassword')}
          </button>
        </form>
        {passwordError && <p className="text-sm text-red-600">{passwordError}</p>}
        {passwordStatus && <p className="text-sm text-slate-500">{passwordStatus}</p>}
      </section>
    </div>
  );
};

export default ProfilePage;
