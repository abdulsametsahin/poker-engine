import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { User } from '../types';
import { STORAGE_KEYS } from '../constants';
import { getStorageItem, setStorageItem, removeStorageItem } from '../utils';
import { authAPI } from '../services/api';

interface StoredUserData {
  user: User;
  _storedAt: number;
}

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

          // Store with timestamp
          const dataToStore: StoredUserData = {
            user: userData,
            _storedAt: Date.now()
          };
          setStorageItem(STORAGE_KEYS.USER_DATA, JSON.stringify(dataToStore));
        } catch (error) {
          console.warn('Failed to fetch fresh user data, checking stored data:', error);

          // If fetch fails, use stored data as fallback only if it's fresh
          const storedUserData = getStorageItem(STORAGE_KEYS.USER_DATA);
          if (storedUserData) {
            try {
              const parsed: StoredUserData = JSON.parse(storedUserData);

              // Check if stored data has timestamp (backward compatibility)
              if (parsed._storedAt) {
                const age = Date.now() - parsed._storedAt;
                const FIVE_MINUTES = 5 * 60 * 1000;

                if (age < FIVE_MINUTES) {
                  setUser(parsed.user);
                  console.log('Using cached user data (age: ' + Math.round(age / 1000) + 's)');
                } else {
                  console.warn('Stored user data is stale (age: ' + Math.round(age / 60000) + 'm), please refresh or re-login');
                  // Still set the user but log a warning
                  setUser(parsed.user);
                }
              } else {
                // Old format without timestamp - treat as potentially stale
                console.warn('Stored user data has no timestamp, treating as stale');
                const userData = parsed as any;
                if (userData.id && userData.username) {
                  setUser(userData);
                }
              }
            } catch (parseError) {
              console.error('Failed to parse stored user data:', parseError);
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

    const dataToStore: StoredUserData = {
      user: userData,
      _storedAt: Date.now()
    };

    setStorageItem(STORAGE_KEYS.AUTH_TOKEN, newToken);
    setStorageItem(STORAGE_KEYS.USER_ID, userData.id);
    setStorageItem(STORAGE_KEYS.USERNAME, userData.username);
    setStorageItem(STORAGE_KEYS.USER_DATA, JSON.stringify(dataToStore));
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

      // Update stored user data with fresh timestamp
      const dataToStore: StoredUserData = {
        user: updatedUser,
        _storedAt: Date.now()
      };
      setStorageItem(STORAGE_KEYS.USER_DATA, JSON.stringify(dataToStore));
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
