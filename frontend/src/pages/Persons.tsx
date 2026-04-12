import { useState, useEffect, useCallback, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import api from '../api';
import { Users, Plus, Trash2, Edit2, Check } from 'lucide-react';
import type { PaginatedResponse, PaginationMeta, Person } from '../types';
import { extractError } from '../utils/apiError';
import { ErrorAlert } from '../components/ErrorAlert';
import { useAuthStore } from '../store';

const PAGE_SIZE = 20;

export const Persons = () => {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const [persons, setPersons] = useState<Person[]>([]);
  const [pagination, setPagination] = useState<PaginationMeta>({ page: 1, pageSize: PAGE_SIZE, totalItems: 0, totalPages: 0 });
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [editingId, setEditingId] = useState<number | null>(null);
  const [page, setPage] = useState(1);
  const [error, setError] = useState<string | null>(null);

  const fetchPersons = useCallback(async (nextPage: number) => {
    const res = await api.get<PaginatedResponse<Person>>('/persons', {
      params: { page: nextPage, pageSize: PAGE_SIZE }
    });
    setPersons(res.data.items);
    setPagination(res.data.pagination);
  }, []);

  useEffect(() => {
    void fetchPersons(page);
  }, [fetchPersons, page]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    try {
      if (editingId) {
        await api.put(`/persons/${editingId}`, { name, description });
      } else {
        await api.post('/persons', { name, description });
      }
      setName('');
      setDescription('');
      setEditingId(null);
      setError(null);
      if (page === 1) {
        await fetchPersons(1);
      } else {
        setPage(1);
      }
    } catch (err) {
      setError(extractError(err, t('action_error')));
    }
  };

  const handleEdit = (p: Person) => {
    setEditingId(p.id);
    setName(p.name);
    setDescription(p.description || '');
    setError(null);
  };

  const handleDelete = async (id: number) => {
    if (window.confirm(t('delete') + '?')) {
      try {
        await api.delete(`/persons/${id}`);
        setError(null);
        void fetchPersons(page);
      } catch (err) {
        setError(extractError(err, t('delete_error')));
      }
    }
  };

  const inputClass = 'block w-full border border-stone-200 rounded-md px-3 py-2 text-sm text-stone-800 focus:outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-500 transition placeholder:text-stone-400';

  return (
    <div>
      <h2 className="text-xl font-semibold text-stone-900 mb-6 flex items-center gap-2">
        <Users size={20} className="text-stone-400" /> {t('persons')}
      </h2>

      {error && <ErrorAlert message={error} onDismiss={() => setError(null)} className="mb-4" />}

      {/* Form card */}
      <div className="bg-white rounded-lg border border-stone-200 p-5 mb-6">
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-stone-700 mb-1.5">{t('name')}</label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className={inputClass}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-stone-700 mb-1.5">{t('description')}</label>
              <input
                type="text"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className={inputClass}
              />
            </div>
          </div>
          <div className="flex justify-end gap-2">
            {editingId && (
              <button
                type="button"
                onClick={() => { setEditingId(null); setName(''); setDescription(''); }}
                className="px-4 py-2 text-sm rounded-md border border-stone-200 text-stone-600 hover:bg-stone-50 transition-colors"
              >
                {t('cancel')}
              </button>
            )}
            <button
              type="submit"
              className="inline-flex items-center gap-2 rounded-md bg-violet-600 px-4 py-2 text-sm font-medium text-white hover:bg-violet-500 transition-colors"
            >
              {editingId ? <Check size={15} /> : <Plus size={15} />}
              {t('save')}
            </button>
          </div>
        </form>
      </div>

      {/* Table */}
      <div className="bg-white rounded-lg border border-stone-200 overflow-hidden">
        <table className="min-w-full">
          <thead>
            <tr className="border-b border-stone-100 bg-stone-50">
              <th className="px-5 py-3 text-left text-xs font-medium text-stone-400 uppercase tracking-wider">{t('name')}</th>
              <th className="px-5 py-3 text-left text-xs font-medium text-stone-400 uppercase tracking-wider">{t('description')}</th>
              <th className="hidden sm:table-cell px-5 py-3 text-left text-xs font-medium text-stone-400 uppercase tracking-wider">{t('created_by')}</th>
              <th className="px-5 py-3 text-right text-xs font-medium text-stone-400 uppercase tracking-wider"></th>
            </tr>
          </thead>
          <tbody className="divide-y divide-stone-100">
            {persons.map((p: Person) => (
              <tr key={p.id} className="hover:bg-stone-50 transition-colors">
                <td className="px-5 py-3.5 text-sm font-medium text-stone-800">{p.name}</td>
                <td className="px-5 py-3.5 text-sm text-stone-500">{p.description}</td>
                <td className="hidden sm:table-cell px-5 py-3.5 text-sm text-stone-500">{p.creator?.email}</td>
                <td className="px-5 py-3.5 text-right">
                  {(user?.id === p.created_by_user_id || user?.role === 'admin') && (
                    <div className="flex justify-end gap-1">
                      <button
                        onClick={() => handleEdit(p)}
                        className="p-1.5 text-stone-400 hover:text-violet-700 hover:bg-violet-50 rounded transition-colors"
                      >
                        <Edit2 size={15} />
                      </button>
                      <button
                        onClick={() => handleDelete(p.id)}
                        className="p-1.5 text-stone-400 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
                      >
                        <Trash2 size={15} />
                      </button>
                    </div>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {pagination.totalPages > 1 && (
        <div className="mt-4 flex items-center justify-between rounded-lg border border-stone-200 bg-white px-4 py-3 text-sm text-stone-500">
          <span>{t('page_status', { page: pagination.page, total: pagination.totalPages })}</span>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setPage((current) => Math.max(1, current - 1))}
              disabled={pagination.page <= 1}
              className="rounded border border-stone-200 px-3 py-1 text-stone-600 hover:bg-stone-50 disabled:cursor-not-allowed disabled:opacity-40 transition-colors"
            >
              {t('previous')}
            </button>
            <button
              type="button"
              onClick={() => setPage((current) => Math.min(pagination.totalPages, current + 1))}
              disabled={pagination.page >= pagination.totalPages}
              className="rounded border border-stone-200 px-3 py-1 text-stone-600 hover:bg-stone-50 disabled:cursor-not-allowed disabled:opacity-40 transition-colors"
            >
              {t('next')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
