import { useState, useEffect, useCallback, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import api from '../api';
import { User, Plus, Trash2, Edit2, Check } from 'lucide-react';
import type { PaginatedResponse, PaginationMeta, Person } from '../types';

const PAGE_SIZE = 20;

export const Persons = () => {
  const { t } = useTranslation();
  const [persons, setPersons] = useState<Person[]>([]);
  const [pagination, setPagination] = useState<PaginationMeta>({ page: 1, pageSize: PAGE_SIZE, totalItems: 0, totalPages: 0 });
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [editingId, setEditingId] = useState<number | null>(null);
  const [page, setPage] = useState(1);

  const fetchPersons = useCallback(async (nextPage = page) => {
    const res = await api.get<PaginatedResponse<Person>>('/persons', {
      params: { page: nextPage, pageSize: PAGE_SIZE }
    });
    setPersons(res.data.items);
    setPagination(res.data.pagination);
  }, [page]);

  useEffect(() => {
    void fetchPersons();
  }, [fetchPersons]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (editingId) {
      await api.put(`/persons/${editingId}`, { name, description });
    } else {
      await api.post('/persons', { name, description });
    }
    setName('');
    setDescription('');
    setEditingId(null);
    if (page === 1) {
      await fetchPersons(1);
    } else {
      setPage(1);
    }
  };

  const handleEdit = (p: Person) => {
    setEditingId(p.id);
    setName(p.name);
    setDescription(p.description || '');
  };

  const handleDelete = async (id: number) => {
    if (window.confirm(t('delete') + '?')) {
      await api.delete(`/persons/${id}`);
      void fetchPersons();
    }
  };

  return (
    <div className="max-w-4xl mx-auto">
      <h2 className="text-2xl font-bold mb-6 flex items-center gap-2">
        <User size={24} /> {t('persons')}
      </h2>

      <div className="bg-white rounded-lg shadow p-6 mb-8">
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">{t('name')}</label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="mt-1 block w-full border rounded-md px-3 py-2"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">{t('description')}</label>
              <input
                type="text"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className="mt-1 block w-full border rounded-md px-3 py-2"
              />
            </div>
          </div>
          <div className="flex justify-end space-x-2">
            {editingId && (
              <button
                type="button"
                onClick={() => { setEditingId(null); setName(''); setDescription(''); }}
                className="bg-gray-200 px-4 py-2 rounded-md hover:bg-gray-300"
              >
                {t('cancel')}
              </button>
            )}
            <button
              type="submit"
              className="bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700 flex items-center gap-2"
            >
              {editingId ? <Check size={18} /> : <Plus size={18} />}
              {t('save')}
            </button>
          </div>
        </form>
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{t('name')}</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{t('description')}</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"></th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {persons.map((p: Person) => (
              <tr key={p.id}>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{p.name}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{p.description}</td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <button onClick={() => handleEdit(p)} className="text-indigo-600 hover:text-indigo-900 mr-4">
                    <Edit2 size={16} />
                  </button>
                  <button onClick={() => handleDelete(p.id)} className="text-red-600 hover:text-red-900">
                    <Trash2 size={16} />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {pagination.totalPages > 1 && (
        <div className="mt-4 flex items-center justify-between rounded-lg border bg-white px-4 py-3 text-sm text-gray-600">
          <span>{t('page_status', { page: pagination.page, total: pagination.totalPages })}</span>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setPage((current) => Math.max(1, current - 1))}
              disabled={pagination.page <= 1}
              className="rounded border px-3 py-1 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {t('previous')}
            </button>
            <button
              type="button"
              onClick={() => setPage((current) => Math.min(pagination.totalPages, current + 1))}
              disabled={pagination.page >= pagination.totalPages}
              className="rounded border px-3 py-1 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {t('next')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
