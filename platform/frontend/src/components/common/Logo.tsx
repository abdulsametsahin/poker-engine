import React from 'react';
import { Box, Typography } from '@mui/material';
import { COLORS } from '../../constants';

interface LogoProps {
  size?: 'small' | 'medium' | 'large';
  variant?: 'full' | 'icon';
  onClick?: () => void;
}

export const Logo: React.FC<LogoProps> = ({
  size = 'medium',
  variant = 'full',
  onClick
}) => {
  const sizes = {
    small: { fontSize: '1.25rem', iconSize: 24 },
    medium: { fontSize: '1.75rem', iconSize: 32 },
    large: { fontSize: '2.5rem', iconSize: 48 },
  };

  const currentSize = sizes[size];

  return (
    <Box
      onClick={onClick}
      sx={{
        display: 'flex',
        alignItems: 'center',
        gap: 1,
        cursor: onClick ? 'pointer' : 'default',
        transition: 'transform 200ms ease-in-out',
        '&:hover': onClick ? {
          transform: 'scale(1.05)',
        } : {},
      }}
    >
      {/* Street sign icon */}
      <Box
        sx={{
          width: currentSize.iconSize,
          height: currentSize.iconSize,
          background: `linear-gradient(135deg, ${COLORS.primary.main} 0%, ${COLORS.primary.dark} 100%)`,
          borderRadius: '6px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          boxShadow: `0 0 20px ${COLORS.primary.glow}`,
          position: 'relative',
          overflow: 'hidden',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: 'linear-gradient(135deg, rgba(255,255,255,0.2) 0%, transparent 100%)',
          },
        }}
      >
        <Typography
          sx={{
            fontSize: currentSize.iconSize * 0.6,
            fontWeight: 800,
            color: COLORS.text.primary,
            textShadow: '0 2px 4px rgba(0,0,0,0.3)',
            fontFamily: 'Georgia, serif',
            fontStyle: 'italic',
          }}
        >
          P
        </Typography>
      </Box>

      {variant === 'full' && (
        <Box>
          <Typography
            sx={{
              fontSize: currentSize.fontSize,
              fontWeight: 700,
              background: `linear-gradient(135deg, ${COLORS.primary.light} 0%, ${COLORS.secondary.light} 100%)`,
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              backgroundClip: 'text',
              letterSpacing: '-0.02em',
              lineHeight: 1,
            }}
          >
            Poker<span style={{ fontWeight: 400 }}>Street</span>
          </Typography>
        </Box>
      )}
    </Box>
  );
};
