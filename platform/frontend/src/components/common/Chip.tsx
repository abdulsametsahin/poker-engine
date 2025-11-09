import React from 'react';
import { Box, Typography } from '@mui/material';
import { COLORS, GAME } from '../../constants';
import { formatChips } from '../../utils';

interface ChipProps {
  amount: number;
  variant?: 'default' | 'pot' | 'bet';
  size?: 'small' | 'medium' | 'large';
  showPlus?: boolean;
}

export const Chip: React.FC<ChipProps> = ({
  amount,
  variant = 'default',
  size = 'medium',
  showPlus = false,
}) => {
  const getChipColor = (amt: number): string => {
    if (amt >= GAME.CHIP_THRESHOLD_HIGH) return COLORS.poker.chipGreen;
    if (amt >= GAME.CHIP_THRESHOLD_LOW) return COLORS.poker.chipYellow;
    return COLORS.poker.chipRed;
  };

  const variantStyles = {
    default: {
      background: `linear-gradient(135deg, ${getChipColor(amount)} 0%, ${getChipColor(amount)}dd 100%)`,
      color: amount >= GAME.CHIP_THRESHOLD_LOW ? COLORS.text.inverse : COLORS.text.primary,
    },
    pot: {
      background: `linear-gradient(135deg, ${COLORS.warning.main} 0%, ${COLORS.warning.dark} 100%)`,
      color: COLORS.text.inverse,
    },
    bet: {
      background: `linear-gradient(135deg, ${COLORS.info.main} 0%, ${COLORS.info.dark} 100%)`,
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
    large: {
      padding: '6px 16px',
      fontSize: '0.875rem',
    },
  };

  return (
    <Box
      sx={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 0.5,
        borderRadius: '6px',
        fontFamily: 'monospace',
        fontWeight: 700,
        ...variantStyles[variant],
        ...sizeStyles[size],
        boxShadow: '0 2px 4px rgba(0, 0, 0, 0.3)',
      }}
    >
      <Typography
        component="span"
        sx={{
          fontSize: 'inherit',
          fontWeight: 'inherit',
          fontFamily: 'inherit',
        }}
      >
        {showPlus && amount > 0 && '+'}
        {formatChips(amount)}
      </Typography>
    </Box>
  );
};
