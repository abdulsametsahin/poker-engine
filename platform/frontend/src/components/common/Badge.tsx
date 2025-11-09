import React from 'react';
import { Box, Typography } from '@mui/material';
import { COLORS, RADIUS } from '../../constants';

interface BadgeProps {
  children: React.ReactNode;
  variant?: 'success' | 'danger' | 'warning' | 'info' | 'primary' | 'secondary';
  size?: 'small' | 'medium';
  pulse?: boolean;
}

export const Badge: React.FC<BadgeProps> = ({
  children,
  variant = 'primary',
  size = 'medium',
  pulse = false,
}) => {
  const variantStyles = {
    primary: {
      background: COLORS.primary.main,
      color: COLORS.text.primary,
    },
    secondary: {
      background: COLORS.secondary.main,
      color: COLORS.text.primary,
    },
    success: {
      background: COLORS.success.main,
      color: COLORS.text.primary,
    },
    danger: {
      background: COLORS.danger.main,
      color: COLORS.text.primary,
    },
    warning: {
      background: COLORS.warning.main,
      color: COLORS.text.inverse,
    },
    info: {
      background: COLORS.info.main,
      color: COLORS.text.primary,
    },
  };

  const sizeStyles = {
    small: {
      padding: '2px 8px',
      fontSize: '0.625rem',
    },
    medium: {
      padding: '4px 12px',
      fontSize: '0.75rem',
    },
  };

  return (
    <Box
      sx={{
        display: 'inline-flex',
        alignItems: 'center',
        borderRadius: RADIUS.sm,
        fontWeight: 600,
        textTransform: 'uppercase',
        letterSpacing: '0.05em',
        ...variantStyles[variant],
        ...sizeStyles[size],
        ...(pulse && {
          '@keyframes pulse': {
            '0%, 100%': {
              opacity: 1,
            },
            '50%': {
              opacity: 0.6,
            },
          },
          animation: 'pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        }),
      }}
    >
      <Typography
        component="span"
        sx={{
          fontSize: 'inherit',
          fontWeight: 'inherit',
          lineHeight: 1,
        }}
      >
        {children}
      </Typography>
    </Box>
  );
};
