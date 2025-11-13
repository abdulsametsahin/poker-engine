import React from 'react';
import { Box, Typography } from '@mui/material';
import { COLORS, TRANSITIONS } from '../../constants';

interface DealerButtonProps {
  position: {
    left: string;
    top: string;
  };
}

export const DealerButton: React.FC<DealerButtonProps> = ({ position }) => {
  return (
    <Box
      sx={{
        position: 'absolute',
        left: position.left,
        top: position.top,
        transform: 'translate(-50%, -50%)',
        width: 42,
        height: 42,
        borderRadius: '50%',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 25,
        transition: `all ${TRANSITIONS.slow} cubic-bezier(0.4, 0, 0.2, 1)`,
        filter: 'drop-shadow(0 6px 12px rgba(0, 0, 0, 0.5))',

        // Premium poker chip design with metallic gold rim
        background: `
          radial-gradient(circle at 30% 30%, rgba(255, 255, 255, 0.4), transparent 50%),
          linear-gradient(145deg, #fbbf24 0%, #f59e0b 50%, #d97706 100%)
        `,
        border: '3px solid',
        borderColor: '#fef3c7',
        boxShadow: `
          0 0 0 1.5px #92400e,
          0 3px 10px rgba(0, 0, 0, 0.5),
          0 6px 20px rgba(251, 191, 36, 0.4),
          inset 0 2px 5px rgba(255, 255, 255, 0.4),
          inset 0 -2px 6px rgba(0, 0, 0, 0.3)
        `,

        // Inner circle design
        '&::before': {
          content: '""',
          position: 'absolute',
          inset: 8,
          borderRadius: '50%',
          background: `
            radial-gradient(circle at 35% 35%, rgba(255, 255, 255, 0.2), transparent 60%),
            linear-gradient(135deg, #1e293b 0%, #0f172a 100%)
          `,
          border: '1.5px solid #451a03',
          boxShadow: `
            inset 0 2px 5px rgba(0, 0, 0, 0.6),
            inset 0 -1px 2px rgba(255, 255, 255, 0.1),
            0 0 10px rgba(0, 0, 0, 0.3)
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
              #fef3c7 0deg 8deg,
              transparent 8deg 12deg
            )
          `,
          mask: 'radial-gradient(circle, transparent 68%, black 68%, black 82%, transparent 82%)',
          WebkitMask: 'radial-gradient(circle, transparent 68%, black 68%, black 82%, transparent 82%)',
          opacity: 0.9,
        },

        // Hover effect
        '&:hover': {
          transform: 'translate(-50%, -50%) scale(1.1) rotate(5deg)',
          filter: 'drop-shadow(0 8px 16px rgba(0, 0, 0, 0.6))',
        },

        // Breathing animation
        animation: 'dealerBreath 3s ease-in-out infinite',

        '@keyframes dealerBreath': {
          '0%, 100%': {
            transform: 'translate(-50%, -50%) scale(1)',
            boxShadow: `
              0 0 0 1.5px #92400e,
              0 3px 10px rgba(0, 0, 0, 0.5),
              0 6px 20px rgba(251, 191, 36, 0.4),
              inset 0 2px 5px rgba(255, 255, 255, 0.4),
              inset 0 -2px 6px rgba(0, 0, 0, 0.3)
            `,
          },
          '50%': {
            transform: 'translate(-50%, -50%) scale(1.05)',
            boxShadow: `
              0 0 0 1.5px #92400e,
              0 5px 14px rgba(0, 0, 0, 0.6),
              0 10px 28px rgba(251, 191, 36, 0.6),
              inset 0 2px 5px rgba(255, 255, 255, 0.5),
              inset 0 -2px 6px rgba(0, 0, 0, 0.4)
            `,
          },
        },

        // Subtle rotating shine effect
        '& > .shine': {
          position: 'absolute',
          inset: 0,
          borderRadius: '50%',
          background: 'linear-gradient(135deg, transparent 0%, rgba(255, 255, 255, 0.3) 50%, transparent 100%)',
          animation: 'rotateShine 4s linear infinite',
        },

        '@keyframes rotateShine': {
          '0%': { transform: 'rotate(0deg)' },
          '100%': { transform: 'rotate(360deg)' },
        },
      }}
    >
      {/* Rotating shine effect */}
      <Box className="shine" />

      {/* Dealer letter with premium styling */}
      <Typography
        sx={{
          fontSize: '1.2rem',
          fontWeight: 900,
          color: '#fef3c7',
          textShadow: `
            0 2px 4px rgba(0, 0, 0, 0.8),
            0 0 6px rgba(251, 191, 36, 0.6),
            0 1px 0 rgba(255, 255, 255, 0.4)
          `,
          letterSpacing: '0.05em',
          zIndex: 2,
          position: 'relative',
          fontFamily: 'serif',
          userSelect: 'none',
        }}
      >
        D
      </Typography>
    </Box>
  );
};
