import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { User } from '../types';
import { STORAGE_KEYS } from '../constants';
import { getStorageItem, setStorageItem, removeStorageItem } from '../utils';

interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (token: string, userId: string, username: string) => void;
  logout: () => void;
  updateUser: (updates: Partial<User>) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Initialize auth from localStorage
  useEffect(() => {
    const storedToken = getStorageItem(STORAGE_KEYS.AUTH_TOKEN);
    const storedUserId = getStorageItem(STORAGE_KEYS.USER_ID);
    const storedUsername = getStorageItem(STORAGE_KEYS.USERNAME);

    if (storedToken && storedUserId && storedUsername) {
      setToken(storedToken);
      setUser({
        id: storedUserId,
        username: storedUsername,
      });
    }

    setIsLoading(false);
  }, []);

  const login = (newToken: string, userId: string, username: string) => {
    setToken(newToken);
    setUser({
      id: userId,
      username,
    });

    setStorageItem(STORAGE_KEYS.AUTH_TOKEN, newToken);
    setStorageItem(STORAGE_KEYS.USER_ID, userId);
    setStorageItem(STORAGE_KEYS.USERNAME, username);
  };

  const logout = () => {
    setToken(null);
    setUser(null);

    removeStorageItem(STORAGE_KEYS.AUTH_TOKEN);
    removeStorageItem(STORAGE_KEYS.USER_ID);
    removeStorageItem(STORAGE_KEYS.USERNAME);
  };

  const updateUser = (updates: Partial<User>) => {
    if (user) {
      const updatedUser = { ...user, ...updates };
      setUser(updatedUser);

      if (updates.username) {
        setStorageItem(STORAGE_KEYS.USERNAME, updates.username);
      }
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isAuthenticated: !!token,
        isLoading,
        login,
        logout,
        updateUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};
