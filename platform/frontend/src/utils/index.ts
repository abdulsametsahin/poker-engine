import { Card, CardObject } from '../types';

// Card utilities
export const parseCard = (card: string | Card): CardObject => {
  if (typeof card === 'object' && 'rank' in card && 'suit' in card) {
    return card as CardObject;
  }

  const cardStr = card as string;

  // Handle formats like "As", "Kh", "10d", "Qc"
  if (cardStr.length >= 2) {
    const rank = cardStr.slice(0, -1);
    const suitChar = cardStr.slice(-1);
    const suitMap: { [key: string]: string } = {
      's': '♠', 'h': '♥', 'd': '♦', 'c': '♣',
      '♠': '♠', '♥': '♥', '♦': '♦', '♣': '♣',
    };
    return {
      rank: rank.toUpperCase(),
      suit: suitMap[suitChar.toLowerCase()] || suitChar,
      display: cardStr,
    };
  }

  // Handle formats like "A♠", "K♥"
  if (cardStr.includes('♠') || cardStr.includes('♥') || cardStr.includes('♦') || cardStr.includes('♣')) {
    const suit = cardStr.match(/[♠♥♦♣]/)?.[0] || '';
    const rank = cardStr.replace(suit, '').toUpperCase();
    return { rank, suit, display: cardStr };
  }

  // Handle formats like "ace_of_spades"
  const parts = cardStr.split('_');
  if (parts.length === 3 && parts[1] === 'of') {
    const rankMap: { [key: string]: string } = {
      ace: 'A', two: '2', three: '3', four: '4', five: '5',
      six: '6', seven: '7', eight: '8', nine: '9', ten: '10',
      jack: 'J', queen: 'Q', king: 'K',
    };
    const suitMap: { [key: string]: string } = {
      spades: '♠', hearts: '♥', diamonds: '♦', clubs: '♣',
    };
    return {
      rank: rankMap[parts[0]] || parts[0].toUpperCase(),
      suit: suitMap[parts[2]] || parts[2],
      display: cardStr,
    };
  }

  return { rank: '', suit: '', display: cardStr };
};

export const getCardColor = (suit: string): string => {
  return suit === '♥' || suit === '♦' ? '#EF4444' : '#000000';
};

export const isRedSuit = (suit: string): boolean => {
  return suit === '♥' || suit === '♦';
};

// Chip formatting
export const formatChips = (amount: number): string => {
  if (amount >= 1000000) {
    return `$${(amount / 1000000).toFixed(1)}M`;
  }
  if (amount >= 1000) {
    return `$${(amount / 1000).toFixed(1)}K`;
  }
  return `$${amount}`;
};

export const formatChipsFull = (amount: number): string => {
  return `$${amount.toLocaleString()}`;
};

// String utilities
export const truncateId = (id: string, length: number = 8): string => {
  if (!id) return '';
  return id.length > length ? `${id.substring(0, length)}...` : id;
};

export const capitalizeFirst = (str: string): string => {
  if (!str) return '';
  return str.charAt(0).toUpperCase() + str.slice(1);
};

export const formatUsername = (username: string | undefined, fallback: string = 'Player'): string => {
  if (!username) return fallback;
  return username.length > 20 ? `${username.substring(0, 20)}...` : username;
};

// Time utilities
export const formatTime = (ms: number): string => {
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) {
    return `${seconds}s`;
  }
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  return `${minutes}m ${remainingSeconds}s`;
};

export const formatTimestamp = (timestamp: string | undefined): string => {
  if (!timestamp) return '';
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;

  return date.toLocaleDateString();
};

// Validation utilities
export const validateEmail = (email: string): boolean => {
  const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return re.test(email);
};

export const validateUsername = (username: string): { valid: boolean; error?: string } => {
  if (!username) return { valid: false, error: 'Username is required' };
  if (username.length < 3) return { valid: false, error: 'Username must be at least 3 characters' };
  if (username.length > 20) return { valid: false, error: 'Username must be at most 20 characters' };
  if (!/^[a-zA-Z0-9_]+$/.test(username)) {
    return { valid: false, error: 'Username can only contain letters, numbers, and underscores' };
  }
  return { valid: true };
};

