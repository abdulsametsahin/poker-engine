import React from 'react';
import { AppBar, Toolbar, Box, IconButton, Menu, MenuItem, Typography, Button as MuiButton } from '@mui/material';
import { AccountCircle, Logout, Settings as SettingsIcon, Home, EmojiEvents } from '@mui/icons-material';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { useWebSocket } from '../../contexts/WebSocketContext';
import { Logo } from './Logo';
import { Badge } from './Badge';
import { COLORS, ROUTES } from '../../constants';

interface AppLayoutProps {
  children: React.ReactNode;
  showHeader?: boolean;
}

export const AppLayout: React.FC<AppLayoutProps> = ({ children, showHeader = true }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuth();
  const { isConnected } = useWebSocket();
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);

  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = () => {
    handleMenuClose();
    logout();
    navigate(ROUTES.LOGIN);
  };

  const handleLogoClick = () => {
    navigate(ROUTES.LOBBY);
  };

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        flexDirection: 'column',
        background: COLORS.background.primary,
      }}
    >
      {showHeader && (
        <AppBar
          position="static"
          elevation={0}
          sx={{
            background: COLORS.background.paper,
            borderBottom: `1px solid ${COLORS.border.main}`,
          }}
        >
          <Toolbar>
            {/* Logo */}
            <Logo size="small" onClick={handleLogoClick} />

            {/* Navigation */}
            <Box sx={{ ml: 4, display: 'flex', gap: 1 }}>
              <MuiButton
                startIcon={<Home />}
                onClick={() => navigate(ROUTES.LOBBY)}
                sx={{
                  color: location.pathname === ROUTES.LOBBY ? COLORS.primary.main : 'text.secondary',
                  fontWeight: location.pathname === ROUTES.LOBBY ? 700 : 500,
                  '&:hover': { background: COLORS.background.secondary },
                }}
              >
                Lobby
              </MuiButton>
              <MuiButton
                startIcon={<EmojiEvents />}
                onClick={() => navigate('/tournaments')}
                sx={{
                  color: location.pathname === '/tournaments' ? COLORS.primary.main : 'text.secondary',
                  fontWeight: location.pathname === '/tournaments' ? 700 : 500,
                  '&:hover': { background: COLORS.background.secondary },
                }}
              >
                Tournaments
              </MuiButton>
            </Box>

            {/* Spacer */}
            <Box sx={{ flexGrow: 1 }} />

            {/* Connection Status */}
            <Box sx={{ mr: 2 }}>
              <Badge
                variant={isConnected ? 'success' : 'danger'}
                size="small"
                pulse={!isConnected}
              >
                {isConnected ? 'Connected' : 'Disconnected'}
              </Badge>
            </Box>

            {/* User Menu */}
            {user && (
              <>
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 2,
                    mr: 1,
                    px: 2,
                    py: 0.75,
                    borderRadius: '8px',
                    background: COLORS.background.secondary,
                  }}
                >
                  <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end' }}>
                    <Typography variant="body2" fontWeight={600}>
                      {user.username}
                    </Typography>
                    <Typography variant="caption" color={COLORS.accent.main} fontWeight={600}>
                      ${user.chips?.toLocaleString() || 0}
                    </Typography>
                  </Box>
                </Box>

                <IconButton
                  size="large"
                  edge="end"
                  aria-label="account menu"
                  aria-controls="account-menu"
                  aria-haspopup="true"
                  onClick={handleMenuOpen}
                  color="inherit"
                >
                  <AccountCircle />
                </IconButton>

                <Menu
                  id="account-menu"
                  anchorEl={anchorEl}
                  anchorOrigin={{
                    vertical: 'bottom',
                    horizontal: 'right',
                  }}
                  keepMounted
                  transformOrigin={{
                    vertical: 'top',
                    horizontal: 'right',
                  }}
                  open={Boolean(anchorEl)}
                  onClose={handleMenuClose}
                  PaperProps={{
                    sx: {
                      mt: 1,
                      minWidth: 180,
                    },
                  }}
                >
                  <MenuItem onClick={() => { handleMenuClose(); navigate(ROUTES.SETTINGS); }}>
                    <SettingsIcon sx={{ mr: 1.5, fontSize: '1.25rem' }} />
                    Settings
                  </MenuItem>
                  <MenuItem onClick={handleLogout}>
                    <Logout sx={{ mr: 1.5, fontSize: '1.25rem' }} />
                    Logout
                  </MenuItem>
                </Menu>
              </>
            )}
          </Toolbar>
        </AppBar>
      )}

      {/* Main Content */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        {children}
      </Box>
    </Box>
  );
};
