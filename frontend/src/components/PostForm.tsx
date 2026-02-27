import { useState, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import api from '../api';
import { Send, Paperclip, X } from 'lucide-react';
import type { Post, Hashtag, Person } from '../types';

interface PostFormProps {
  onSuccess: () => void;
  initialData?: Post | null;
}

export const PostForm = ({ onSuccess, initialData }: PostFormProps) => {
  const { t } = useTranslation();
  const [text, setText] = useState(initialData?.text || '');
  const [date, setDate] = useState(initialData?.date || new Date().toISOString().split('T')[0]);
  const [files, setFiles] = useState<File[]>([]);
  const [showHashtagSuggestions, setShowHashtagSuggestions] = useState(false);
  const [showPersonSuggestions, setShowPersonSuggestions] = useState(false);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [allHashtags, setAllHashtags] = useState<string[]>([]);
  const [allPersons, setAllPersons] = useState<string[]>([]);

  const textareaRef = useRef<HTMLTextAreaElement>(null);

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
    // Fetch hashtags and persons for autocomplete
    const fetchData = async () => {
       try {
         const [hRes, pRes] = await Promise.all([
           api.get('/hashtags'),
           api.get('/persons')
         ]);
         setAllHashtags(hRes.data.map((h: Hashtag) => h.name));
         setAllPersons(pRes.data.map((p: Person) => p.name));
       } catch (err) {
         console.error(err);
       }
    };
    fetchData();
  }, []);

  const handleTextChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    setText(value);

    const words = value.split(/\s/);
    const lastWord = words[words.length - 1];

    if (lastWord.startsWith('#')) {
      const query = lastWord.slice(1).toLowerCase();
      setShowHashtagSuggestions(true);
      setShowPersonSuggestions(false);
      setSuggestions(allHashtags.filter(h => h.toLowerCase().includes(query)));
    } else if (lastWord.startsWith('@')) {
      const query = lastWord.slice(1).toLowerCase();
      setShowPersonSuggestions(true);
      setShowHashtagSuggestions(false);
      setSuggestions(allPersons.filter(p => p.toLowerCase().includes(query)));
    } else {
      setShowHashtagSuggestions(false);
      setShowPersonSuggestions(false);
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
    textareaRef.current?.focus();
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      setFiles(Array.from(e.target.files));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      let postId = initialData?.id;

      if (initialData) {
        const response = await api.put(`/posts/${initialData.id}`, { text, date });
        postId = response.data.id;
      } else {
        const response = await api.post('/posts', { text, date });
        postId = response.data.id;
      }

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
    }
  };

  return (
    <div className="bg-white rounded-lg shadow p-4 mb-6">
      <form onSubmit={handleSubmit}>
        <div className="flex items-center justify-between mb-2">
           <input
             type="date"
             value={date}
             onChange={(e) => setDate(e.target.value)}
             className="border rounded px-2 py-1 text-sm"
           />
        </div>
        <div className="relative">
          <textarea
            ref={textareaRef}
            value={text}
            onChange={handleTextChange}
            placeholder={t('new_post')}
            className="w-full border rounded-lg p-3 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none min-h-[100px]"
          />
          {(showHashtagSuggestions || showPersonSuggestions) && suggestions.length > 0 && (
            <div className="absolute z-10 bg-white border rounded shadow-lg mt-1 w-full max-h-40 overflow-y-auto">
              {suggestions.map(s => (
                <button
                  key={s}
                  type="button"
                  onClick={() => applySuggestion(s)}
                  className="block w-full text-left px-4 py-2 hover:bg-gray-100"
                >
                  {s}
                </button>
              ))}
            </div>
          )}
        </div>

        <div className="mt-2 flex flex-wrap gap-2">
           {files.map((file, i) => (
             <div key={i} className="flex items-center bg-gray-100 px-2 py-1 rounded text-xs">
                <span className="truncate max-w-[100px]">{file.name}</span>
                <button type="button" onClick={() => setFiles(files.filter((_, idx) => idx !== i))} className="ml-1 text-red-500">
                  <X size={14} />
                </button>
             </div>
           ))}
        </div>

        <div className="flex items-center justify-between mt-3">
          <label className="cursor-pointer text-gray-500 hover:text-indigo-600 flex items-center gap-1">
            <Paperclip size={20} />
            <span className="text-sm">{t('upload_files')}</span>
            <input type="file" multiple onChange={handleFileChange} className="hidden" />
          </label>
          <button
            type="submit"
            disabled={!text.trim()}
            className="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 disabled:opacity-50 flex items-center gap-2"
          >
            <Send size={18} />
            {t('save')}
          </button>
        </div>
      </form>
    </div>
  );
};
