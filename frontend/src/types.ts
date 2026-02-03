export interface User {
  id: number;
  email: string;
  role: 'admin' | 'user';
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Person {
  id: number;
  name: string;
  description: string;
  created_by_user_id: number;
  created_at: string;
  updated_at: string;
}

export interface Hashtag {
  id: number;
  name: string;
  created_at: string;
}

export interface Attachment {
  id: number;
  post_id: number;
  file_name: string;
  file_type: string;
  file_size: number;
  storage_path: string;
  created_at: string;
}

export interface Comment {
  id: number;
  post_id: number;
  user_id: number;
  user?: User;
  text: string;
  created_at: string;
  updated_at: string;
}

export interface Post {
  id: number;
  user_id: number;
  user?: User;
  date: string;
  text: string;
  hashtags: Hashtag[];
  mentions: Person[];
  attachments: Attachment[];
  comments: Comment[];
  created_at: string;
  updated_at: string;
}
