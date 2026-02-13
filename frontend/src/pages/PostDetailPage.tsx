import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useParams } from 'react-router-dom';
import { apiFetch } from '../api/client';

interface Comment {
  id: number;
  text: string;
  author_email: string;
  created_at: string;
}

interface Attachment {
  id: number;
  file_name: string;
  url: string;
}

interface Post {
  id: number;
  text: string;
  date: string;
  comments: Comment[];
  attachments: Attachment[];
}

const asArray = <T,>(value: unknown): T[] => (Array.isArray(value) ? (value as T[]) : []);

const PostDetailPage = () => {
  const { t } = useTranslation();
  const { id } = useParams();
  const [post, setPost] = useState<Post | null>(null);
  const [comment, setComment] = useState('');
  const [commentError, setCommentError] = useState('');

  const loadPost = async () => {
    if (!id) return;
    const data = await apiFetch(`/posts/${id}`);
    setPost({
      ...data,
      comments: asArray<Comment>(data.comments),
      attachments: asArray<Attachment>(data.attachments)
    });
  };

  useEffect(() => {
    void loadPost();
  }, [id]);

  const handleComment = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!id) return;
    setCommentError('');
    if (comment.trim() === '') {
      setCommentError(t('validation.commentRequired'));
      return;
    }
    try {
      await apiFetch(`/posts/${id}/comments`, {
        method: 'POST',
        body: JSON.stringify({ text: comment })
      });
      setComment('');
      await loadPost();
    } catch (err) {
      setCommentError(String(err));
    }
  };

  if (!post) {
    return <div className="p-6">Loading...</div>;
  }

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-semibold">{t('post.title')}</h2>
        <Link to={`/posts/${post.id}/edit`} className="text-sm text-slate-600">{t('post.update')}</Link>
      </div>
      <p className="text-sm text-slate-500">{post.date.slice(0, 10)}</p>
      <p className="whitespace-pre-wrap bg-white p-4 rounded shadow-sm">{post.text}</p>
      <div>
        <h3 className="font-semibold">{t('post.attachments')}</h3>
        <ul className="space-y-2">
          {post.attachments.map((attachment) => (
            <li key={attachment.id}>
              <a className="text-slate-600 underline" href={attachment.url} target="_blank" rel="noreferrer">
                {attachment.file_name}
              </a>
            </li>
          ))}
        </ul>
      </div>
      <div>
        <h3 className="font-semibold">{t('post.comments')}</h3>
        <ul className="space-y-2">
          {post.comments.map((item) => (
            <li key={item.id} className="bg-white p-3 rounded shadow-sm">
              <p className="text-sm text-slate-500">{item.author_email}</p>
              <p>{item.text}</p>
            </li>
          ))}
        </ul>
        <form onSubmit={handleComment} className="mt-4 flex gap-2">
          <input
            className="flex-1 border rounded px-3 py-2"
            value={comment}
            onChange={(event) => setComment(event.target.value)}
            placeholder={t('post.addComment')}
          />
          <button className="bg-slate-900 text-white px-3 rounded" type="submit">
            {t('post.addComment')}
          </button>
        </form>
        {commentError && <p className="text-sm text-red-600 mt-2">{commentError}</p>}
      </div>
    </div>
  );
};

export default PostDetailPage;
