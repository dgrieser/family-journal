import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import api from '../api';
import { PostForm } from '../components/PostForm';
import { PostCard } from '../components/PostCard';
import { Calendar, ChevronLeft, ChevronRight, Search } from 'lucide-react';

export const Timeline = () => {
  const { t } = useTranslation();
  const [posts, setPosts] = useState([]);
  const [date, setDate] = useState(new Date().toISOString().split('T')[0]);
  const [search, setSearch] = useState('');
  const [editingPost, setEditingPost] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  const fetchPosts = useCallback(async () => {
    setLoading(true);
    try {
      const response = await api.get('/posts', {
        params: { date, search }
      });
      setPosts(response.data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [date, search]);

  useEffect(() => {
    fetchPosts();
  }, [fetchPosts]);

  const changeDate = (days: number) => {
    const d = new Date(date);
    d.setDate(d.getDate() + days);
    setDate(d.toISOString().split('T')[0]);
  };

  return (
    <div className="max-w-3xl mx-auto">
      <div className="mb-6 space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="flex items-center space-x-2">
            <button onClick={() => changeDate(-1)} className="p-2 hover:bg-gray-200 rounded">
              <ChevronLeft size={20} />
            </button>
            <div className="flex items-center space-x-2 bg-white border px-3 py-1.5 rounded-md shadow-sm">
              <Calendar size={18} className="text-gray-400" />
              <input
                type="date"
                value={date}
                onChange={(e) => setDate(e.target.value)}
                className="outline-none text-sm font-medium"
              />
            </div>
            <button onClick={() => changeDate(1)} className="p-2 hover:bg-gray-200 rounded">
              <ChevronRight size={20} />
            </button>
          </div>

          <div className="relative flex-1">
            <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t('search')}
              className="w-full pl-10 pr-4 py-2 border rounded-md shadow-sm outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
        </div>
      </div>

      <PostForm
        onSuccess={() => {
          fetchPosts();
          setEditingPost(null);
        }}
        initialData={editingPost}
      />

      {loading ? (
        <div className="text-center py-10 text-gray-500">{t('loading')}</div>
      ) : posts.length === 0 ? (
        <div className="text-center py-10 bg-white rounded-lg shadow text-gray-500">
          {t('no_posts')}
        </div>
      ) : (
        <div className="space-y-4">
          {posts.map((post: any) => (
            <PostCard
              key={post.id}
              post={post}
              onUpdate={fetchPosts}
              onEdit={(p) => {
                setEditingPost({ id: p.id, text: p.text, date: p.date.split('T')[0] });
                window.scrollTo({ top: 0, behavior: 'smooth' });
              }}
            />
          ))}
        </div>
      )}
    </div>
  );
};
