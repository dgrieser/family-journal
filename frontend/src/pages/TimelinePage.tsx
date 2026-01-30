import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { apiFetch } from '../api/client';

interface Hashtag {
  id: number;
  name: string;
}

interface Person {
  id: number;
  name: string;
}

interface Attachment {
  id: number;
  file_name: string;
  url: string;
}

interface Post {
  id: number;
  text: string;
  created_at: string;
  hashtags: Hashtag[];
  persons: Person[];
  attachments: Attachment[];
}

const TimelinePage = () => {
  const { t } = useTranslation();
  const [date, setDate] = useState(() => new Date().toISOString().slice(0, 10));
  const [posts, setPosts] = useState<Post[]>([]);
  const [hashtags, setHashtags] = useState<Hashtag[]>([]);
  const [persons, setPersons] = useState<Person[]>([]);
  const [selectedHashtags, setSelectedHashtags] = useState<string[]>([]);
  const [selectedPersons, setSelectedPersons] = useState<string[]>([]);
  const [search, setSearch] = useState('');

  useEffect(() => {
    void apiFetch('/hashtags').then(setHashtags);
    void apiFetch('/persons').then(setPersons);
  }, []);

  const query = useMemo(() => {
    const params = new URLSearchParams();
    params.set('date', date);
    if (selectedHashtags.length) {
      params.set('hashtags', selectedHashtags.join(','));
    }
    if (selectedPersons.length) {
      params.set('persons', selectedPersons.join(','));
    }
    if (search) {
      params.set('search', search);
    }
    return params.toString();
  }, [date, search, selectedHashtags, selectedPersons]);

  useEffect(() => {
    void apiFetch(`/posts?${query}`).then(setPosts);
  }, [query]);

  const toggleFilter = (value: string, list: string[], setList: (next: string[]) => void) => {
    setList(list.includes(value) ? list.filter((item) => item !== value) : [...list, value]);
  };

  return (
    <div className="max-w-5xl mx-auto p-4 space-y-4">
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <h2 className="text-2xl font-semibold">{t('timeline.heading')}</h2>
          <p className="text-sm text-slate-500">{date}</p>
        </div>
        <div className="flex items-center gap-2">
          <input
            type="date"
            className="border rounded px-3 py-2"
            value={date}
            onChange={(event) => setDate(event.target.value)}
          />
          <Link to="/posts/new" className="bg-slate-900 text-white px-3 py-2 rounded text-sm">
            {t('timeline.addEntry')}
          </Link>
        </div>
      </div>
      <div className="bg-white rounded-lg shadow-sm p-4 space-y-4">
        <div>
          <label className="text-sm font-semibold">{t('timeline.search')}</label>
          <input
            className="w-full border rounded px-3 py-2 mt-2"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <p className="text-sm font-semibold">{t('timeline.hashtags')}</p>
            <div className="flex flex-wrap gap-2 mt-2">
              {hashtags.map((tag) => (
                <button
                  key={tag.id}
                  type="button"
                  className={`px-2 py-1 text-xs rounded border ${
                    selectedHashtags.includes(tag.name) ? 'bg-slate-900 text-white' : 'bg-white'
                  }`}
                  onClick={() => toggleFilter(tag.name, selectedHashtags, setSelectedHashtags)}
                >
                  #{tag.name}
                </button>
              ))}
            </div>
          </div>
          <div>
            <p className="text-sm font-semibold">{t('timeline.persons')}</p>
            <div className="flex flex-wrap gap-2 mt-2">
              {persons.map((person) => (
                <button
                  key={person.id}
                  type="button"
                  className={`px-2 py-1 text-xs rounded border ${
                    selectedPersons.includes(person.name) ? 'bg-slate-900 text-white' : 'bg-white'
                  }`}
                  onClick={() => toggleFilter(person.name, selectedPersons, setSelectedPersons)}
                >
                  @{person.name}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>
      <div className="space-y-3">
        {posts.length === 0 && (
          <div className="text-sm text-slate-500">{t('timeline.empty')}</div>
        )}
        {posts.map((post) => (
          <Link
            key={post.id}
            to={`/posts/${post.id}`}
            className="block bg-white rounded-lg shadow-sm p-4 space-y-2"
          >
            <p className="text-sm text-slate-500">{new Date(post.created_at).toLocaleTimeString()}</p>
            <p className="whitespace-pre-wrap">{post.text}</p>
            <div className="flex flex-wrap gap-2 text-xs text-slate-600">
              {post.hashtags.map((tag) => (
                <span key={tag.id}>#{tag.name}</span>
              ))}
              {post.persons.map((person) => (
                <span key={person.id}>@{person.name}</span>
              ))}
            </div>
            {post.attachments.length > 0 && (
              <div className="text-xs text-slate-500">
                {post.attachments.length} {t('post.attachments')}
              </div>
            )}
          </Link>
        ))}
      </div>
    </div>
  );
};

export default TimelinePage;
