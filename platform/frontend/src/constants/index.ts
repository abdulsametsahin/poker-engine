// PokerStreet Brand Colors
export const COLORS = {
  // Primary brand colors
  primary: {
    main: '#7C3AED',      // Deep purple
    light: '#A78BFA',     // Light purple
    dark: '#5B21B6',      // Dark purple
    glow: 'rgba(124, 58, 237, 0.5)',
  },
  secondary: {
    main: '#06B6D4',      // Neon cyan
    light: '#22D3EE',     // Light cyan
    dark: '#0891B2',      // Dark cyan
    glow: 'rgba(6, 182, 212, 0.5)',
  },
  accent: {
    main: '#F59E0B',      // Gold
    light: '#FBB F24',     // Light gold
    dark: '#D97706',      // Dark gold
    glow: 'rgba(245, 158, 11, 0.5)',
  },

  // Semantic colors
  success: {
    main: '#10B981',      // Emerald
    light: '#34D399',
    dark: '#059669',
    glow: 'rgba(16, 185, 129, 0.5)',
  },
  danger: {
    main: '#EF4444',      // Red
    light: '#F87171',
    dark: '#DC2626',
    glow: 'rgba(239, 68, 68, 0.5)',
  },
  warning: {
    main: '#F59E0B',      // Amber
    light: '#FBBF24',
    dark: '#D97706',
    glow: 'rgba(245, 158, 11, 0.5)',
  },
  info: {
    main: '#3B82F6',      // Blue
    light: '#60A5FA',
    dark: '#2563EB',
    glow: 'rgba(59, 130, 246, 0.5)',
  },

  // Background colors
  background: {
    primary: '#0F0F0F',   // Almost black
    secondary: '#1A1A1A', // Dark gray
    tertiary: '#262626',  // Medium gray
    paper: '#1E1E1E',     // Card background
    felt: '#0B6B3E',      // Poker table green
  },

  // Text colors
  text: {
    primary: '#FFFFFF',
    secondary: '#A3A3A3',
    disabled: '#737373',
    inverse: '#000000',
  },

  // UI colors
  border: {
    main: 'rgba(255, 255, 255, 0.1)',
    light: 'rgba(255, 255, 255, 0.05)',
    heavy: 'rgba(255, 255, 255, 0.2)',
  },

  // Poker specific
  poker: {
    felt: 'linear-gradient(135deg, #0B6B3E 0%, #0A5C35 100%)',
    chipGreen: '#10B981',
    chipYellow: '#FBBF24',
    chipRed: '#EF4444',
    chipBlue: '#3B82F6',
    chipBlack: '#000000',
  },
} as const;

// Typography
export const TYPOGRAPHY = {
  // Font families
  fontFamily: {
    primary: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
    mono: '"SF Mono", Monaco, "Cascadia Code", "Roboto Mono", Consolas, "Courier New", monospace',
    display: '"Inter", -apple-system, BlinkMacSystemFont, sans-serif',
  },

  // Font sizes
  fontSize: {
    display: {
      large: '48px',
      medium: '36px',
      small: '24px',
    },
    heading: {
      h1: '32px',
      h2: '28px',
      h3: '24px',
      h4: '20px',
      h5: '18px',
      h6: '16px',
    },
    body: {
      large: '16px',
      medium: '14px',
      small: '12px',
    },
    caption: {
      large: '11px',
      small: '10px',
    },
  },

  // Font weights
  fontWeight: {
    light: 300,
    regular: 400,
    medium: 500,
    semibold: 600,
    bold: 700,
    extrabold: 800,
  },

  // Line heights
  lineHeight: {
    tight: 1.2,
    normal: 1.5,
    relaxed: 1.75,
  },
} as const;

// Spacing
export const SPACING = {
  xs: '4px',
  sm: '8px',
  md: '16px',
  lg: '24px',
  xl: '32px',
  xxl: '48px',
  xxxl: '64px',
} as const;

// Border radius
export const RADIUS = {
  sm: '8px',
  md: '12px',
  lg: '16px',
  xl: '24px',
  full: '9999px',
} as const;

// Transitions
export const TRANSITIONS = {
  fast: '150ms ease-in-out',
  normal: '200ms ease-in-out',
  slow: '300ms ease-in-out',
  verySlow: '500ms ease-in-out',
} as const;

// Shadows
export const SHADOWS = {
  sm: '0 1px 2px 0 rgba(0, 0, 0, 0.3)',
  md: '0 4px 6px -1px rgba(0, 0, 0, 0.4)',
  lg: '0 10px 15px -3px rgba(0, 0, 0, 0.5)',
  xl: '0 20px 25px -5px rgba(0, 0, 0, 0.6)',
  glow: '0 0 20px rgba(124, 58, 237, 0.4)',
  glowStrong: '0 0 30px rgba(124, 58, 237, 0.6)',
} as const;

