import React, { useState, useEffect } from 'react';
import { Box, Typography } from '@mui/material';
import { keyframes } from '@mui/system';

interface BalanceAnimationProps {
  change: number;
  onComplete?: () => void;
}

const floatUp = keyframes`
  0% {
    opacity: 0;
    transform: translateY(0);
  }
  20% {
    opacity: 1;
  }
  100% {
    opacity: 0;
    transform: translateY(-40px);
  }
`;

export const BalanceAnimation: React.FC<BalanceAnimationProps> = ({ change, onComplete }) => {
  const [isVisible, setIsVisible] = useState(true);

  useEffect(() => {
    const timer = setTimeout(() => {
      setIsVisible(false);
      onComplete?.();
    }, 2000);

    return () => clearTimeout(timer);
  }, [onComplete]);

  if (!isVisible) return null;

  const isPositive = change > 0;
  const color = isPositive ? '#10b981' : '#ef4444';
  const sign = isPositive ? '+' : '';

  return (
    <Box
      sx={{
        position: 'absolute',
        top: -10,
        right: -10,
        zIndex: 1000,
        animation: `${floatUp} 2s ease-out`,
        pointerEvents: 'none',
      }}
    >
      <Typography
        sx={{
          fontSize: '18px',
          fontWeight: 800,
          color: color,
          textShadow: `0 0 10px ${color}80, 0 0 20px ${color}40`,
          fontFamily: 'monospace',
        }}
      >
        {sign}{change}
      </Typography>
    </Box>
  );
};
