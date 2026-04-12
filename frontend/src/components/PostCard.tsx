import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import api from '../api';
import { MessageSquare, Trash2, Edit2, Download, User as UserIcon, Tag, Send, Paperclip } from 'lucide-react';
import { useAuthStore } from '../store';
import type { Post, Hashtag, Person, Attachment, Comment } from '../types';
import { getTagColors, TAG_PATTERN } from '../utils/tagColors';
import { extractError } from '../utils/apiError';
import { ErrorAlert } from './ErrorAlert';

function renderTextWithTags(text: string) {
  const parts = text.split(TAG_PATTERN);
  return parts.map((part, i) => {
    if (/^[@#][\p{L}\d_]+$/u.test(part)) {
      const name = part.slice(1);
      const { color, background, border } = getTagColors(name);
      return (
        <span
          key={i}
          style={{ color, background, border: `1px solid ${border}` }}
          className="inline rounded px-1 py-0.5 text-xs font-medium mx-0.5 whitespace-nowrap"
        >
          {part}
        </span>
      );
    }
    return part;
  });
}

interface PostCardProps {
  post: Post;
  onUpdate: () => void;
  onEdit: (post: Post) => void;
}

export const PostCard = ({ post, onUpdate, onEdit }: PostCardProps) => {
  const { t, i18n } = useTranslation();
  const { user } = useAuthStore();
  const [commentText, setCommentText] = useState('');
  const [showComments, setShowComments] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleDelete = async () => {
    if (window.confirm(t('delete') + '?')) {
      try {
        await api.delete(`/posts/${post.id}`);
        setError(null);
        onUpdate();
      } catch (err) {
        setError(extractError(err, t('delete_error')));
      }
    }
  };

  const handleAddComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!commentText.trim()) return;
    try {
      await api.post(`/posts/${post.id}/comments`, { text: commentText });
      setCommentText('');
      setError(null);
      onUpdate();
    } catch (err) {
      setError(extractError(err, t('comment_error')));
    }
  };

  const handleDeleteComment = async (commentId: number) => {
    try {
      await api.delete(`/comments/${commentId}`);
      setError(null);
      onUpdate();
    } catch (err) {
      setError(extractError(err, t('delete_error')));
    }
  };

  return (
    <div className="bg-white rounded-lg border border-slate-200 p-5 mb-3">
      {/* Header */}
      <div className="flex justify-between items-start mb-3">
        <div>
          <span className="text-xs text-stone-900 block">{new Date(post.created_at).toLocaleString(i18n.language, { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit' })}</span>
          <span className="text-xs text-stone-400 block mt-0.5">{post.user?.email}</span>
        </div>
        {(user?.id === post.user_id || user?.role === 'admin') && (
          <div className="flex gap-1">
            <button
              onClick={() => onEdit(post)}
              className="p-1.5 text-stone-400 hover:text-stone-600 hover:bg-stone-100 rounded transition-colors"
            >
              <Edit2 size={15} />
            </button>
            <button
              onClick={handleDelete}
              className="p-1.5 text-stone-400 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
            >
              <Trash2 size={15} />
            </button>
          </div>
        )}
      </div>

      {/* Error */}
      {error && <ErrorAlert message={error} onDismiss={() => setError(null)} className="mb-3" />}

      {/* Content */}
      <p className="text-stone-700 whitespace-pre-wrap mb-4 leading-relaxed text-sm">{renderTextWithTags(post.text)}</p>

      {/* Tags */}
      {(post.hashtags?.length > 0 || post.persons?.length > 0) && (
        <div className="flex flex-wrap gap-1.5 mb-4">
          {post.hashtags?.map((h: Hashtag) => {
            const { color, background, border } = getTagColors(h.name);
            return (
              <span key={h.id} style={{ color, background, border: `1px solid ${border}` }} className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium">
                <Tag size={11} /> {h.name}
              </span>
            );
          })}
          {post.persons?.map((p: Person) => {
            const { color, background, border } = getTagColors(p.name);
            return (
              <span key={p.id} style={{ color, background, border: `1px solid ${border}` }} className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium">
                <UserIcon size={11} /> {p.name}
              </span>
            );
          })}
        </div>
      )}

      {/* Attachments */}
      {post.attachments?.length > 0 && (
        <div className="border-t border-slate-100 pt-4 mb-4">
          <h4 className="text-xs font-medium text-stone-400 uppercase tracking-wider mb-2.5">{t('attachments')}</h4>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
            {post.attachments.map((a: Attachment) => (
              <div key={a.id} className="space-y-1">
                {a.file_type.startsWith('image/') ? (
                  <img
                    src={`${api.defaults.baseURL}/attachments/${a.id}/download`}
                    alt={a.file_name}
                    className="w-full h-28 object-cover rounded border border-slate-200"
                  />
                ) : (
                  <div className="w-full h-28 bg-slate-50 rounded border border-slate-200 flex items-center justify-center text-slate-300">
                    <Paperclip size={22} />
                  </div>
                )}
                <a
                  href={`${api.defaults.baseURL}/attachments/${a.id}/download`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center justify-between p-2 border border-slate-200 rounded hover:bg-slate-50 text-xs text-slate-600 transition-colors"
                >
                  <span className="truncate">{a.file_name}</span>
                  <Download size={12} className="flex-shrink-0 ml-1 text-stone-400" />
                </a>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Comments toggle */}
      <div className="border-t border-slate-100 pt-3">
        <button
          onClick={() => setShowComments(!showComments)}
          className="inline-flex items-center gap-1.5 text-stone-400 hover:text-stone-600 text-xs font-medium transition-colors"
        >
          <MessageSquare size={15} />
          <span>{post.comments?.length || 0}</span>
        </button>
      </div>

      {/* Comments */}
      {showComments && (
        <div className="mt-3 space-y-2">
          {post.comments?.map((c: Comment) => (
            <div key={c.id} className="bg-slate-50 border border-slate-100 px-3 py-2 rounded group relative">
              <div className="text-xs text-stone-500 mb-0.5">
                <span>{c.user?.email}</span>
                <span className="mx-1 text-stone-300">·</span>
                <span className="text-stone-400">{new Date(c.created_at).toLocaleString(i18n.language, { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}</span>
              </div>
              <div className="text-sm text-stone-700">{c.text}</div>
              {(user?.id === c.user_id || user?.role === 'admin') && (
                <button
                  onClick={() => handleDeleteComment(c.id)}
                  className="absolute top-2 right-2 p-0.5 text-stone-300 hover:text-red-500 opacity-0 group-hover:opacity-100 transition"
                >
                  <Trash2 size={13} />
                </button>
              )}
            </div>
          ))}
          <form onSubmit={handleAddComment} className="flex mt-2 rounded-md overflow-hidden border border-slate-200 focus-within:border-violet-500 focus-within:ring-1 focus-within:ring-violet-500 transition">
            <input
              type="text"
              value={commentText}
              onChange={(e) => setCommentText(e.target.value)}
              placeholder={t('add_comment')}
              className="flex-1 px-3 py-2 text-sm bg-white outline-none text-stone-800 placeholder:text-stone-400"
            />
            <button
              type="submit"
              className="bg-violet-600 text-white px-3 py-2 text-sm hover:bg-violet-500 transition-colors flex items-center"
            >
              <Send size={14} />
            </button>
          </form>
        </div>
      )}
    </div>
  );
};
