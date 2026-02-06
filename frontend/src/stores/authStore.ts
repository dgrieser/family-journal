import { create } from 'zustand';
import { apiFetch } from '../api/client';

interface User {
  id: number;
  email: string;
  role: string;
  active: boolean;
}

interface AuthState {
  user: User | null;
  loading: boolean;
  fetchProfile: () => Promise<void>;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  loading: false,
  fetchProfile: async () => {
    set({ loading: true });
    try {
      const user = await apiFetch('/auth/profile');
      set({ user, loading: false });
    } catch {
      set({ user: null, loading: false });
    }
  },
  login: async (email, password) => {
    const user = await apiFetch('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password })
    });
    set({ user });
  },
  register: async (email, password) => {
    await apiFetch('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password })
    });
    set({ user: null });
  },
  logout: async () => {
    await apiFetch('/auth/logout', { method: 'POST' });
    set({ user: null });
  }
}));
