import { create } from 'zustand'

const TOKEN_KEY = 'token'

function readToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

interface AuthState {
  token: string | null
  setToken: (token: string) => void
  logout: () => void
}

export const useAuth = create<AuthState>((set) => ({
  token: readToken(),
  setToken: (token) => {
    try {
      localStorage.setItem(TOKEN_KEY, token)
    } catch {
      // storage unavailable; keep the token in memory only
    }
    set({ token })
  },
  logout: () => {
    try {
      localStorage.removeItem(TOKEN_KEY)
    } catch {
      // ignore
    }
    set({ token: null })
  },
}))

// Keep the store in sync when the token changes in another tab
if (typeof window !== 'undefined') {
  window.addEventListener('storage', (e) => {
    if (e.key === TOKEN_KEY || e.key === null) {
      useAuth.setState({ token: readToken() })
    }
  })
}
