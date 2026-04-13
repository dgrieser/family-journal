import { useState, useEffect, useCallback, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import api from '../api';
import { Hash, Plus, Trash2, Edit2, Check } from 'lucide-react';
import type { Hashtag } from '../types';
import { extractError } from '../utils/apiError';
import { ErrorAlert } from '../components/ErrorAlert';
import { getTagColors } from '../utils/tagColors';
import { useAuthStore } from '../store';

export const Hashtags = () => {
  const { t, i18n } = useTranslation();
  const { user } = useAuthStore();
  const [hashtags, setHashtags] = useState<Hashtag[]>([]);
  const [name, setName] = useState('');
  const [editingId, setEditingId] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);

  const fetchHashtags = useCallback(async () => {
    const res = await api.get<Hashtag[]>('/hashtags');
    setHashtags(res.data);
  }, []);

  useEffect(() => {
    void fetchHashtags();
  }, [fetchHashtags]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const normalized = name.replace(/^#/, '').trim();
    if (!normalized) return;
    try {
      if (editingId) {
        await api.put(`/hashtags/${editingId}`, { name: normalized });
      } else {
        await api.post('/hashtags', { name: normalized });
      }
      setName('');
      setEditingId(null);
      setError(null);
      await fetchHashtags();
    } catch (err: unknown) {
      const raw = extractError(err, t('action_error'));
      const msg = raw.toLowerCase().includes('already exists') ? t('hashtag_exists') : raw;
      setError(msg);
    }
  };

  const handleEdit = (h: Hashtag) => {
    setEditingId(h.id);
    setName(h.name);
    setError(null);
  };

  const handleDelete = async (id: number) => {
    if (window.confirm(t('delete') + '?')) {
      try {
        await api.delete(`/hashtags/${id}`);
        setError(null);
        void fetchHashtags();
      } catch (err) {
        setError(extractError(err, t('delete_error')));
      }
    }
  };

  const inputClass =
    'block w-full border border-stone-200 rounded-md px-3 py-2 text-sm text-stone-800 focus:outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-500 transition placeholder:text-stone-400';

  return (
    <div>
      <h2 className="text-xl font-semibold text-stone-900 mb-6 flex items-center gap-2">
        <Hash size={20} className="text-stone-400" /> {t('hashtags')}
      </h2>

      {error && <ErrorAlert message={error} onDismiss={() => setError(null)} className="mb-4" />}

      {/* Form card */}
      <div className="bg-white rounded-lg border border-stone-200 p-5 mb-6">
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-stone-700 mb-1.5">
              {t('hashtag_name')}
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className={inputClass}
              placeholder="#example"
              required
            />
          </div>
          <div className="flex justify-end gap-2">
            {editingId && (
              <button
                type="button"
                onClick={() => { setEditingId(null); setName(''); }}
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
        {hashtags.length === 0 ? (
          <p className="px-5 py-8 text-center text-sm text-stone-400">{t('no_hashtags')}</p>
        ) : (
          <table className="min-w-full">
            <thead>
              <tr className="border-b border-stone-100 bg-stone-50">
                <th className="px-5 py-3 text-left text-xs font-medium text-stone-400 uppercase tracking-wider">
                  {t('name')}
                </th>
                <th className="hidden sm:table-cell px-5 py-3 text-left text-xs font-medium text-stone-400 uppercase tracking-wider">
                  {t('created_by')}
                </th>
                <th className="hidden sm:table-cell px-5 py-3 text-left text-xs font-medium text-stone-400 uppercase tracking-wider">
                  {t('created_at')}
                </th>
                <th className="px-5 py-3 text-right text-xs font-medium text-stone-400 uppercase tracking-wider"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-stone-100">
              {hashtags.map((h) => {
                const colors = getTagColors(h.name);
                return (
                  <tr key={h.id} className="hover:bg-stone-50 transition-colors">
                    <td className="px-5 py-3.5">
                      <span
                        className="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-sm font-medium"
                        style={{ color: colors.color, background: colors.background, border: `1px solid ${colors.border}` }}
                      >
                        <Hash size={11} />
                        {h.name}
                      </span>
                      <div className="sm:hidden text-xs text-stone-400 mt-1">
                        {h.creator?.email}
                      </div>
                      <div className="sm:hidden text-xs text-stone-400 mt-0.5">
                        {new Date(h.created_at).toLocaleDateString(i18n.language, {
                          year: 'numeric', month: 'short', day: 'numeric',
                        })}
                      </div>
                    </td>
                    <td className="hidden sm:table-cell px-5 py-3.5 text-sm text-stone-500">
                      {h.creator?.email}
                    </td>
                    <td className="hidden sm:table-cell px-5 py-3.5 text-sm text-stone-500">
                      {new Date(h.created_at).toLocaleDateString(i18n.language, {
                        year: 'numeric', month: 'short', day: 'numeric',
                      })}
                    </td>
                    <td className="px-5 py-3.5 text-right">
                      {(user?.id === h.created_by_user_id || user?.role === 'admin') && (
                        <div className="flex justify-end gap-1">
                          <button
                            onClick={() => handleEdit(h)}
                            className="p-1.5 text-stone-400 hover:text-violet-700 hover:bg-violet-50 rounded transition-colors"
                          >
                            <Edit2 size={15} />
                          </button>
                          <button
                            onClick={() => handleDelete(h.id)}
                            className="p-1.5 text-stone-400 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
                          >
                            <Trash2 size={15} />
                          </button>
                        </div>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
