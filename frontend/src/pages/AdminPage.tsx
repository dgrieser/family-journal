import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { apiFetch } from '../api/client';

interface User {
  id: number;
  email: string;
  role: string;
  active: boolean;
}

const AdminPage = () => {
  const { t } = useTranslation();
  const [users, setUsers] = useState<User[]>([]);

  const loadUsers = async () => {
    const data = await apiFetch('/admin/users');
    setUsers(data);
  };

  useEffect(() => {
    void loadUsers();
  }, []);

  const updateRole = async (id: number, role: string) => {
    await apiFetch(`/admin/users/${id}/role`, {
      method: 'PATCH',
      body: JSON.stringify({ role })
    });
    await loadUsers();
  };

  const updateActive = async (id: number, active: boolean) => {
    await apiFetch(`/admin/users/${id}/active`, {
      method: 'PATCH',
      body: JSON.stringify({ active })
    });
    await loadUsers();
  };

  return (
    <div className="max-w-4xl mx-auto p-4 space-y-4">
      <h2 className="text-2xl font-semibold">{t('admin.title')}</h2>
      <div className="space-y-3">
        {users.map((user) => (
          <div key={user.id} className="bg-white p-4 rounded shadow-sm flex flex-col md:flex-row md:items-center md:justify-between gap-2">
            <div>
              <p className="font-semibold">{user.email}</p>
              <p className="text-sm text-slate-500">{t('admin.role')}: {user.role}</p>
            </div>
            <div className="flex gap-2">
              <button
                className="text-sm border px-3 py-1 rounded"
                onClick={() => void updateRole(user.id, user.role === 'admin' ? 'user' : 'admin')}
              >
                {t('admin.role')}
              </button>
              <button
                className="text-sm border px-3 py-1 rounded"
                onClick={() => void updateActive(user.id, !user.active)}
              >
                {user.active ? t('admin.deactivate') : t('admin.activate')}
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default AdminPage;
