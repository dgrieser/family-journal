import { create } from 'zustand';

interface User {
  id: number;
  email: string;
  role: 'admin' | 'user';
}

interface AuthState {
  user: User | null;
  setUser: (user: User | null) => void;
  isAuthenticated: boolean;
  initialized: boolean;
  setInitialized: (initialized: boolean) => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  initialized: false,
  setUser: (user) => set({ user, isAuthenticated: !!user, initialized: true }),
  setInitialized: (initialized) => set({ initialized }),
}));
