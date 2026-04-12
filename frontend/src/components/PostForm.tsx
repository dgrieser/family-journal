import { useState, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import type { AxiosError } from 'axios';
import api from '../api';
import { searchPersons } from '../persons';
import { Send, Paperclip, X } from 'lucide-react';
import type { Post, Hashtag } from '../types';
import { buildHighlightHtml } from '../utils/tagColors';

interface PostFormProps {
  onSuccess: () => void;
  initialData?: Post | null;
}

export const PostForm = ({ onSuccess, initialData }: PostFormProps) => {
  const { t, i18n } = useTranslation();
  const [text, setText] = useState(initialData?.text || '');
  const [date, setDate] = useState(initialData?.date || new Date().toISOString().split('T')[0]);
  const [files, setFiles] = useState<File[]>([]);
  const [showHashtagSuggestions, setShowHashtagSuggestions] = useState(false);
  const [showPersonSuggestions, setShowPersonSuggestions] = useState(false);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [allHashtags, setAllHashtags] = useState<string[]>([]);
  const [submitError, setSubmitError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const backdropRef = useRef<HTMLDivElement>(null);
  const personRequestIdRef = useRef(0);
  const personSearchTimeoutRef = useRef<number | null>(null);

  const cancelPendingPersonSearch = () => {
    if (personSearchTimeoutRef.current !== null) {
      window.clearTimeout(personSearchTimeoutRef.current);
      personSearchTimeoutRef.current = null;
    }
    personRequestIdRef.current += 1;
  };

  useEffect(() => {
    if (initialData) {
      setText(initialData.text);
      setDate(initialData.date.split('T')[0]);
    } else {
      setText('');
      setDate(new Date().toISOString().split('T')[0]);
    }
  }, [initialData]);

  useEffect(() => {
    // Fetch hashtags for autocomplete.
    const fetchData = async () => {
      try {
        const hRes = await api.get('/hashtags');
        setAllHashtags(hRes.data.map((h: Hashtag) => h.name));
      } catch (err) {
        console.error(err);
      }
    };
    fetchData();
  }, []);

  useEffect(() => {
    return () => {
      if (personSearchTimeoutRef.current !== null) {
        window.clearTimeout(personSearchTimeoutRef.current);
      }
    };
  }, []);

  const fetchPersonSuggestions = async (query: string) => {
    const requestId = personRequestIdRef.current + 1;
    personRequestIdRef.current = requestId;

    try {
      const persons = await searchPersons(query);
      if (personRequestIdRef.current === requestId) {
        setSuggestions(persons.map((person) => person.name));
      }
    } catch (err) {
      console.error(err);
      if (personRequestIdRef.current === requestId) {
        setSuggestions([]);
      }
    }
  };

  const handleTextChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    setText(value);

    const words = value.split(/\s/);
    const lastWord = words[words.length - 1];

    if (lastWord.startsWith('#')) {
      cancelPendingPersonSearch();
      const query = lastWord.slice(1).toLowerCase();
      setShowHashtagSuggestions(true);
      setShowPersonSuggestions(false);
      setSuggestions(allHashtags.filter(h => h.toLowerCase().includes(query)));
    } else if (lastWord.startsWith('@')) {
      const query = lastWord.slice(1).toLowerCase();
      setShowPersonSuggestions(true);
      setShowHashtagSuggestions(false);
      if (personSearchTimeoutRef.current !== null) {
        window.clearTimeout(personSearchTimeoutRef.current);
      }
      personSearchTimeoutRef.current = window.setTimeout(() => {
        void fetchPersonSuggestions(query);
      }, 300);
    } else {
      cancelPendingPersonSearch();
      setShowHashtagSuggestions(false);
      setShowPersonSuggestions(false);
      setSuggestions([]);
    }
  };

  const applySuggestion = (suggestion: string) => {
    const words = text.split(/\s/);
    const lastWord = words[words.length - 1];
    const prefix = lastWord.startsWith('#') ? '#' : '@';
    words[words.length - 1] = prefix + suggestion + ' ';
    setText(words.join(' '));
    setShowHashtagSuggestions(false);
    setShowPersonSuggestions(false);
    cancelPendingPersonSearch();
    textareaRef.current?.focus();
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      setFiles(Array.from(e.target.files));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError('');
    setIsSubmitting(true);
    let postSaved = false;

    try {
      const isUpdate = !!initialData;
      const url = isUpdate ? `/posts/${initialData.id}` : '/posts';
      const payload = { text, date };
      const config = { headers: { 'Content-Type': 'application/json' } };
      const response = await (isUpdate ? api.put(url, payload, config) : api.post(url, payload, config));
      const postId = response.data.id;
      postSaved = true;

      if (postId && files.length > 0) {
        const formData = new FormData();
        files.forEach(file => formData.append('files', file));
        await api.post(`/posts/${postId}/attachments`, formData);
      }

      setText('');
      setFiles([]);
      onSuccess();
    } catch (err) {
      console.error(err);
      const apiErr = err as AxiosError<{ error?: string }>;
      const backendMessage = apiErr.response?.data?.error;
      if (backendMessage) {
        setSubmitError(backendMessage);
      } else if (postSaved) {
        setSubmitError(t('post_partial_upload_error'));
      } else {
        setSubmitError(t('post_submit_error'));
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="bg-white rounded-lg border border-stone-200 p-4 mb-5">
      <form onSubmit={handleSubmit}>
        <div className="mb-2">
          <div className="relative inline-flex items-center gap-2 border border-stone-200 rounded px-2.5 py-1.5 cursor-pointer bg-white">
            <span className="text-sm text-stone-700 select-none whitespace-nowrap">
              {new Date(date + 'T12:00:00').toLocaleDateString(i18n.language, { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' })}
            </span>
            <input
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
              className="absolute inset-0 opacity-0 cursor-pointer w-full"
            />
          </div>
        </div>

        <div className="relative">
          {/* Backdrop that renders highlighted @mentions and #hashtags */}
          <div
            ref={backdropRef}
            aria-hidden="true"
            className="tag-highlight-backdrop"
            dangerouslySetInnerHTML={{ __html: buildHighlightHtml(text) }}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              padding: '0.75rem',
              fontSize: '0.875rem',
              fontFamily: 'inherit',
              lineHeight: 'inherit',
              whiteSpace: 'pre-wrap',
              overflowWrap: 'break-word',
              overflowY: 'auto',
              border: '1px solid transparent',
              borderRadius: '0.375rem',
              color: '#1c1917',
              pointerEvents: 'none',
              userSelect: 'none',
            }}
          />
          <textarea
            ref={textareaRef}
            value={text}
            onChange={handleTextChange}
            onScroll={(e) => {
              if (backdropRef.current) {
                backdropRef.current.scrollTop = e.currentTarget.scrollTop;
              }
            }}
            placeholder={text ? '' : t('new_post')}
            style={{ background: 'transparent', caretColor: '#57534e', color: text ? 'transparent' : undefined }}
            className="tag-textarea w-full border border-stone-200 rounded-md p-3 text-sm text-stone-800 placeholder:text-stone-400 focus:outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-500 min-h-[100px] resize-none transition relative"
          />
          {(showHashtagSuggestions || showPersonSuggestions) && suggestions.length > 0 && (
            <div className="absolute z-10 bg-white border border-stone-200 rounded-md shadow-lg mt-1 w-full max-h-40 overflow-y-auto">
              {suggestions.map(s => (
                <button
                  key={s}
                  type="button"
                  onClick={() => applySuggestion(s)}
                  className="block w-full text-left px-3.5 py-2 text-sm hover:bg-stone-50 text-stone-700 transition-colors"
                >
                  {s}
                </button>
              ))}
            </div>
          )}
        </div>

        {files.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1.5">
            {files.map((file, i) => (
              <div key={i} className="flex items-center bg-stone-100 border border-stone-200 px-2 py-1 rounded text-xs text-stone-600">
                <span className="truncate max-w-[100px]">{file.name}</span>
                <button
                  type="button"
                  onClick={() => setFiles(files.filter((_, idx) => idx !== i))}
                  className="ml-1.5 text-stone-400 hover:text-red-500 transition-colors"
                >
                  <X size={12} />
                </button>
              </div>
            ))}
          </div>
        )}

        {submitError && (
          <div className="mt-3 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
            {submitError}
          </div>
        )}

        <div className="flex items-center justify-between mt-3">
          <label className="cursor-pointer inline-flex items-center gap-1.5 text-stone-400 hover:text-stone-600 text-sm transition-colors">
            <Paperclip size={17} />
            <span>{t('upload_files')}</span>
            <input type="file" multiple onChange={handleFileChange} className="hidden" />
          </label>
          <button
            type="submit"
            disabled={!text.trim() || isSubmitting}
            className="inline-flex items-center gap-2 rounded-md bg-violet-600 px-4 py-2 text-sm font-medium text-white hover:bg-violet-500 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
          >
            <Send size={15} />
            {t('save')}
          </button>
        </div>
      </form>
    </div>
  );
};
