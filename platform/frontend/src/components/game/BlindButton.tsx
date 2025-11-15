import React from 'react';
import { Box, Typography } from '@mui/material';
import { TRANSITIONS, COLORS } from '../../constants';

interface BlindButtonProps {
  type: 'SB' | 'BB';
  position: {
    left: string;
    top: string;
  };
}

export const BlindButton: React.FC<BlindButtonProps> = ({ type, position }) => {
  const isSB = type === 'SB';
  const bgColor = isSB ? COLORS.secondary.main : COLORS.info.main;
  const bgDark = isSB ? COLORS.secondary.dark : COLORS.info.dark;
  const glowColor = isSB ? COLORS.secondary.glow : COLORS.info.glow;

  return (
    <Box
      sx={{
        position: 'absolute',
        left: position.left,
        top: position.top,
        transform: 'translate(-50%, -50%)',
        width: 36,
        height: 36,
        borderRadius: '50%',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 25,
        transition: `all ${TRANSITIONS.slow} cubic-bezier(0.4, 0, 0.2, 1)`,
        filter: 'drop-shadow(0 4px 8px rgba(0, 0, 0, 0.5))',

        // Poker chip design with gradient
        background: `
          radial-gradient(circle at 30% 30%, rgba(255, 255, 255, 0.3), transparent 50%),
          linear-gradient(145deg, ${bgColor} 0%, ${bgDark} 100%)
        `,
        border: '2px solid',
        borderColor: 'rgba(255, 255, 255, 0.3)',
        boxShadow: `
          0 0 0 1px rgba(0, 0, 0, 0.5),
          0 2px 8px rgba(0, 0, 0, 0.4),
          0 4px 16px ${glowColor}40,
          inset 0 1px 3px rgba(255, 255, 255, 0.3),
          inset 0 -1px 4px rgba(0, 0, 0, 0.3)
        `,

        // Inner circle design
        '&::before': {
          content: '""',
          position: 'absolute',
          inset: 6,
          borderRadius: '50%',
          background: `
            radial-gradient(circle at 35% 35%, rgba(255, 255, 255, 0.15), transparent 60%),
            linear-gradient(135deg, rgba(0, 0, 0, 0.3) 0%, rgba(0, 0, 0, 0.5) 100%)
          `,
          border: '1px solid rgba(0, 0, 0, 0.4)',
          boxShadow: `
            inset 0 1px 3px rgba(0, 0, 0, 0.5),
            inset 0 -1px 1px rgba(255, 255, 255, 0.1)
          `,
        },

        // Chip edge notches pattern
        '&::after': {
          content: '""',
          position: 'absolute',
          inset: 0,
          borderRadius: '50%',
          background: `
            repeating-conic-gradient(
              from 0deg,
              rgba(255, 255, 255, 0.4) 0deg 9deg,
              transparent 9deg 15deg
            )
          `,
          mask: 'radial-gradient(circle, transparent 70%, black 70%, black 85%, transparent 85%)',
          WebkitMask: 'radial-gradient(circle, transparent 70%, black 70%, black 85%, transparent 85%)',
          opacity: 0.8,
        },

        // Hover effect
        '&:hover': {
          transform: 'translate(-50%, -50%) scale(1.1)',
          filter: 'drop-shadow(0 6px 12px rgba(0, 0, 0, 0.6))',
        },

        // Subtle breathing animation
        animation: 'blindBreath 2.5s ease-in-out infinite',

        '@keyframes blindBreath': {
          '0%, 100%': {
            transform: 'translate(-50%, -50%) scale(1)',
          },
          '50%': {
            transform: 'translate(-50%, -50%) scale(1.03)',
          },
        },
      }}
    >
      {/* Blind label with premium styling */}
      <Typography
        sx={{
          fontSize: '0.75rem',
          fontWeight: 900,
          color: 'white',
          textShadow: `
            0 1px 3px rgba(0, 0, 0, 0.8),
            0 0 4px ${glowColor}60
          `,
          letterSpacing: '0.03em',
          zIndex: 2,
          position: 'relative',
          fontFamily: 'monospace',
          userSelect: 'none',
        }}
      >
        {type}
      </Typography>
    </Box>
  );
};
