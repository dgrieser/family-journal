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

  return (
    <div className="max-w-md mx-auto">
      <h2 className="text-2xl font-bold mb-6 flex items-center gap-2">
        <User size={24} /> {t('profile')}
      </h2>

      <div className="bg-white rounded-lg shadow p-6">
        {message.text && (
          <div className={`mb-4 p-3 rounded-md flex items-center gap-2 ${message.type === 'success' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
            {message.type === 'success' ? <CheckCircle size={18} /> : <AlertCircle size={18} />}
            {message.text}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700">{t('email')}</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="mt-1 block w-full border rounded-md px-3 py-2"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700">{t('current_password')}</label>
            <input
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              className="mt-1 block w-full border rounded-md px-3 py-2"
              placeholder="••••••••"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700">New password ({t('leave_blank')})</label>
            <input
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className="mt-1 block w-full border rounded-md px-3 py-2"
              placeholder="••••••••"
            />
          </div>
          <button
            type="submit"
            className="w-full bg-indigo-600 text-white py-2 rounded-md hover:bg-indigo-700 flex items-center justify-center gap-2"
          >
            <Save size={18} />
            {t('update')}
          </button>
        </form>
      </div>
    </div>
  );
};
