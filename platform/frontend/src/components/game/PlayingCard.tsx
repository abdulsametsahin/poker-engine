import React, { memo } from 'react';
import { Box } from '@mui/material';
import { COLORS, RADIUS, GAME, TRANSITIONS } from '../../constants';
import { parseCard, getCardColor } from '../../utils';

interface PlayingCardProps {
  card: string;
  size?: 'small' | 'medium' | 'large';
  faceDown?: boolean;
  highlight?: boolean;
  dealAnimation?: boolean;
}

export const PlayingCard: React.FC<PlayingCardProps> = memo(({
  card,
  size = 'medium',
  faceDown = false,
  highlight = false,
  dealAnimation = false,
}) => {
  const parsedCard = parseCard(card);
  const { width, height, fontSize } = GAME.CARD_SIZES[size];

  if (faceDown) {
    return (
      <Box
        sx={{
          width,
          height,
          borderRadius: RADIUS.sm,
          background: `linear-gradient(135deg, ${COLORS.primary.main} 0%, ${COLORS.primary.dark} 100%)`,
          border: `2px solid ${COLORS.primary.light}`,
          backdropFilter: 'blur(10px)',
          boxShadow: `0 4px 12px rgba(0, 0, 0, 0.4), 0 0 12px ${COLORS.primary.glow}`,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          position: 'relative',
          overflow: 'hidden',
          transition: TRANSITIONS.normal,
          '&:hover': {
            transform: 'translateY(-4px)',
            boxShadow: `0 6px 16px rgba(0, 0, 0, 0.5), 0 0 16px ${COLORS.primary.glow}`,
          },
          ...(dealAnimation && {
            '@keyframes deal-card': {
              '0%': {
                transform: 'translateX(-200px) translateY(-200px) rotate(-180deg)',
                opacity: 0,
              },
              '60%': {
                transform: 'translateX(0) translateY(0) rotate(10deg)',
                opacity: 1,
              },
              '100%': {
                transform: 'translateX(0) translateY(0) rotate(0deg)',
                opacity: 1,
              },
            },
            animation: 'deal-card 0.5s cubic-bezier(0.34, 1.56, 0.64, 1)',
          }),
          '&::before': {
            content: '""',
            position: 'absolute',
            inset: 4,
            borderRadius: RADIUS.sm,
            background: 'repeating-linear-gradient(45deg, transparent, transparent 8px, rgba(255,255,255,0.05) 8px, rgba(255,255,255,0.05) 16px)',
          },
          '&::after': {
            content: '""',
            position: 'absolute',
            inset: '30%',
            borderRadius: '50%',
            background: `radial-gradient(circle, ${COLORS.secondary.main} 0%, transparent 70%)`,
            opacity: 0.3,
          },
        }}
      />
    );
  }

  const suitColor = getCardColor(parsedCard.suit);
  const isRed = suitColor === COLORS.danger.main;

  return (
    <Box
      sx={{
        width,
        height,
        borderRadius: RADIUS.sm,
        background: '#FFFFFF',
        border: highlight
          ? `3px solid ${COLORS.accent.main}`
          : '2px solid rgba(0, 0, 0, 0.1)',
        boxShadow: highlight
          ? `0 6px 16px rgba(0, 0, 0, 0.4), 0 0 20px ${COLORS.accent.glow}`
          : '0 4px 12px rgba(0, 0, 0, 0.3), 0 2px 4px rgba(0, 0, 0, 0.2)',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'space-between',
        p: 0.75,
        position: 'relative',
        transition: TRANSITIONS.normal,
        transform: highlight ? 'translateY(-8px) scale(1.05)' : 'none',
        '&:hover': {
          transform: highlight ? 'translateY(-10px) scale(1.05)' : 'translateY(-4px)',
          boxShadow: highlight
            ? `0 8px 20px rgba(0, 0, 0, 0.5), 0 0 24px ${COLORS.accent.glow}`
            : '0 6px 16px rgba(0, 0, 0, 0.4)',
        },
        ...(dealAnimation && {
          '@keyframes deal-card': {
            '0%': {
              transform: 'translateX(-200px) translateY(-200px) rotate(-180deg)',
              opacity: 0,
            },
            '60%': {
              transform: 'translateX(0) translateY(0) rotate(10deg)',
              opacity: 1,
            },
            '100%': {
              transform: 'translateX(0) translateY(0) rotate(0deg)',
              opacity: 1,
            },
          },
          animation: 'deal-card 0.5s cubic-bezier(0.34, 1.56, 0.64, 1)',
        }),
        '&::before': {
          content: '""',
          position: 'absolute',
          inset: 3,
          border: '1px solid rgba(0, 0, 0, 0.05)',
          borderRadius: RADIUS.sm,
          pointerEvents: 'none',
        },
      }}
    >
      {/* Top rank */}
      <Box
        sx={{
          fontSize: size === 'small' ? fontSize * 0.9 : fontSize,
          fontWeight: 900,
          color: suitColor,
          lineHeight: 1,
          textShadow: '0 1px 2px rgba(0, 0, 0, 0.1)',
        }}
      >
        {parsedCard.rank}
      </Box>

      {/* Center suit */}
      <Box
        sx={{
          fontSize: size === 'small' ? fontSize * 1.3 : fontSize * 1.5,
          color: suitColor,
          lineHeight: 1,
          filter: 'drop-shadow(0 2px 3px rgba(0, 0, 0, 0.15))',
        }}
      >
        {parsedCard.suit}
      </Box>

      {/* Bottom rank */}
      <Box
        sx={{
          fontSize: size === 'small' ? fontSize * 0.9 : fontSize,
          fontWeight: 900,
          color: suitColor,
          lineHeight: 1,
          textShadow: '0 1px 2px rgba(0, 0, 0, 0.1)',
          transform: 'rotate(180deg)',
        }}
      >
        {parsedCard.rank}
      </Box>
    </Box>
  );
});

PlayingCard.displayName = 'PlayingCard';

export default PlayingCard;
