import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import api from '../api';
import { searchPersons } from '../persons';
import { PostForm } from '../components/PostForm';
import { PostCard } from '../components/PostCard';
import { Calendar, ChevronLeft, ChevronRight, Search, Filter, X } from 'lucide-react';
import type { Post, Hashtag, PaginatedResponse, PaginationMeta } from '../types';

const PAGE_SIZE = 20;

export const Timeline = () => {
  const { t, i18n } = useTranslation();
  const [posts, setPosts] = useState<Post[]>([]);
  const [pagination, setPagination] = useState<PaginationMeta>({ page: 1, pageSize: PAGE_SIZE, totalItems: 0, totalPages: 0 });
  const [date, setDate] = useState(new Date().toISOString().split('T')[0]);
  const [search, setSearch] = useState('');
  const [selectedHashtags, setSelectedHashtags] = useState<string[]>([]);
  const [selectedPersons, setSelectedPersons] = useState<string[]>([]);
  const [allHashtags, setAllHashtags] = useState<string[]>([]);
  const [personSearch, setPersonSearch] = useState('');
  const [matchingPersons, setMatchingPersons] = useState<string[]>([]);
  const [showFilters, setShowFilters] = useState(false);
  const [editingPost, setEditingPost] = useState<Post | null>(null);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);

  const fetchPosts = useCallback(async (params: {
    page: number;
    date: string;
    search: string;
    selectedHashtags: string[];
    selectedPersons: string[];
  }) => {
    setLoading(true);
    try {
      const response = await api.get<PaginatedResponse<Post>>('/posts', {
        params: {
          page: params.page,
          pageSize: PAGE_SIZE,
          date: params.date,
          search: params.search,
          hashtags: params.selectedHashtags.join(','),
          persons: params.selectedPersons.join(',')
        }
      });
      setPosts(response.data.items);
      setPagination(response.data.pagination);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchPosts({ page, date, search, selectedHashtags, selectedPersons });
  }, [date, fetchPosts, page, search, selectedHashtags, selectedPersons]);

  useEffect(() => {
    const fetchFilters = async () => {
      try {
        const hRes = await api.get('/hashtags');
        setAllHashtags(hRes.data.map((h: Hashtag) => h.name));
      } catch (err) {
        console.error(err);
      }
    };
    fetchFilters();
  }, []);

  useEffect(() => {
    if (!showFilters) {
      return;
    }

    let cancelled = false;
    const timeoutId = window.setTimeout(() => {
      const fetchMatchingPersons = async () => {
        try {
          const persons = await searchPersons(personSearch, 12);
          if (!cancelled) {
            setMatchingPersons(persons.map((person) => person.name));
          }
        } catch (err) {
          console.error(err);
          if (!cancelled) {
            setMatchingPersons([]);
          }
        }
      };

      void fetchMatchingPersons();
    }, 300);

    return () => {
      cancelled = true;
      window.clearTimeout(timeoutId);
    };
  }, [personSearch, showFilters]);

  const changeDate = (days: number) => {
    const d = new Date(date);
    d.setDate(d.getDate() + days);
    setPage(1);
    setDate(d.toISOString().split('T')[0]);
  };

  const hasActiveFilters = selectedHashtags.length > 0 || selectedPersons.length > 0;

  return (
    <div>
      {/* Controls */}
      <div className="mb-5 space-y-3">
        <div className="flex flex-col md:flex-row md:items-center gap-3">
          {/* Date picker */}
          <div className="flex items-center gap-1">
            <button
              onClick={() => changeDate(-1)}
              className="p-1.5 text-stone-400 hover:text-stone-700 hover:bg-stone-100 rounded transition-colors"
            >
              <ChevronLeft size={18} />
            </button>
            <div className="relative flex items-center gap-2 bg-white border border-stone-200 px-3 py-2 rounded-md shadow-sm cursor-pointer">
              <Calendar size={15} className="text-stone-400 flex-shrink-0" />
              <span className="text-sm text-stone-700 select-none whitespace-nowrap">
                {new Date(date + 'T12:00:00').toLocaleDateString(i18n.language, { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' })}
              </span>
              <input
                type="date"
                value={date}
                onChange={(e) => {
                  setPage(1);
                  setDate(e.target.value);
                }}
                className="absolute inset-0 opacity-0 cursor-pointer w-full"
              />
            </div>
            <button
              onClick={() => changeDate(1)}
              className="p-1.5 text-stone-400 hover:text-stone-700 hover:bg-stone-100 rounded transition-colors"
            >
              <ChevronRight size={18} />
            </button>
          </div>

          {/* Search + filter toggle */}
          <div className="flex gap-2 flex-1">
            <button
              onClick={() => setShowFilters(!showFilters)}
              className={`inline-flex items-center gap-1.5 px-3 py-2 text-sm border rounded-md shadow-sm transition-colors ${
                showFilters || hasActiveFilters
                  ? 'bg-violet-50 border-violet-300 text-violet-700'
                  : 'bg-white border-stone-200 text-stone-500 hover:bg-stone-50'
              }`}
            >
              <Filter size={15} />
              <span className="hidden sm:inline">{t('filter')}</span>
              {hasActiveFilters && (
                <span className="bg-violet-600 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center font-medium">
                  {selectedHashtags.length + selectedPersons.length}
                </span>
              )}
            </button>
            <div className="relative flex-1">
              <Search size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-stone-400 pointer-events-none" />
              <input
                type="text"
                value={search}
                onChange={(e) => {
                  setPage(1);
                  setSearch(e.target.value);
                }}
                placeholder={t('search')}
                className="w-full pl-9 pr-4 py-2 text-sm border border-stone-200 bg-white rounded-md shadow-sm outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-500 transition placeholder:text-stone-400"
              />
            </div>
          </div>
        </div>

        {/* Filter panel */}
        {showFilters && (
          <div className="bg-white border border-stone-200 rounded-lg p-4 space-y-4 shadow-sm">
            <div>
              <h4 className="text-xs font-medium text-stone-400 uppercase tracking-wider mb-2.5">{t('hashtags')}</h4>
              <div className="flex flex-wrap gap-1.5">
                {allHashtags.map(h => (
                  <button
                    key={h}
                    onClick={() => {
                      setPage(1);
                      setSelectedHashtags(prev => prev.includes(h) ? prev.filter(x => x !== h) : [...prev, h]);
                    }}
                    className={`px-2.5 py-1 rounded text-xs font-medium transition-colors ${
                      selectedHashtags.includes(h)
                        ? 'bg-violet-600 text-white'
                        : 'bg-stone-100 text-stone-600 hover:bg-stone-200'
                    }`}
                  >
                    #{h}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <h4 className="text-xs font-medium text-stone-400 uppercase tracking-wider mb-2.5">{t('persons')}</h4>
              <input
                type="text"
                value={personSearch}
                onChange={(e) => setPersonSearch(e.target.value)}
                placeholder={t('search')}
                className="mb-3 w-full rounded-md border border-stone-200 px-3 py-2 text-sm outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-500 transition placeholder:text-stone-400"
              />
              {selectedPersons.length > 0 && (
                <div className="mb-2.5 flex flex-wrap gap-1.5">
                  {selectedPersons.map(p => (
                    <button
                      key={p}
                      onClick={() => {
                        setPage(1);
                        setSelectedPersons(prev => prev.filter(x => x !== p));
                      }}
                      className="inline-flex items-center gap-1 rounded bg-stone-700 px-2.5 py-1 text-xs text-white font-medium hover:bg-stone-600 transition-colors"
                    >
                      @{p} <X size={11} />
                    </button>
                  ))}
                </div>
              )}
              <div className="flex flex-wrap gap-1.5">
                {matchingPersons
                  .filter((p) => !selectedPersons.includes(p))
                  .map(p => (
                    <button
                      key={p}
                      onClick={() => {
                        setPage(1);
                        setSelectedPersons(prev => [...prev, p]);
                      }}
                      className="px-2.5 py-1 rounded text-xs font-medium bg-stone-100 text-stone-600 hover:bg-stone-200 transition-colors"
                    >
                      @{p}
                    </button>
                  ))}
              </div>
            </div>

            {hasActiveFilters && (
              <button
                onClick={() => {
                  setPage(1);
                  setSelectedHashtags([]);
                  setSelectedPersons([]);
                }}
                className="inline-flex items-center gap-1 text-xs text-red-500 hover:text-red-700 transition-colors font-medium"
              >
                <X size={13} /> {t('clear_filters')}
              </button>
            )}
          </div>
        )}
      </div>

      <PostForm
        onSuccess={() => {
          void fetchPosts({ page, date, search, selectedHashtags, selectedPersons });
          setEditingPost(null);
        }}
        initialData={editingPost}
      />

      {loading ? (
        <div className="text-center py-12 text-stone-400 text-sm">{t('loading')}</div>
      ) : posts.length === 0 ? (
        <div className="text-center py-12 bg-white border border-stone-200 rounded-lg text-stone-400 text-sm">
          {t('no_posts')}
        </div>
      ) : (
        <div className="space-y-3">
          {posts.map((post: Post) => (
            <PostCard
              key={post.id}
              post={post}
              onUpdate={() => void fetchPosts({ page, date, search, selectedHashtags, selectedPersons })}
              onEdit={(p) => {
                setEditingPost(p);
                window.scrollTo({ top: 0, behavior: 'smooth' });
              }}
            />
          ))}
        </div>
      )}

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
