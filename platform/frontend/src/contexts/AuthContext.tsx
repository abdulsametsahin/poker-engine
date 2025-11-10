import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { User } from '../types';
import { STORAGE_KEYS } from '../constants';
import { getStorageItem, setStorageItem, removeStorageItem } from '../utils';
import { authAPI } from '../services/api';

interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (token: string, user: User) => void;
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

  // Initialize auth from localStorage and fetch fresh user data
  useEffect(() => {
    const initAuth = async () => {
      const storedToken = getStorageItem(STORAGE_KEYS.AUTH_TOKEN);

      if (storedToken) {
        setToken(storedToken);

        // Try to fetch fresh user data from server
        try {
          const response = await authAPI.getCurrentUser();
          const userData = response.data;
          setUser(userData);
          setStorageItem(STORAGE_KEYS.USER_DATA, JSON.stringify(userData));
        } catch (error) {
          // If fetch fails, use stored data as fallback
          const storedUserData = getStorageItem(STORAGE_KEYS.USER_DATA);
          if (storedUserData) {
            try {
              const userData = JSON.parse(storedUserData);
              setUser(userData);
            } catch (parseError) {
              // If parsing fails, fall back to basic user info
              const storedUserId = getStorageItem(STORAGE_KEYS.USER_ID);
              const storedUsername = getStorageItem(STORAGE_KEYS.USERNAME);
              if (storedUserId && storedUsername) {
                setUser({
                  id: storedUserId,
                  username: storedUsername,
                });
              }
            }
          }
        }
      }

      setIsLoading(false);
    };

    initAuth();
  }, []);

  const login = (newToken: string, userData: User) => {
    setToken(newToken);
    setUser(userData);

    setStorageItem(STORAGE_KEYS.AUTH_TOKEN, newToken);
    setStorageItem(STORAGE_KEYS.USER_ID, userData.id);
    setStorageItem(STORAGE_KEYS.USERNAME, userData.username);
    setStorageItem(STORAGE_KEYS.USER_DATA, JSON.stringify(userData));
  };

  const logout = () => {
    setToken(null);
    setUser(null);

    removeStorageItem(STORAGE_KEYS.AUTH_TOKEN);
    removeStorageItem(STORAGE_KEYS.USER_ID);
    removeStorageItem(STORAGE_KEYS.USERNAME);
    removeStorageItem(STORAGE_KEYS.USER_DATA);
  };

  const updateUser = (updates: Partial<User>) => {
    if (user) {
      const updatedUser = { ...user, ...updates };
      setUser(updatedUser);

      if (updates.username) {
        setStorageItem(STORAGE_KEYS.USERNAME, updates.username);
      }
      // Update stored user data
      setStorageItem(STORAGE_KEYS.USER_DATA, JSON.stringify(updatedUser));
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
