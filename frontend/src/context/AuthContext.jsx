import { createContext, useContext, useState, useEffect, useCallback } from 'react';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  const refreshUser = useCallback(async () => {
    const token = localStorage.getItem('token');
    const storedUser = localStorage.getItem('user');
    if (!token || !storedUser) return;
    try {
      const { getUserProfile } = await import('../api/client');
      const profile = await getUserProfile();
      const parsedUser = JSON.parse(storedUser);
      const updatedUser = {
        ...parsedUser,
        tokens: profile.tokens,
        current_streak: profile.current_streak,
        highest_streak: profile.highest_streak,
      };
      setUser(updatedUser);
      localStorage.setItem('user', JSON.stringify(updatedUser));
    } catch (err) {
      console.error('Failed to sync profile:', err);
    }
  }, []);

  useEffect(() => {
    // Restore user from localStorage on mount
    const token = localStorage.getItem('token');
    const storedUser = localStorage.getItem('user');

    if (token && storedUser) {
      try {
        const parsedUser = JSON.parse(storedUser);
        setUser(parsedUser);
        
        // Sync tokens in background
        refreshUser();
      } catch {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
      }
    }
    setLoading(false);
  }, [refreshUser]);

  const login = useCallback((token, userData) => {
    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(userData));
    setUser(userData);
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    setUser(null);
  }, []);

  const updateLocalUser = useCallback((updatedFields) => {
    const storedUser = localStorage.getItem('user');
    if (storedUser) {
      const parsed = JSON.parse(storedUser);
      const newUserData = { ...parsed, ...updatedFields };
      setUser(newUserData);
      localStorage.setItem('user', JSON.stringify(newUserData));
    }
  }, []);

  const isAuthenticated = !!user;

  return (
    <AuthContext.Provider value={{ user, login, logout, isAuthenticated, loading, refreshUser, updateLocalUser }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

export default AuthContext;
