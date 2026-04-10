import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import axios from 'axios';
import api from '../api';
import { useAuthStore } from '../store';
import { User, Save, CheckCircle, AlertCircle } from 'lucide-react';
import { API_ROUTES } from '../constants/routes';

export const Profile = () => {
  const { t } = useTranslation();
  const { user, setUser } = useAuthStore();
  const [email, setEmail] = useState(user?.email || '');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [message, setMessage] = useState({ type: '', text: '' });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await api.put(API_ROUTES.AUTH_PROFILE, {
        email,
        currentPassword,
        newPassword,
      });
      setUser(res.data);
      setMessage({ type: 'success', text: t('success') });
      setCurrentPassword('');
      setNewPassword('');
    } catch (err: unknown) {
      let errorMsg = t('error');
      if (axios.isAxiosError(err)) {
        errorMsg = err.response?.data?.error || err.response?.data?.message || t('error');
      }
      setMessage({ type: 'error', text: errorMsg });
    }
  };

  const inputClass = 'block w-full border border-stone-200 rounded-md px-3.5 py-2.5 text-sm text-stone-800 placeholder:text-stone-400 focus:outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-500 transition';

  return (
    <div className="max-w-lg">
      <h2 className="text-xl font-semibold text-stone-900 mb-6 flex items-center gap-2">
        <User size={20} className="text-stone-400" /> {t('profile')}
      </h2>

      <div className="bg-white rounded-lg border border-stone-200 p-6">
        {message.text && (
          <div className={`mb-5 flex items-center gap-2 rounded-md px-4 py-3 text-sm border ${
            message.type === 'success'
              ? 'bg-green-50 border-green-200 text-green-800'
              : 'bg-red-50 border-red-200 text-red-700'
          }`}>
            {message.type === 'success' ? <CheckCircle size={16} /> : <AlertCircle size={16} />}
            {message.text}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-stone-700 mb-1.5">{t('email')}</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className={inputClass}
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-stone-700 mb-1.5">{t('current_password')}</label>
            <input
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              className={inputClass}
              placeholder="••••••••"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-stone-700 mb-1.5">
              {t('new_password')} <span className="text-stone-400 font-normal">({t('leave_blank')})</span>
            </label>
            <input
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className={inputClass}
              placeholder="••••••••"
            />
          </div>
          <button
            type="submit"
            className="w-full inline-flex items-center justify-center gap-2 rounded-md bg-violet-600 px-4 py-2.5 text-sm font-medium text-white hover:bg-violet-500 transition-colors"
          >
            <Save size={15} />
            {t('update')}
          </button>
        </form>
      </div>
    </div>
  );
};
