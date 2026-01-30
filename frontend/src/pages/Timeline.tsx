import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import api from '../api';
import { PostForm } from '../components/PostForm';
import { PostCard } from '../components/PostCard';
import { Calendar, ChevronLeft, ChevronRight, Search, Filter, X } from 'lucide-react';

export const Timeline = () => {
  const { t } = useTranslation();
  const [posts, setPosts] = useState([]);
  const [date, setDate] = useState(new Date().toISOString().split('T')[0]);
  const [search, setSearch] = useState('');
  const [selectedHashtags, setSelectedHashtags] = useState<string[]>([]);
  const [selectedPersons, setSelectedPersons] = useState<string[]>([]);
  const [allHashtags, setAllHashtags] = useState<string[]>([]);
  const [allPersons, setAllPersons] = useState<string[]>([]);
  const [showFilters, setShowFilters] = useState(false);
  const [editingPost, setEditingPost] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  const fetchPosts = useCallback(async () => {
    setLoading(true);
    try {
      const response = await api.get('/posts', {
        params: {
          date,
          search,
          hashtags: selectedHashtags.join(','),
          persons: selectedPersons.join(',')
        }
      });
      setPosts(response.data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [date, search, selectedHashtags, selectedPersons]);

  useEffect(() => {
    fetchPosts();
  }, [fetchPosts]);

  useEffect(() => {
    const fetchFilters = async () => {
      try {
        const [hRes, pRes] = await Promise.all([
          api.get('/hashtags'),
          api.get('/persons')
        ]);
        setAllHashtags(hRes.data.map((h: any) => h.name));
        setAllPersons(pRes.data.map((p: any) => p.name));
      } catch (err) {
        console.error(err);
      }
    };
    fetchFilters();
  }, []);

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

          <div className="relative flex-1 flex gap-2">
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
            <button
              onClick={() => setShowFilters(!showFilters)}
              className={`p-2 border rounded-md shadow-sm hover:bg-gray-50 flex items-center gap-1 ${showFilters ? 'bg-indigo-50 border-indigo-200 text-indigo-600' : 'bg-white'}`}
            >
              <Filter size={20} />
              <span className="hidden sm:inline">{t('hashtags')} / {t('mentions')}</span>
            </button>
          </div>
        </div>

        {showFilters && (
          <div className="bg-white p-4 rounded-lg shadow-sm border space-y-4">
            <div>
              <h4 className="text-sm font-semibold text-gray-700 mb-2">{t('hashtags')}</h4>
              <div className="flex flex-wrap gap-2">
                {allHashtags.map(h => (
                  <button
                    key={h}
                    onClick={() => setSelectedHashtags(prev => prev.includes(h) ? prev.filter(x => x !== h) : [...prev, h])}
                    className={`px-3 py-1 rounded-full text-xs transition ${selectedHashtags.includes(h) ? 'bg-indigo-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}`}
                  >
                    #{h}
                  </button>
                ))}
              </div>
            </div>
            <div>
              <h4 className="text-sm font-semibold text-gray-700 mb-2">{t('persons')}</h4>
              <div className="flex flex-wrap gap-2">
                {allPersons.map(p => (
                  <button
                    key={p}
                    onClick={() => setSelectedPersons(prev => prev.includes(p) ? prev.filter(x => x !== p) : [...prev, p])}
                    className={`px-3 py-1 rounded-full text-xs transition ${selectedPersons.includes(p) ? 'bg-green-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}`}
                  >
                    @{p}
                  </button>
                ))}
              </div>
            </div>
            {(selectedHashtags.length > 0 || selectedPersons.length > 0) && (
              <button
                onClick={() => { setSelectedHashtags([]); setSelectedPersons([]); }}
                className="text-xs text-red-500 flex items-center gap-1 hover:underline"
              >
                <X size={14} /> {t('clear_filters')}
              </button>
            )}
          </div>
        )}
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
