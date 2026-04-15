import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import api from '../api';
import { UserCog, Shield, ShieldAlert, Power } from 'lucide-react';
import type { User } from '../types';
import { extractError } from '../utils/apiError';
import { ErrorAlert } from '../components/ErrorAlert';

export const Admin = () => {
  const { t } = useTranslation();
  const [users, setUsers] = useState<User[]>([]);
  const [error, setError] = useState<string | null>(null);

  const fetchUsers = async () => {
    try {
      const res = await api.get('/admin/users');
      setUsers(res.data);
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  const handleRoleChange = async (userId: number, newRole: string) => {
    try {
      const res = await api.patch(`/admin/users/${userId}/role`, { role: newRole });
      setUsers(users.map(u => u.id === userId ? res.data : u));
    } catch (err) {
      setError(extractError(err, t('action_error')));
    }
  };

  const handleToggleActive = async (userId: number, isActive: boolean) => {
    try {
      const res = await api.patch(`/admin/users/${userId}/active`, { is_active: isActive });
      setUsers(users.map(u => u.id === userId ? res.data : u));
    } catch (err) {
      setError(extractError(err, t('action_error')));
    }
  };

  return (
    <div>
      <h2 className="text-xl font-semibold text-stone-900 mb-6 flex items-center gap-2">
        <UserCog size={20} className="text-stone-400" /> {t('admin')}
      </h2>

      {error && <ErrorAlert message={error} onDismiss={() => setError(null)} className="mb-4" />}

      <div className="bg-white rounded-lg border border-stone-200 overflow-hidden overflow-x-auto">
        <table className="min-w-full">
          <thead>
            <tr className="border-b border-stone-100 bg-stone-50">
              <th className="px-5 py-3 text-left text-xs font-medium text-stone-400 uppercase tracking-wider">{t('email')}</th>
              <th className="px-5 py-3 text-left text-xs font-medium text-stone-400 uppercase tracking-wider hidden md:table-cell">{t('role')}</th>
              <th className="px-5 py-3 text-left text-xs font-medium text-stone-400 uppercase tracking-wider hidden md:table-cell">{t('active')}</th>
              <th className="px-5 py-3 text-right text-xs font-medium text-stone-400 uppercase tracking-wider"></th>
            </tr>
          </thead>
          <tbody className="divide-y divide-stone-100">
            {users.map((u: User) => (
              <tr key={u.id} className="hover:bg-stone-50 transition-colors">
                <td className="px-5 py-3.5 text-sm text-stone-800">
                  {u.email}
                  {/* Mobile-only: show role + active badges below email */}
                  <div className="md:hidden flex flex-wrap gap-1.5 mt-1">
                    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                      u.role === 'admin' ? 'bg-stone-800 text-stone-100' : 'bg-stone-100 text-stone-600'
                    }`}>
                      {u.role === 'admin' ? t('admin_role') : t('user')}
                    </span>
                    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                      u.is_active
                        ? 'bg-green-50 text-green-700 border border-green-100'
                        : 'bg-red-50 text-red-600 border border-red-100'
                    }`}>
                      {u.is_active ? t('active') : t('inactive')}
                    </span>
                  </div>
                </td>
                <td className="px-5 py-3.5 hidden md:table-cell">
                  <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                    u.role === 'admin'
                      ? 'bg-stone-800 text-stone-100'
                      : 'bg-stone-100 text-stone-600'
                  }`}>
                    {u.role === 'admin' ? t('admin_role') : t('user')}
                  </span>
                </td>
                <td className="px-5 py-3.5 hidden md:table-cell">
                  <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                    u.is_active
                      ? 'bg-green-50 text-green-700 border border-green-100'
                      : 'bg-red-50 text-red-600 border border-red-100'
                  }`}>
                    {u.is_active ? t('active') : t('inactive')}
                  </span>
                </td>
                <td className="px-5 py-3.5 text-right">
                  {/* Mobile: icon-only buttons */}
                  <div className="flex justify-end gap-2 md:hidden">
                    {u.role === 'admin' ? (
                      <button
                        onClick={() => handleRoleChange(u.id, 'user')}
                        title={t('demote')}
                        className="p-1.5 text-red-500 hover:text-red-700 transition-colors"
                      >
                        <ShieldAlert size={16} />
                      </button>
                    ) : (
                      <button
                        onClick={() => handleRoleChange(u.id, 'admin')}
                        title={t('promote')}
                        className="p-1.5 text-violet-700 hover:text-violet-600 transition-colors"
                      >
                        <Shield size={16} />
                      </button>
                    )}
                    <button
                      onClick={() => handleToggleActive(u.id, !u.is_active)}
                      title={u.is_active ? t('deactivate') : t('activate')}
                      className={`p-1.5 transition-colors ${
                        u.is_active ? 'text-stone-500 hover:text-stone-700' : 'text-green-600 hover:text-green-700'
                      }`}
                    >
                      <Power size={16} />
                    </button>
                  </div>
                  {/* Desktop: text + icon buttons */}
                  <div className="hidden md:flex justify-end gap-3">
                    {u.role === 'admin' ? (
                      <button
                        onClick={() => handleRoleChange(u.id, 'user')}
                        className="inline-flex items-center gap-1 text-xs text-red-500 hover:text-red-700 font-medium transition-colors"
                      >
                        <ShieldAlert size={14} /> {t('demote')}
                      </button>
                    ) : (
                      <button
                        onClick={() => handleRoleChange(u.id, 'admin')}
                        className="inline-flex items-center gap-1 text-xs text-violet-700 hover:text-violet-600 font-medium transition-colors"
                      >
                        <Shield size={14} /> {t('promote')}
                      </button>
                    )}
                    <button
                      onClick={() => handleToggleActive(u.id, !u.is_active)}
                      className={`inline-flex items-center gap-1 text-xs font-medium transition-colors ${
                        u.is_active
                          ? 'text-stone-500 hover:text-stone-700'
                          : 'text-green-600 hover:text-green-700'
                      }`}
                    >
                      {u.is_active ? t('deactivate') : t('activate')}
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