export const validatePassword = (password: string): { valid: boolean; error?: string } => {
  if (!password) return { valid: false, error: 'Password is required' };
  if (password.length < 6) return { valid: false, error: 'Password must be at least 6 characters' };
  return { valid: true };
};

export const getPasswordStrength = (password: string): {
  strength: 'weak' | 'medium' | 'strong';
  score: number;
} => {
  let score = 0;
  if (password.length >= 8) score++;
  if (password.length >= 12) score++;
  if (/[a-z]/.test(password) && /[A-Z]/.test(password)) score++;
  if (/\d/.test(password)) score++;
  if (/[^a-zA-Z0-9]/.test(password)) score++;

  if (score <= 2) return { strength: 'weak', score };
  if (score <= 3) return { strength: 'medium', score };
  return { strength: 'strong', score };
};

// Game utilities
export const getGameModeName = (mode: string): string => {
  const names: { [key: string]: string } = {
    heads_up: 'Heads-Up',
    '3_player': '3-Player',
    '6_player': '6-Player',
  };
  return names[mode] || mode;
};

export const getBettingRoundName = (round: string): string => {
  const names: { [key: string]: string } = {
    preflop: 'PREFLOP',
    flop: 'FLOP',
    turn: 'TURN',
    river: 'RIVER',
  };
  return names[round] || round.toUpperCase();
};

export const getActionColor = (action: string): string => {
  const colors: { [key: string]: string } = {
    fold: '#EF4444',
    check: '#3B82F6',
    call: '#10B981',
    raise: '#7C3AED',
    all_in: '#F59E0B',
  };
  return colors[action] || '#A3A3A3';
};

// Array utilities
export const shuffleArray = <T>(array: T[]): T[] => {
  const shuffled = [...array];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }
  return shuffled;
};

export const range = (start: number, end: number): number[] => {
  return Array.from({ length: end - start + 1 }, (_, i) => start + i);
};

// Local storage utilities
export const getStorageItem = (key: string): string | null => {
  try {
    return localStorage.getItem(key);
  } catch (error) {
    console.error('Error reading from localStorage:', error);
    return null;
  }
};

export const setStorageItem = (key: string, value: string): void => {
  try {
    localStorage.setItem(key, value);
  } catch (error) {
    console.error('Error writing to localStorage:', error);
  }
};

export const removeStorageItem = (key: string): void => {
  try {
    localStorage.removeItem(key);
  } catch (error) {
    console.error('Error removing from localStorage:', error);
  }
};

export const clearStorage = (): void => {
  try {
    localStorage.clear();
  } catch (error) {
    console.error('Error clearing localStorage:', error);
  }
};

// Number utilities
export const clamp = (value: number, min: number, max: number): number => {
  return Math.min(Math.max(value, min), max);
};

export const randomInt = (min: number, max: number): number => {
  return Math.floor(Math.random() * (max - min + 1)) + min;
};

// Delay utility
export const delay = (ms: number): Promise<void> => {
  return new Promise((resolve) => setTimeout(resolve, ms));
};

// UUID generator (simple version)
export const generateId = (): string => {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
};

// Debounce utility
export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): ((...args: Parameters<T>) => void) => {
  let timeout: NodeJS.Timeout | null = null;
  return (...args: Parameters<T>) => {
    if (timeout) clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
};

// Throttle utility
export const throttle = <T extends (...args: any[]) => any>(
  func: T,
  limit: number
): ((...args: Parameters<T>) => void) => {
  let inThrottle: boolean = false;
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args);
      inThrottle = true;
      setTimeout(() => (inThrottle = false), limit);
    }
  };
};

// Check if user is current player
export const isCurrentPlayer = (userId: string, cards?: any[]): boolean => {
  // If we have cards and they have rank/suit, we're the current player
  return !!(cards && cards.length > 0 && cards[0].rank);
};

// Calculate win rate
export const calculateWinRate = (gamesWon: number, gamesPlayed: number): number => {
  if (gamesPlayed === 0) return 0;
  return Math.round((gamesWon / gamesPlayed) * 100);
};

// Format percentage
export const formatPercentage = (value: number): string => {
  return `${Math.round(value)}%`;
};