// Z-index layers
export const Z_INDEX = {
  background: -1,
  base: 0,
  dropdown: 1000,
  sticky: 1100,
  fixed: 1200,
  modalBackdrop: 1300,
  modal: 1400,
  popover: 1500,
  tooltip: 1600,
  toast: 1700,
} as const;

// Breakpoints
export const BREAKPOINTS = {
  xs: 0,
  sm: 600,
  md: 960,
  lg: 1280,
  xl: 1920,
} as const;

// Game constants
export const GAME = {
  // Timing
  ACTION_TIMER_DURATION: 30000,        // 30 seconds
  WINNER_MODAL_DURATION: 5000,         // 5 seconds
  TOAST_DURATION: 2000,                // 2 seconds
  POLLING_INTERVAL: 5000,              // 5 seconds (should be removed)
  RECONNECT_DELAY: 1000,               // 1 second
  RECONNECT_MAX_DELAY: 30000,          // 30 seconds

  // Chips
  CHIP_THRESHOLD_HIGH: 500,
  CHIP_THRESHOLD_LOW: 100,

  // Table
  MAX_PLAYERS: {
    heads_up: 2,
    '3_player': 3,
    '6_player': 6,
  },

  // Buy-in ranges
  BUY_IN: {
    heads_up: { min: 100, max: 1000 },
    '3_player': { min: 100, max: 1000 },
    '6_player': { min: 100, max: 2000 },
  },

  // Blinds
  BLINDS: {
    heads_up: { small: 5, big: 10 },
    '3_player': { small: 5, big: 10 },
    '6_player': { small: 10, big: 20 },
  },

  // UI
  PLAYER_SEAT_SIZE: 120,
  CARD_SIZES: {
    small: { width: 45, height: 65, fontSize: 14 },
    medium: { width: 60, height: 85, fontSize: 18 },
    large: { width: 75, height: 105, fontSize: 22 },
  },
  AVATAR_SIZES: {
    small: 32,
    medium: 48,
    large: 64,
  },

  // Validation
  MIN_RAISE_MULTIPLIER: 2,
  MAX_USERNAME_LENGTH: 20,
  MIN_USERNAME_LENGTH: 3,
  MIN_PASSWORD_LENGTH: 6,
} as const;

// WebSocket
export const WEBSOCKET = {
  HEARTBEAT_INTERVAL: 25000,           // 25 seconds
  RECONNECT_ATTEMPTS: 10,
  RECONNECT_BACKOFF_MULTIPLIER: 1.5,
} as const;

// API
export const API = {
  BASE_URL: process.env.REACT_APP_API_URL || 'http://localhost:8080',
  TIMEOUT: 10000,                      // 10 seconds
  RETRY_ATTEMPTS: 3,
  RETRY_DELAY: 1000,
} as const;

// Routes
export const ROUTES = {
  HOME: '/',
  LOGIN: '/login',
  REGISTER: '/register',
  LOBBY: '/lobby',
  GAME: '/game/:tableId',
  PROFILE: '/profile',
  SETTINGS: '/settings',
} as const;

// Local storage keys
export const STORAGE_KEYS = {
  AUTH_TOKEN: 'poker_auth_token',
  USER_ID: 'poker_user_id',
  USERNAME: 'poker_username',
  THEME: 'poker_theme',
  SETTINGS: 'poker_settings',
} as const;

// Animation keyframes (CSS-in-JS)
export const KEYFRAMES = {
  pulse: {
    '@keyframes pulse': {
      '0%, 100%': {
        opacity: 1,
      },
      '50%': {
        opacity: 0.5,
      },
    },
  },
  bounce: {
    '@keyframes bounce': {
      '0%, 100%': {
        transform: 'translateY(0)',
      },
      '50%': {
        transform: 'translateY(-20px)',
      },
    },
  },
  shimmer: {
    '@keyframes shimmer': {
      '0%': {
        backgroundPosition: '-1000px 0',
      },
      '100%': {
        backgroundPosition: '1000px 0',
      },
    },
  },
  spin: {
    '@keyframes spin': {
      '0%': {
        transform: 'rotate(0deg)',
      },
      '100%': {
        transform: 'rotate(360deg)',
      },
    },
  },
  fadeIn: {
    '@keyframes fadeIn': {
      '0%': {
        opacity: 0,
      },
      '100%': {
        opacity: 1,
      },
    },
  },
  slideUp: {
    '@keyframes slideUp': {
      '0%': {
        transform: 'translateY(20px)',
        opacity: 0,
      },
      '100%': {
        transform: 'translateY(0)',
        opacity: 1,
      },
    },
  },
  glow: {
    '@keyframes glow': {
      '0%, 100%': {
        boxShadow: '0 0 20px rgba(124, 58, 237, 0.4)',
      },
      '50%': {
        boxShadow: '0 0 30px rgba(124, 58, 237, 0.8)',
      },
    },
  },
} as const;
