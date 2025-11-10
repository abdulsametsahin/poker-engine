import React, { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider, CssBaseline } from '@mui/material';
import { theme } from './theme';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { ToastProvider } from './contexts/ToastContext';
import { WebSocketProvider } from './contexts/WebSocketContext';
import { LoadingSpinner } from './components/common';

// Lazy load pages for better performance
const Login = lazy(() => import('./pages/Login').then(module => ({ default: module.Login })));
const Lobby = lazy(() => import('./pages/Lobby').then(module => ({ default: module.Lobby })));
const GameView = lazy(() => import('./pages/GameView').then(module => ({ default: module.GameView })));
const Settings = lazy(() => import('./pages/Settings').then(module => ({ default: module.Settings })));
const Tournaments = lazy(() => import('./pages/Tournaments').then(module => ({ default: module.Tournaments })));

const ProtectedRoute: React.FC<{ children: React.ReactElement }> = ({ children }) => {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return <LoadingSpinner fullScreen message="Loading..." />;
  }

  return isAuthenticated ? children : <Navigate to="/login" />;
};

const AppRoutes: React.FC = () => {
  return (
    <Suspense fallback={<LoadingSpinner fullScreen message="Loading..." />}>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/lobby"
          element={
            <ProtectedRoute>
              <Lobby />
            </ProtectedRoute>
          }
        />
        <Route
          path="/game/:tableId"
          element={
            <ProtectedRoute>
              <GameView />
            </ProtectedRoute>
          }
        />
        <Route
          path="/settings"
          element={
            <ProtectedRoute>
              <Settings />
            </ProtectedRoute>
          }
        />
        <Route
          path="/tournaments"
          element={
            <ProtectedRoute>
              <Tournaments />
            </ProtectedRoute>
          }
        />
        <Route path="/" element={<Navigate to="/login" />} />
      </Routes>
    </Suspense>
  );
};

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <BrowserRouter>
        <AuthProvider>
          <ToastProvider>
            <WebSocketProvider>
              <AppRoutes />
            </WebSocketProvider>
          </ToastProvider>
        </AuthProvider>
      </BrowserRouter>
    </ThemeProvider>
  );
}

export default App;
