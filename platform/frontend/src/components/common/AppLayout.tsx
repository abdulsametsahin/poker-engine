import React from 'react';
import { AppBar, Toolbar, Box, IconButton, Menu, MenuItem, Typography } from '@mui/material';
import { AccountCircle, Logout, Settings as SettingsIcon } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
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
                    gap: 1,
                    mr: 1,
                  }}
                >
                  <Typography variant="body2" color="text.secondary">
                    {user.username}
                  </Typography>
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
                  <MenuItem onClick={handleMenuClose}>
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
