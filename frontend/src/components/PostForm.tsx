import { useState, useEffect, useRef, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import type { AxiosError } from 'axios';
import api from '../api';
import { searchPersons } from '../persons';
import { Send, Paperclip, X, Image, Clock } from 'lucide-react';
import type { Post, Hashtag, Attachment } from '../types';
import { buildHighlightHtml } from '../utils/tagColors';
import { getCaretCoordinates } from '../utils/caretCoordinates';

interface CaretPos {
  top: number;
  height: number;
}

interface PostFormProps {
  onSuccess: () => void;
  onCancel?: () => void;
  initialData?: Post | null;
  embedded?: boolean;
}

const localDateStr = (d: Date) =>
  `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;

export const PostForm = ({ onSuccess, onCancel, initialData, embedded }: PostFormProps) => {
  const { t, i18n } = useTranslation();
  const [text, setText] = useState(initialData?.text || '');
  // For existing posts: slice the stored wall-clock datetime string directly so
  // no timezone conversion occurs. For new posts: use local date/time components.
  const [date, setDate] = useState(initialData?.date.split('T')[0] ?? localDateStr(new Date()));
  const [time, setTime] = useState(initialData ? initialData.date.slice(11, 16) : new Date().toTimeString().slice(0, 5));
  const [files, setFiles] = useState<File[]>([]);
  const [pendingDeleteIds, setPendingDeleteIds] = useState<number[]>([]);
  const [showHashtagSuggestions, setShowHashtagSuggestions] = useState(false);
  const [showPersonSuggestions, setShowPersonSuggestions] = useState(false);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [allHashtags, setAllHashtags] = useState<string[]>([]);
  const [submitError, setSubmitError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [caretPos, setCaretPos] = useState<CaretPos | null>(null);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [isMobile, setIsMobile] = useState(
    typeof window !== 'undefined' ? window.matchMedia('(max-width: 767px)').matches : false,
  );

  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const backdropRef = useRef<HTMLDivElement>(null);
  const dateInputRef = useRef<HTMLInputElement>(null);
  const timeInputRef = useRef<HTMLInputElement>(null);
  const personRequestIdRef = useRef(0);
  const personSearchTimeoutRef = useRef<number | null>(null);
  const suggestionItemRefs = useRef<Array<HTMLButtonElement | null>>([]);

  const suggestionsOpen =
    (showHashtagSuggestions || showPersonSuggestions) && suggestions.length > 0;

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
      setTime(initialData.date.slice(11, 16));
    } else {
      setText('');
      setDate(localDateStr(new Date()));
      setTime(new Date().toTimeString().slice(0, 5));
    }
    setFiles([]);
    setPendingDeleteIds([]);
  }, [initialData]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const hRes = await api.get('/hashtags');
        setAllHashtags(hRes.data.map((h: Hashtag) => h.name));
      } catch (err) {
        console.error(err);
      }
    };
    void fetchData();
  }, []);

  useEffect(() => {
    return () => {
      if (personSearchTimeoutRef.current !== null) {
        window.clearTimeout(personSearchTimeoutRef.current);
      }
    };
  }, []);

  // Sync backdrop scroll after each text change (covers browser auto-scroll when
  // the cursor moves below the visible area, which does not fire onScroll on iOS).
  useEffect(() => {
    const id = requestAnimationFrame(() => {
      if (textareaRef.current && backdropRef.current) {
        backdropRef.current.scrollTop = textareaRef.current.scrollTop;
      }
    });
    return () => cancelAnimationFrame(id);
  }, [text]);

  useEffect(() => {
    const mql = window.matchMedia('(max-width: 767px)');
    const handler = (e: MediaQueryListEvent) => setIsMobile(e.matches);
    setIsMobile(mql.matches);
    mql.addEventListener('change', handler);
    return () => mql.removeEventListener('change', handler);
  }, []);

  useEffect(() => {
    setSelectedIndex(0);
  }, [suggestions]);

  const updateCaretPosition = useCallback(() => {
    const el = textareaRef.current;
    if (!el) return;
    const coords = getCaretCoordinates(el, el.selectionStart);
    setCaretPos({ top: coords.top - el.scrollTop, height: coords.height });
  }, []);

  useEffect(() => {
    if (!suggestionsOpen) return;
    const onResize = () => updateCaretPosition();
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  }, [suggestionsOpen, updateCaretPosition]);

  useEffect(() => {
    if (!suggestionsOpen) return;
    suggestionItemRefs.current[selectedIndex]?.scrollIntoView({ block: 'nearest' });
  }, [selectedIndex, suggestionsOpen]);

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

  // Returns the whitespace-delimited token containing the caret, or null if
  // the caret is on whitespace.
  const findWordAtCursor = (value: string, cursor: number) => {
    if (cursor > 0 && /\s/.test(value[cursor - 1]) && (cursor >= value.length || /\s/.test(value[cursor]))) {
      return null;
    }
    let start = cursor;
    while (start > 0 && !/\s/.test(value[start - 1])) start--;
    let end = cursor;
    while (end < value.length && !/\s/.test(value[end])) end++;
    return { start, end, word: value.slice(start, end) };
  };

  const handleTextChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    setText(value);

    const cursor = e.target.selectionStart;
    const token = findWordAtCursor(value, cursor);
    const word = token?.word ?? '';

    if (word.startsWith('#')) {
      cancelPendingPersonSearch();
      const query = word.slice(1);
      const lowerQuery = query.toLowerCase();
      setShowHashtagSuggestions(true);
      setShowPersonSuggestions(false);
      setSuggestions(allHashtags.filter(h => h.toLowerCase().includes(lowerQuery)));
      requestAnimationFrame(updateCaretPosition);
    } else if (word.startsWith('@')) {
      const query = word.slice(1).toLowerCase();
      setShowPersonSuggestions(true);
      setShowHashtagSuggestions(false);
      if (personSearchTimeoutRef.current !== null) {
        window.clearTimeout(personSearchTimeoutRef.current);
      }
      personSearchTimeoutRef.current = window.setTimeout(() => {
        void fetchPersonSuggestions(query);
      }, 300);
      requestAnimationFrame(updateCaretPosition);
    } else {
      cancelPendingPersonSearch();
      setShowHashtagSuggestions(false);
      setShowPersonSuggestions(false);
      setSuggestions([]);
      setCaretPos(null);
    }
  };

  const handleTextareaKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (!suggestionsOpen) return;
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex((i) => (i + 1) % suggestions.length);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex((i) => (i - 1 + suggestions.length) % suggestions.length);
    } else if (e.key === 'Enter' || e.key === 'Tab') {
      e.preventDefault();
      const idx = Math.min(selectedIndex, suggestions.length - 1);
      applySuggestion(suggestions[idx]);
    } else if (e.key === 'Escape') {
      e.preventDefault();
      setShowHashtagSuggestions(false);
      setShowPersonSuggestions(false);
      setCaretPos(null);
    }
  };

  const applySuggestion = (suggestion: string) => {
    const el = textareaRef.current;
    if (!el) return;
    const cursor = el.selectionStart;
    const token = findWordAtCursor(text, cursor);
    if (!token) return;
    const { start, end, word } = token;

    let replacement: string;
    if (word.startsWith('#')) {
      const typedPart = word.slice(1);
      const completed = suggestion.toLowerCase().startsWith(typedPart.toLowerCase())
        ? typedPart + suggestion.slice(typedPart.length)
        : suggestion;
      replacement = '#' + completed;
    } else {
      replacement = '@' + suggestion;
    }

    // Append a trailing space only if the following character isn't already whitespace.
    const needsSpace = end >= text.length || !/\s/.test(text[end]);
    const insert = needsSpace ? replacement + ' ' : replacement;

    const newText = text.slice(0, start) + insert + text.slice(end);
    const newCursor = start + insert.length;

    setText(newText);
    setShowHashtagSuggestions(false);
    setShowPersonSuggestions(false);
    setCaretPos(null);
    cancelPendingPersonSearch();

    requestAnimationFrame(() => {
      el.focus();
      el.setSelectionRange(newCursor, newCursor);
    });
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      const newFiles = Array.from(e.target.files);
      setFiles(prev => [...prev, ...newFiles]);
      e.target.value = '';
    }
  };

  const togglePendingDelete = (id: number) => {
    setPendingDeleteIds(prev =>
      prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError('');
    setIsSubmitting(true);
    let postSaved = false;
    let attachmentError = false;

    try {
      const isUpdate = !!initialData;
      const url = isUpdate ? `/posts/${initialData.id}` : '/posts';
      const payload = { text, date, time };
      const config = { headers: { 'Content-Type': 'application/json' } };
      const response = await (isUpdate ? api.put(url, payload, config) : api.post(url, payload, config));
      const postId = response.data.id;
      postSaved = true;

      await Promise.all(pendingDeleteIds.map(async (id) => {
        try {
          await api.delete(`/attachments/${id}`);
        } catch (err) {
          console.error('Failed to delete attachment', id, err);
          attachmentError = true;
        }
      }));

      if (postId && files.length > 0) {
        const formData = new FormData();
        files.forEach(file => formData.append('files', file));
        await api.post(`/posts/${postId}/attachments`, formData);
      }

      if (!attachmentError) {
        setText('');
        setFiles([]);
        setPendingDeleteIds([]);
        onSuccess();
      }
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
      if (attachmentError) {
        setSubmitError(t('attachment_delete_error'));
      }
      setIsSubmitting(false);
    }
  };

  const existingAttachments = initialData?.attachments ?? [];

  const formContent = (
    <form onSubmit={handleSubmit}>
      <div className="mb-2 flex items-center gap-2 flex-wrap">
        <div className="relative inline-flex items-center gap-2 border border-stone-200 rounded px-2.5 py-1.5 cursor-pointer bg-white" onClick={() => dateInputRef.current?.showPicker()}>
          <span className="text-sm text-stone-700 select-none whitespace-nowrap">
            {new Date(date + 'T12:00:00').toLocaleDateString(i18n.language, { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' })}
          </span>
          <input
            ref={dateInputRef}
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            className="absolute inset-0 opacity-0 pointer-events-none w-full"
          />
        </div>
        <div className="relative inline-flex items-center gap-1.5 border border-stone-200 rounded px-2.5 py-1.5 cursor-pointer bg-white" onClick={() => timeInputRef.current?.showPicker()}>
          <Clock size={14} className="text-stone-400 flex-shrink-0" />
          <span className="text-sm text-stone-700 select-none whitespace-nowrap">
            {time ? new Date(`1970-01-01T${time}`).toLocaleTimeString(i18n.language, { hour: '2-digit', minute: '2-digit' }) : '--:--'}
          </span>
          <input
            ref={timeInputRef}
            type="time"
            value={time}
            onChange={(e) => setTime(e.target.value)}
            className="absolute inset-0 opacity-0 pointer-events-none w-full"
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
            lineHeight: '1.25rem',
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
          onKeyDown={handleTextareaKeyDown}
          onScroll={(e) => {
            if (backdropRef.current) {
              backdropRef.current.scrollTop = e.currentTarget.scrollTop;
            }
            if (suggestionsOpen) {
              requestAnimationFrame(updateCaretPosition);
            }
          }}
          placeholder={text ? '' : t('new_post')}
          style={{ background: 'transparent', caretColor: '#57534e', color: text ? 'transparent' : undefined }}
          className="tag-textarea w-full border border-stone-200 rounded-md p-3 text-sm text-stone-800 placeholder:text-stone-400 focus:outline-none focus:border-violet-500 focus:ring-inset focus:ring-1 focus:ring-violet-500 min-h-[100px] resize-none transition relative"
        />
        {suggestionsOpen && (
          <div
            className="absolute z-10 left-0 right-0 bg-white border border-stone-200 rounded-md shadow-lg max-h-40 overflow-y-auto"
            style={
              caretPos
                ? isMobile
                  ? {
                      top: caretPos.top,
                      transform: 'translateY(-100%) translateY(-4px)',
                    }
                  : {
                      top: caretPos.top + caretPos.height + 4,
                    }
                : { top: '100%', marginTop: 4 }
            }
          >
            {suggestions.map((s, i) => (
              <button
                key={s}
                ref={(el) => {
                  suggestionItemRefs.current[i] = el;
                }}
                type="button"
                onMouseEnter={() => setSelectedIndex(i)}
                onMouseDown={(e) => e.preventDefault()}
                onClick={() => applySuggestion(s)}
                className={`block w-full text-left px-3.5 py-2 text-sm text-stone-700 transition-colors ${
                  i === selectedIndex ? 'bg-stone-100' : 'hover:bg-stone-50'
                }`}
              >
                {s}
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Existing attachments (edit mode only) */}
      {existingAttachments.length > 0 && (
        <div className="mt-3 border border-stone-200 rounded-md p-2.5 space-y-1.5">
          <p className="text-xs font-medium text-stone-400 uppercase tracking-wider mb-1.5">{t('attachments')}</p>
          <div className="flex flex-wrap gap-2">
            {existingAttachments.map((a: Attachment) => {
              const markedForDeletion = pendingDeleteIds.includes(a.id);
              return (
                <div
                  key={a.id}
                  className={`flex items-center gap-1.5 border rounded px-2 py-1 text-xs transition-all ${
                    markedForDeletion
                      ? 'bg-red-50 border-red-200 text-red-400 opacity-60'
                      : 'bg-stone-50 border-stone-200 text-stone-600'
                  }`}
                >
                  {a.file_type.startsWith('image/') ? (
                    <Image size={11} className="flex-shrink-0" />
                  ) : (
                    <Paperclip size={11} className="flex-shrink-0" />
                  )}
                  <span className={`truncate max-w-[120px] ${markedForDeletion ? 'line-through' : ''}`}>
                    {a.file_name}
                  </span>
                  <button
                    type="button"
                    onClick={() => togglePendingDelete(a.id)}
                    className={`ml-0.5 transition-colors ${
                      markedForDeletion
                        ? 'text-red-400 hover:text-stone-500'
                        : 'text-stone-400 hover:text-red-500'
                    }`}
                    title={markedForDeletion ? t('cancel') : t('delete')}
                  >
                    <X size={11} />
                  </button>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* New files to upload */}
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
        <div className="flex items-center gap-2">
          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              className="inline-flex items-center rounded-md border border-stone-200 px-4 py-2 text-sm font-medium text-stone-600 hover:bg-stone-50 transition-colors"
            >
              {t('cancel')}
            </button>
          )}
          <button
            type="submit"
            disabled={!text.trim() || isSubmitting}
            className="inline-flex items-center gap-2 rounded-md bg-violet-600 px-4 py-2 text-sm font-medium text-white hover:bg-violet-500 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
          >
            <Send size={15} />
            {t('save')}
          </button>
        </div>
      </div>
    </form>
  );

  if (embedded) {
    return formContent;
  }

  return (
    <div className="bg-white rounded-lg border border-stone-200 p-4 mb-5">
      {formContent}
    </div>
  );
};
