import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import api from '../api';
import { UserCog, Shield, ShieldAlert } from 'lucide-react';
import type { User } from '../types';

export const Admin = () => {
  const { t } = useTranslation();
  const [users, setUsers] = useState<User[]>([]);

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
      await api.put(`/admin/users/${userId}/role`, { role: newRole });
      fetchUsers();
    } catch (err) {
      console.error(err);
    }
  };

  const handleToggleActive = async (userId: number, isActive: boolean) => {
    try {
      await api.put(`/admin/users/${userId}/active`, { is_active: isActive });
      fetchUsers();
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <div className="max-w-4xl mx-auto">
      <h2 className="text-2xl font-bold mb-6 flex items-center gap-2">
        <UserCog size={24} /> {t('admin')}
      </h2>

      <div className="bg-white rounded-lg shadow overflow-hidden overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{t('email')}</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{t('role')}</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{t('active')}</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"></th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {users.map((u: User) => (
              <tr key={u.id}>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">{u.email}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                  <span className={`px-2 py-1 rounded text-xs font-semibold ${u.role === 'admin' ? 'bg-purple-100 text-purple-800' : 'bg-green-100 text-green-800'}`}>
                    {u.role === 'admin' ? t('admin_role') : t('user')}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                   <span className={`px-2 py-1 rounded text-xs font-semibold ${u.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                    {u.is_active ? t('active') : t('inactive')}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <div className="flex justify-end gap-3">
                    {u.role === 'admin' ? (
                      <button
                        onClick={() => handleRoleChange(u.id, 'user')}
                        className="text-red-600 hover:text-red-900 flex items-center gap-1"
                      >
                        <ShieldAlert size={16} /> {t('demote')}
                      </button>
                    ) : (
                      <button
                        onClick={() => handleRoleChange(u.id, 'admin')}
                        className="text-indigo-600 hover:text-indigo-900 flex items-center gap-1"
                      >
                        <Shield size={16} /> {t('promote')}
                      </button>
                    )}
                    <button
                      onClick={() => handleToggleActive(u.id, !u.is_active)}
                      className={`${u.is_active ? 'text-orange-600 hover:text-orange-900' : 'text-green-600 hover:text-green-900'} flex items-center gap-1`}
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
