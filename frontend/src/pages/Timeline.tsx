import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import api from '../api';
import { searchPersons } from '../persons';
import { PostForm } from '../components/PostForm';
import { PostCard } from '../components/PostCard';
import { Calendar, ChevronLeft, ChevronRight, Search, Filter, X } from 'lucide-react';
import type { Post, Hashtag, PaginatedResponse, PaginationMeta } from '../types';
import { getTagColors } from '../utils/tagColors';

const PAGE_SIZE = 20;

type ViewMode = 'day' | 'search';
type Timespan = 'all' | 'last_week' | 'last_30_days' | 'this_year' | 'custom';

function localDateString(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

function getDateRange(timespan: Timespan, customStart: string, customEnd: string): { startDate?: string; endDate?: string } {
  const today = localDateString(new Date());
  if (timespan === 'last_week') {
    const d = new Date();
    d.setDate(d.getDate() - 7);
    return { startDate: localDateString(d), endDate: today };
  }
  if (timespan === 'last_30_days') {
    const d = new Date();
    d.setDate(d.getDate() - 30);
    return { startDate: localDateString(d), endDate: today };
  }
  if (timespan === 'this_year') {
    return { startDate: `${new Date().getFullYear()}-01-01`, endDate: today };
  }
  if (timespan === 'custom') {
    return {
      startDate: customStart || undefined,
      endDate: customEnd || undefined,
    };
  }
  return {};
}

export const Timeline = () => {
  const { t, i18n } = useTranslation();
  const [posts, setPosts] = useState<Post[]>([]);
  const [pagination, setPagination] = useState<PaginationMeta>({ page: 1, pageSize: PAGE_SIZE, totalItems: 0, totalPages: 0 });
  const [date, setDate] = useState(() => localDateString(new Date()));
  const [viewMode, setViewMode] = useState<ViewMode>('day');
  const [timespan, setTimespan] = useState<Timespan>('all');
  const [customStart, setCustomStart] = useState('');
  const [customEnd, setCustomEnd] = useState('');
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
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
    viewMode: ViewMode;
    date: string;
    timespan: Timespan;
    customStart: string;
    customEnd: string;
    search: string;
    selectedHashtags: string[];
    selectedPersons: string[];
  }) => {
    setLoading(true);
    try {
      const apiParams: Record<string, string | number> = {
        page: params.page,
        pageSize: PAGE_SIZE,
        search: params.search,
        hashtags: params.selectedHashtags.join(','),
        persons: params.selectedPersons.join(','),
      };

      if (params.viewMode === 'day') {
        apiParams.date = params.date;
      } else {
        const range = getDateRange(params.timespan, params.customStart, params.customEnd);
        if (range.startDate) apiParams.startDate = range.startDate;
        if (range.endDate) apiParams.endDate = range.endDate;
      }

      const response = await api.get<PaginatedResponse<Post>>('/posts', { params: apiParams });
      setPosts(response.data.items);
      setPagination(response.data.pagination);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const id = window.setTimeout(() => setDebouncedSearch(search), 300);
    return () => window.clearTimeout(id);
  }, [search]);

  useEffect(() => {
    void fetchPosts({ page, viewMode, date, timespan, customStart, customEnd, search: debouncedSearch, selectedHashtags, selectedPersons });
  }, [date, fetchPosts, page, viewMode, timespan, customStart, customEnd, debouncedSearch, selectedHashtags, selectedPersons]);

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
    const d = new Date(date + 'T12:00:00');
    d.setDate(d.getDate() + days);
    setPage(1);
    setDate(localDateString(d));
  };

  const switchMode = (mode: ViewMode) => {
    setViewMode(mode);
    setPage(1);
    if (mode === 'day') {
      setSearch('');
      setSelectedHashtags([]);
      setSelectedPersons([]);
      setShowFilters(false);
    }
  };

  const hasActiveFilters = selectedHashtags.length > 0 || selectedPersons.length > 0;

  const timespanOptions: { key: Timespan; label: string }[] = [
    { key: 'all', label: t('timespan_all') },
    { key: 'last_week', label: t('timespan_last_week') },
    { key: 'last_30_days', label: t('timespan_last_30_days') },
    { key: 'this_year', label: t('timespan_this_year') },
    { key: 'custom', label: t('timespan_custom') },
  ];

  return (
    <div>
      {/* Mode toggle */}
      <div className="mb-4 flex gap-1 p-1 bg-stone-100 rounded-lg w-fit">
        <button
          onClick={() => switchMode('day')}
          className={`inline-flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-md font-medium transition-colors ${
            viewMode === 'day'
              ? 'bg-white text-stone-800 shadow-sm'
              : 'text-stone-500 hover:text-stone-700'
          }`}
        >
          <Calendar size={14} />
          {t('view_day')}
        </button>
        <button
          onClick={() => switchMode('search')}
          className={`inline-flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-md font-medium transition-colors ${
            viewMode === 'search'
              ? 'bg-white text-stone-800 shadow-sm'
              : 'text-stone-500 hover:text-stone-700'
          }`}
        >
          <Search size={14} />
          {t('view_search')}
        </button>
      </div>

      {/* Controls */}
      <div className="mb-5 space-y-3">
        {viewMode === 'day' ? (
          /* Day mode: only date selector */
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
        ) : (
          /* Search mode: timespan + search + filter */
          <>
            {/* Timespan selector */}
            <div className="space-y-2">
              <div className="flex flex-wrap gap-1.5">
                {timespanOptions.map(({ key, label }) => (
                  <button
                    key={key}
                    onClick={() => { setTimespan(key); setPage(1); }}
                    className={`px-3 py-1.5 text-sm rounded-full border font-medium transition-colors ${
                      timespan === key
                        ? 'bg-violet-600 text-white border-violet-600'
                        : 'bg-white text-stone-600 border-stone-200 hover:bg-stone-50'
                    }`}
                  >
                    {label}
                  </button>
                ))}
              </div>
              {timespan === 'custom' && (
                <div className="flex flex-col sm:flex-row gap-2">
                  <div className="flex items-center gap-2">
                    <label className="text-xs text-stone-500 w-8 shrink-0">{t('start_date')}</label>
                    <input
                      type="date"
                      value={customStart}
                      onChange={(e) => { setCustomStart(e.target.value); setPage(1); }}
                      className="text-sm border border-stone-200 bg-white rounded-md px-3 py-1.5 shadow-sm outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-500 transition"
                    />
                  </div>
                  <div className="flex items-center gap-2">
                    <label className="text-xs text-stone-500 w-8 shrink-0">{t('end_date')}</label>
                    <input
                      type="date"
                      value={customEnd}
                      onChange={(e) => { setCustomEnd(e.target.value); setPage(1); }}
                      className="text-sm border border-stone-200 bg-white rounded-md px-3 py-1.5 shadow-sm outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-500 transition"
                    />
                  </div>
                </div>
              )}
            </div>

            {/* Search + filter toggle */}
            <div className="flex flex-col md:flex-row md:items-center gap-3">
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
                    {allHashtags.map(h => {
                      const { color, background, border } = getTagColors(h);
                      const selected = selectedHashtags.includes(h);
                      return (
                        <button
                          key={h}
                          onClick={() => {
                            setPage(1);
                            setSelectedHashtags(prev => prev.includes(h) ? prev.filter(x => x !== h) : [...prev, h]);
                          }}
                          style={selected
                            ? { background: color, color: 'white', border: 'none' }
                            : { color, background, border: `1px solid ${border}` }
                          }
                          className="px-2.5 py-1 rounded text-xs font-medium transition-colors"
                        >
                          #{h}
                        </button>
                      );
                    })}
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
                      {selectedPersons.map(p => {
                        const { color } = getTagColors(p);
                        return (
                          <button
                            key={p}
                            onClick={() => {
                              setPage(1);
                              setSelectedPersons(prev => prev.filter(x => x !== p));
                            }}
                            style={{ background: color, color: 'white' }}
                            className="inline-flex items-center gap-1 rounded px-2.5 py-1 text-xs font-medium transition-colors"
                          >
                            @{p} <X size={11} />
                          </button>
                        );
                      })}
                    </div>
                  )}
                  <div className="flex flex-wrap gap-1.5">
                    {matchingPersons
                      .filter((p) => !selectedPersons.includes(p))
                      .map(p => {
                        const { color, background, border } = getTagColors(p);
                        return (
                          <button
                            key={p}
                            onClick={() => {
                              setPage(1);
                              setSelectedPersons(prev => [...prev, p]);
                            }}
                            style={{ color, background, border: `1px solid ${border}` }}
                            className="px-2.5 py-1 rounded text-xs font-medium transition-colors"
                          >
                            @{p}
                          </button>
                        );
                      })}
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
          </>
        )}
      </div>

      <PostForm
        onSuccess={() => {
          void fetchPosts({ page, viewMode, date, timespan, customStart, customEnd, search, selectedHashtags, selectedPersons });
          setEditingPost(null);
        }}
        initialData={editingPost}
      />

      {loading ? (
        <div className="text-center py-12 text-stone-400 text-sm">{t('loading')}</div>
      ) : posts.length === 0 ? (
        <div className="text-center py-12 bg-white border border-stone-200 rounded-lg text-stone-400 text-sm">
          {viewMode === 'day' ? t('no_posts') : t('no_posts_found')}
        </div>
      ) : (
        <div className="space-y-3">
          {posts.map((post: Post) => (
            <PostCard
              key={post.id}
              post={post}
              onUpdate={() => void fetchPosts({ page, viewMode, date, timespan, customStart, customEnd, search, selectedHashtags, selectedPersons })}
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
