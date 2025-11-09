import { createTheme } from '@mui/material/styles';
import { COLORS } from './constants';

// PokerStreet Theme - Premium poker experience with street vibes
export const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: COLORS.primary.main,        // Deep purple
      light: COLORS.primary.light,
      dark: COLORS.primary.dark,
      contrastText: '#FFFFFF',
    },
    secondary: {
      main: COLORS.secondary.main,      // Neon cyan
      light: COLORS.secondary.light,
      dark: COLORS.secondary.dark,
      contrastText: '#FFFFFF',
    },
    error: {
      main: COLORS.danger.main,         // Red
      light: COLORS.danger.light,
      dark: COLORS.danger.dark,
    },
    warning: {
      main: COLORS.warning.main,        // Amber
      light: COLORS.warning.light,
      dark: COLORS.warning.dark,
    },
    info: {
      main: COLORS.info.main,           // Blue
      light: COLORS.info.light,
      dark: COLORS.info.dark,
    },
    success: {
      main: COLORS.success.main,        // Emerald
      light: COLORS.success.light,
      dark: COLORS.success.dark,
    },
    background: {
      default: COLORS.background.primary,   // Almost black
      paper: COLORS.background.paper,       // Dark gray
    },
    text: {
      primary: COLORS.text.primary,
      secondary: COLORS.text.secondary,
      disabled: COLORS.text.disabled,
    },
  },

  typography: {
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',

    // Display
    h1: {
      fontSize: '2.5rem',
      fontWeight: 700,
      letterSpacing: '-0.02em',
    },
    h2: {
      fontSize: '2rem',
      fontWeight: 700,
      letterSpacing: '-0.01em',
    },
    h3: {
      fontSize: '1.75rem',
      fontWeight: 600,
      letterSpacing: '-0.01em',
    },
    h4: {
      fontSize: '1.5rem',
      fontWeight: 600,
    },
    h5: {
      fontSize: '1.25rem',
      fontWeight: 600,
    },
    h6: {
      fontSize: '1rem',
      fontWeight: 600,
    },

    // Body
    body1: {
      fontSize: '1rem',
      lineHeight: 1.5,
    },
    body2: {
      fontSize: '0.875rem',
      lineHeight: 1.5,
    },

    // Small
    caption: {
      fontSize: '0.75rem',
      lineHeight: 1.4,
    },

    // Button
    button: {
      textTransform: 'none',
      fontWeight: 600,
      letterSpacing: '0.02em',
    },
  },

  shape: {
    borderRadius: 12,
  },

  shadows: [
    'none',
    '0 1px 2px 0 rgba(0, 0, 0, 0.3)',
    '0 2px 4px 0 rgba(0, 0, 0, 0.3)',
    '0 4px 6px -1px rgba(0, 0, 0, 0.4)',
    '0 6px 8px -2px rgba(0, 0, 0, 0.4)',
    '0 8px 10px -3px rgba(0, 0, 0, 0.4)',
    '0 10px 15px -3px rgba(0, 0, 0, 0.5)',
    '0 12px 17px -4px rgba(0, 0, 0, 0.5)',
    '0 15px 20px -5px rgba(0, 0, 0, 0.5)',
    '0 18px 23px -6px rgba(0, 0, 0, 0.5)',
    '0 20px 25px -5px rgba(0, 0, 0, 0.6)',
    '0 0 20px rgba(124, 58, 237, 0.4)',          // Glow
    '0 0 30px rgba(124, 58, 237, 0.5)',          // Strong glow
    '0 0 40px rgba(124, 58, 237, 0.6)',          // Very strong glow
    '0 25px 30px -10px rgba(0, 0, 0, 0.6)',
    '0 30px 35px -12px rgba(0, 0, 0, 0.6)',
    '0 35px 40px -15px rgba(0, 0, 0, 0.7)',
    '0 40px 45px -17px rgba(0, 0, 0, 0.7)',
    '0 45px 50px -20px rgba(0, 0, 0, 0.7)',
    '0 50px 55px -22px rgba(0, 0, 0, 0.7)',
    '0 55px 60px -25px rgba(0, 0, 0, 0.8)',
    '0 60px 65px -27px rgba(0, 0, 0, 0.8)',
    '0 65px 70px -30px rgba(0, 0, 0, 0.8)',
    '0 70px 75px -32px rgba(0, 0, 0, 0.8)',
    '0 0 50px rgba(124, 58, 237, 0.7)',          // Maximum glow
  ],

  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: '8px',
          padding: '10px 24px',
          fontSize: '0.875rem',
          fontWeight: 600,
          textTransform: 'none',
          transition: 'all 200ms ease-in-out',
          '&:hover': {
            transform: 'translateY(-1px)',
            boxShadow: '0 4px 12px rgba(0, 0, 0, 0.4)',
          },
        },
        contained: {
          boxShadow: '0 2px 4px rgba(0, 0, 0, 0.3)',
        },
        sizeLarge: {
          padding: '12px 32px',
          fontSize: '1rem',
        },
      },
    },

    MuiCard: {
      styleOverrides: {
        root: {
          borderRadius: '12px',
          backgroundImage: 'none',
          backgroundColor: COLORS.background.paper,
          border: `1px solid ${COLORS.border.main}`,
        },
      },
    },

    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
          backgroundColor: COLORS.background.paper,
        },
      },
    },

    MuiTextField: {
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            borderRadius: '8px',
            transition: 'all 200ms ease-in-out',
            '&:hover': {
              '& .MuiOutlinedInput-notchedOutline': {
                borderColor: COLORS.primary.light,
              },
            },
            '&.Mui-focused': {
              '& .MuiOutlinedInput-notchedOutline': {
                borderColor: COLORS.primary.main,
                borderWidth: '2px',
              },
            },
          },
        },
      },
    },

    MuiChip: {
      styleOverrides: {
        root: {
          borderRadius: '6px',
          fontWeight: 600,
          fontSize: '0.75rem',
        },
      },
    },

    MuiDialog: {
      styleOverrides: {
        paper: {
          borderRadius: '16px',
          border: `1px solid ${COLORS.border.main}`,
        },
      },
    },

    MuiBackdrop: {
      styleOverrides: {
        root: {
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          backdropFilter: 'blur(4px)',
        },
      },
    },

    MuiAppBar: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
          backgroundColor: COLORS.background.paper,
          borderBottom: `1px solid ${COLORS.border.main}`,
        },
      },
    },

    MuiTab: {
      styleOverrides: {
        root: {
          textTransform: 'none',
          fontWeight: 600,
          fontSize: '0.875rem',
          minHeight: 48,
        },
      },
    },
  },
});
