import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import api from '../api';
import { MessageSquare, Trash2, Edit2, Download, User as UserIcon, Tag, Send, Paperclip } from 'lucide-react';
import { useAuthStore } from '../store';
import type { Post, Hashtag, Person, Attachment, Comment } from '../types';

interface PostCardProps {
  post: Post;
  onUpdate: () => void;
  onEdit: (post: Post) => void;
}

export const PostCard = ({ post, onUpdate, onEdit }: PostCardProps) => {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const [commentText, setCommentText] = useState('');
  const [showComments, setShowComments] = useState(false);

  const handleDelete = async () => {
    if (window.confirm(t('delete') + '?')) {
      await api.delete(`/posts/${post.id}`);
      onUpdate();
    }
  };

  const handleAddComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!commentText.trim()) return;
    try {
      await api.post(`/posts/${post.id}/comments`, { text: commentText });
      setCommentText('');
      onUpdate();
    } catch (err) {
      console.error(err);
    }
  };

  const handleDeleteComment = async (commentId: number) => {
    await api.delete(`/comments/${commentId}`);
    onUpdate();
  };

  return (
    <div className="bg-white rounded-lg shadow p-4 mb-4">
      <div className="flex justify-between items-start mb-3">
        <div>
          <span className="font-semibold text-gray-800">{post.user?.email}</span>
          <span className="text-xs text-gray-500 block">{new Date(post.created_at).toLocaleString()}</span>
        </div>
        {(user?.id === post.user_id || user?.role === 'admin') && (
          <div className="flex space-x-2">
            <button onClick={() => onEdit(post)} className="text-gray-400 hover:text-indigo-600">
              <Edit2 size={18} />
            </button>
            <button onClick={handleDelete} className="text-gray-400 hover:text-red-600">
              <Trash2 size={18} />
            </button>
          </div>
        )}
      </div>

      <p className="text-gray-700 whitespace-pre-wrap mb-4">{post.text}</p>

      <div className="flex flex-wrap gap-2 mb-4">
        {post.hashtags?.map((h: Hashtag) => (
          <span key={h.id} className="bg-indigo-50 text-indigo-700 px-2 py-0.5 rounded text-xs flex items-center">
            <Tag size={12} className="mr-1" /> {h.name}
          </span>
        ))}
        {post.persons?.map((p: Person) => (
          <span key={p.id} className="bg-green-50 text-green-700 px-2 py-0.5 rounded text-xs flex items-center">
            <UserIcon size={12} className="mr-1" /> {p.name}
          </span>
        ))}
      </div>

      {post.attachments?.length > 0 && (
        <div className="border-t pt-3 mb-4">
          <h4 className="text-xs font-semibold text-gray-500 mb-2">{t('attachments')}</h4>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
            {post.attachments.map((a: Attachment) => (
              <div key={a.id} className="space-y-1">
                {a.file_type.startsWith('image/') ? (
                   <img
                     src={`${api.defaults.baseURL}/attachments/${a.id}/download`}
                     alt={a.file_name}
                     className="w-full h-32 object-cover rounded border"
                   />
                ) : (
                   <div className="w-full h-32 bg-gray-100 rounded border flex items-center justify-center text-gray-400">
                      <Paperclip size={24} />
                   </div>
                )}
                <a
                  href={`${api.defaults.baseURL}/attachments/${a.id}/download`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center justify-between p-2 border rounded hover:bg-gray-50 text-sm truncate"
                >
                  <span className="truncate">{a.file_name}</span>
                  <Download size={14} className="flex-shrink-0 ml-1" />
                </a>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="border-t pt-3 flex items-center justify-between">
        <button
          onClick={() => setShowComments(!showComments)}
          className="text-gray-500 hover:text-indigo-600 flex items-center space-x-1 text-sm"
        >
          <MessageSquare size={18} />
          <span>{post.comments?.length || 0}</span>
        </button>
      </div>

      {showComments && (
        <div className="mt-4 space-y-3">
          {post.comments?.map((c: Comment) => (
            <div key={c.id} className="bg-gray-50 p-2 rounded text-sm group relative">
              <div className="font-semibold text-xs mb-1">{c.user?.email}</div>
              <div>{c.text}</div>
              {(user?.id === c.user_id || user?.role === 'admin') && (
                <button
                  onClick={() => handleDeleteComment(c.id)}
                  className="absolute top-2 right-2 text-red-400 opacity-0 group-hover:opacity-100 transition"
                >
                  <Trash2 size={14} />
                </button>
              )}
            </div>
          ))}
          <form onSubmit={handleAddComment} className="flex mt-2">
            <input
              type="text"
              value={commentText}
              onChange={(e) => setCommentText(e.target.value)}
              placeholder={t('add_comment')}
              className="flex-1 border rounded-l-md px-3 py-1 text-sm outline-none focus:ring-1 focus:ring-indigo-500"
            />
            <button
              type="submit"
              className="bg-indigo-600 text-white px-3 py-1 rounded-r-md text-sm hover:bg-indigo-700"
            >
              <Send size={14} />
            </button>
          </form>
        </div>
      )}
    </div>
  );
};
