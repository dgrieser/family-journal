import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router-dom';
import { apiFetch } from '../api/client';

interface Suggestion {
  id: number;
  name: string;
}

const PostEditorPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { id } = useParams();
  const [date, setDate] = useState(() => new Date().toISOString().slice(0, 10));
  const [text, setText] = useState('');
  const [category, setCategory] = useState('');
  const [mood, setMood] = useState('');
  const [hashtags, setHashtags] = useState<Suggestion[]>([]);
  const [persons, setPersons] = useState<Suggestion[]>([]);
  const [files, setFiles] = useState<File[]>([]);
  const [error, setError] = useState('');

  useEffect(() => {
    void apiFetch('/hashtags').then(setHashtags);
    void apiFetch('/persons').then(setPersons);
  }, []);

  useEffect(() => {
    if (!id) return;
    void apiFetch(`/posts/${id}`).then((post) => {
      setDate(post.date.slice(0, 10));
      setText(post.text);
      setCategory(post.category || '');
      setMood(post.mood || '');
    });
  }, [id]);

  const token = useMemo(() => {
    const last = text.split(/\s+/).pop() || '';
    return last.startsWith('#') || last.startsWith('@') ? last : '';
  }, [text]);

  const suggestions = useMemo(() => {
    if (token.startsWith('#')) {
      return hashtags.filter((tag) => tag.name.startsWith(token.slice(1)));
    }
    if (token.startsWith('@')) {
      return persons.filter((person) => person.name.startsWith(token.slice(1)));
    }
    return [];
  }, [token, hashtags, persons]);

  const applySuggestion = (value: string) => {
    const parts = text.split(/\s+/);
    parts[parts.length - 1] = value;
    setText(parts.join(' ') + ' ');
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError('');
    try {
      const payload = {
        date,
        text,
        category: category || null,
        mood: mood || null
      };
      const post = id
        ? await apiFetch(`/posts/${id}`, { method: 'PUT', body: JSON.stringify(payload) })
        : await apiFetch('/posts', { method: 'POST', body: JSON.stringify(payload) });
      if (files.length > 0) {
        const formData = new FormData();
        files.forEach((file) => formData.append('files', file));
        await apiFetch(`/posts/${post.id}/attachments`, { method: 'POST', body: formData });
      }
      navigate(`/posts/${post.id}`);
    } catch (err) {
      setError(String(err));
    }
  };

  return (
    <div className="max-w-2xl mx-auto p-4">
      <h2 className="text-2xl font-semibold mb-4">{t('post.title')}</h2>
      <form onSubmit={handleSubmit} className="space-y-4">
        <input
          type="date"
          className="border rounded px-3 py-2"
          value={date}
          onChange={(event) => setDate(event.target.value)}
        />
        <div>
          <textarea
            className="w-full border rounded px-3 py-2 min-h-[150px]"
            placeholder={t('post.text')}
            value={text}
            onChange={(event) => setText(event.target.value)}
          />
          {suggestions.length > 0 && (
            <div className="border rounded mt-1 bg-white shadow-sm">
              {suggestions.map((suggestion) => (
                <button
                  key={suggestion.id}
                  type="button"
                  className="block w-full text-left px-3 py-2 hover:bg-slate-100"
                  onClick={() => applySuggestion(`${token.startsWith('#') ? '#' : '@'}${suggestion.name}`)}
                >
                  {token.startsWith('#') ? '#' : '@'}{suggestion.name}
                </button>
              ))}
            </div>
          )}
        </div>
        <input
          className="w-full border rounded px-3 py-2"
          placeholder={t('post.category')}
          value={category}
          onChange={(event) => setCategory(event.target.value)}
        />
        <input
          className="w-full border rounded px-3 py-2"
          placeholder={t('post.mood')}
          value={mood}
          onChange={(event) => setMood(event.target.value)}
        />
        <div>
          <label className="text-sm font-semibold">{t('post.attachments')}</label>
          <input
            className="mt-2"
            type="file"
            multiple
            onChange={(event) => setFiles(Array.from(event.target.files || []))}
          />
          {files.length > 0 && (
            <ul className="text-xs text-slate-500 mt-2">
              {files.map((file) => (
                <li key={file.name}>{file.name}</li>
              ))}
            </ul>
          )}
        </div>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <button className="bg-slate-900 text-white px-4 py-2 rounded" type="submit">
          {id ? t('post.update') : t('post.save')}
        </button>
      </form>
    </div>
  );
};

export default PostEditorPage;
