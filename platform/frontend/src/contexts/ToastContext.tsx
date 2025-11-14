import React, { createContext, useContext, useState, ReactNode, useCallback, useRef } from 'react';
import { Snackbar, Alert, AlertColor, Slide, SlideProps } from '@mui/material';
import { ToastMessage } from '../types';
import { GAME } from '../constants';
import { generateId } from '../utils';

interface ToastContextType {
  showToast: (message: string, type?: AlertColor) => void;
  showSuccess: (message: string) => void;
  showError: (message: string) => void;
  showWarning: (message: string) => void;
  showInfo: (message: string) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export const useToast = () => {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
};

interface ToastProviderProps {
  children: ReactNode;
}

function SlideTransition(props: SlideProps) {
  return <Slide {...props} direction="down" />;
}

export const ToastProvider: React.FC<ToastProviderProps> = ({ children }) => {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);
  const [currentToast, setCurrentToast] = useState<ToastMessage | null>(null);
  const recentToastsRef = useRef<Set<string>>(new Set());

  const showToast = useCallback((message: string, type: AlertColor = 'info') => {
    // Create a hash key for deduplication
    const toastKey = `${type}:${message}`;

    // Check if this exact toast was shown recently (within 2 seconds)
    if (recentToastsRef.current.has(toastKey)) {
      console.log(`Toast deduplicated: "${message}"`);
      return;
    }

    const newToast: ToastMessage = {
      id: generateId(),
      message,
      type,
      duration: GAME.TOAST_DURATION,
    };

    setToasts((prev) => [...prev, newToast]);

    // Mark this toast as shown
    recentToastsRef.current.add(toastKey);

    // Remove from deduplication set after 2 seconds
    setTimeout(() => {
      recentToastsRef.current.delete(toastKey);
    }, 2000);

    // If no toast is currently showing, show this one
    if (!currentToast) {
      setCurrentToast(newToast);
    }
  }, [currentToast]);

  const showSuccess = useCallback((message: string) => {
    showToast(message, 'success');
  }, [showToast]);

  const showError = useCallback((message: string) => {
    showToast(message, 'error');
  }, [showToast]);

  const showWarning = useCallback((message: string) => {
    showToast(message, 'warning');
  }, [showToast]);

  const showInfo = useCallback((message: string) => {
    showToast(message, 'info');
  }, [showToast]);

  const handleClose = useCallback(() => {
    setCurrentToast(null);

    // Show next toast in queue after a brief delay
    setTimeout(() => {
      setToasts((prev) => {
        const remaining = prev.slice(1);
        if (remaining.length > 0) {
          setCurrentToast(remaining[0]);
        }
        return remaining;
      });
    }, 200);
  }, []);

  return (
    <ToastContext.Provider
      value={{
        showToast,
        showSuccess,
        showError,
        showWarning,
        showInfo,
      }}
    >
      {children}
      <Snackbar
        open={!!currentToast}
        autoHideDuration={currentToast?.duration || GAME.TOAST_DURATION}
        onClose={handleClose}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        TransitionComponent={SlideTransition}
        sx={{ top: { xs: 16, sm: 24 } }}
      >
        <Alert
          onClose={handleClose}
          severity={currentToast?.type || 'info'}
          variant="filled"
          sx={{
            width: '100%',
            minWidth: 300,
            boxShadow: '0 8px 24px rgba(0, 0, 0, 0.5)',
          }}
        >
          {currentToast?.message}
        </Alert>
      </Snackbar>
    </ToastContext.Provider>
  );
};
