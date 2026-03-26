import { create } from "zustand";
import api from "./api";

export interface User {
  id: string;
  githubId: string;
  email: string;
  displayName: string;
  avatarUrl: string;
  nim: string;
  role: "internal" | "external" | "superadmin";
  createdAt: string;
  updatedAt: string;
}

interface AuthState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: () => void;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
  setUser: (user: User | null) => void;
}

let refreshTimer: NodeJS.Timeout | null = null;

export const useAuth = create<AuthState>((set, get) => ({
  user: null,
  isLoading: true,
  isAuthenticated: false,

  login: () => {
    window.location.href = "/api/auth/login";
  },

  logout: async () => {
    try {
      await api.post("/auth/logout");
    } catch (error) {
      console.error("Logout error:", error);
    }
    set({ user: null, isAuthenticated: false, isLoading: false });
    window.location.href = "/";
  },

  refreshUser: async () => {
    try {
      const { data } = await api.get("/users/me");
      set({ user: data, isAuthenticated: true, isLoading: false });

      // Schedule silent refresh at 50% expiry (7.5 minutes for 15-min token)
      if (refreshTimer) clearTimeout(refreshTimer);
      refreshTimer = setTimeout(() => {
        api
          .post("/auth/refresh")
          .then(({ data }) => {
            if (data.access_token) {
              // Token is stored in api interceptor
            }
          })
          .catch(() => {
            get().logout();
          });
      }, 7.5 * 60 * 1000);
    } catch (error) {
      set({ user: null, isAuthenticated: false, isLoading: false });
    }
  },

  setUser: (user) => {
    set({ user, isAuthenticated: !!user, isLoading: false });
  },
}));

// Initialize auth on mount
export function initAuth() {
  const { refreshUser } = useAuth.getState();
  refreshUser();
}
