import React from 'react';
import { Avatar as MuiAvatar, Box } from '@mui/material';
import { AccountCircle } from '@mui/icons-material';
import { COLORS } from '../../constants';

interface AvatarProps {
  username?: string;
  src?: string;
  size?: 'small' | 'medium' | 'large';
  online?: boolean;
  dealer?: boolean;
}

export const Avatar: React.FC<AvatarProps> = ({
  username,
  src,
  size = 'medium',
  online = false,
  dealer = false,
}) => {
  const sizes = {
    small: 32,
    medium: 48,
    large: 64,
  };

  const avatarSize = sizes[size];

  const getInitials = (name?: string): string => {
    if (!name) return '?';
    const parts = name.split(' ');
    if (parts.length >= 2) {
      return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
    }
    return name.substring(0, 2).toUpperCase();
  };

  const getColorFromName = (name?: string): string => {
    if (!name) return COLORS.primary.main;
    const colors = [
      COLORS.primary.main,
      COLORS.secondary.main,
      COLORS.success.main,
      COLORS.warning.main,
      COLORS.info.main,
    ];
    const index = name.charCodeAt(0) % colors.length;
    return colors[index];
  };

  return (
    <Box sx={{ position: 'relative', display: 'inline-block' }}>
      <MuiAvatar
        src={src}
        sx={{
          width: avatarSize,
          height: avatarSize,
          bgcolor: getColorFromName(username),
          fontSize: avatarSize * 0.4,
          fontWeight: 600,
          border: online ? `2px solid ${COLORS.success.main}` : 'none',
          boxShadow: online ? `0 0 12px ${COLORS.success.glow}` : 'none',
        }}
      >
        {src ? null : username ? getInitials(username) : <AccountCircle sx={{ fontSize: avatarSize * 0.8 }} />}
      </MuiAvatar>

      {/* Online indicator */}
      {online && (
        <Box
          sx={{
            position: 'absolute',
            bottom: 0,
            right: 0,
            width: avatarSize * 0.25,
            height: avatarSize * 0.25,
            borderRadius: '50%',
            bgcolor: COLORS.success.main,
            border: `2px solid ${COLORS.background.paper}`,
            boxShadow: `0 0 8px ${COLORS.success.glow}`,
          }}
        />
      )}

      {/* Dealer indicator */}
      {dealer && (
        <Box
          sx={{
            position: 'absolute',
            top: -4,
            right: -4,
            width: avatarSize * 0.35,
            height: avatarSize * 0.35,
            borderRadius: '50%',
            bgcolor: COLORS.danger.main,
            border: `2px solid ${COLORS.background.paper}`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: avatarSize * 0.2,
            fontWeight: 700,
            color: COLORS.text.primary,
            boxShadow: `0 0 8px ${COLORS.danger.glow}`,
          }}
        >
          D
        </Box>
      )}
    </Box>
  );
};
