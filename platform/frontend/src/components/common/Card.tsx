import React from 'react';
import { Box, SxProps, Theme } from '@mui/material';
import { COLORS, RADIUS } from '../../constants';

interface CardProps {
  children: React.ReactNode;
  variant?: 'default' | 'glass' | 'elevated';
  noPadding?: boolean;
  onClick?: () => void;
  sx?: SxProps<Theme>;
}

export const Card: React.FC<CardProps> = ({
  children,
  variant = 'default',
  noPadding = false,
  onClick,
  sx,
}) => {
  const variantStyles = {
    default: {
      background: COLORS.background.paper,
      border: `1px solid ${COLORS.border.main}`,
      boxShadow: '0 2px 8px rgba(0, 0, 0, 0.3)',
    },
    glass: {
      background: 'rgba(30, 30, 30, 0.7)',
      border: `1px solid ${COLORS.border.light}`,
      backdropFilter: 'blur(10px)',
      boxShadow: '0 4px 12px rgba(0, 0, 0, 0.4)',
    },
    elevated: {
      background: COLORS.background.paper,
      border: `1px solid ${COLORS.border.heavy}`,
      boxShadow: '0 8px 24px rgba(0, 0, 0, 0.5)',
    },
  };

  return (
    <Box
      onClick={onClick}
      sx={{
        borderRadius: RADIUS.md,
        padding: noPadding ? 0 : 3,
        transition: 'all 200ms ease-in-out',
        cursor: onClick ? 'pointer' : 'default',
        ...variantStyles[variant],
        ...(onClick && {
          '&:hover': {
            transform: 'translateY(-2px)',
            boxShadow: '0 8px 16px rgba(0, 0, 0, 0.5)',
            borderColor: COLORS.primary.main,
          },
        }),
        ...sx,
      }}
    >
      {children}
    </Box>
  );
};
