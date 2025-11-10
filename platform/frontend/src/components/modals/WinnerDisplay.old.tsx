import React, { useEffect, useState, memo } from 'react';
import { Box, Modal, Fade, Stack, Typography } from '@mui/material';
import { EmojiEvents, Stars } from '@mui/icons-material';
import { PlayingCard } from '../game/PlayingCard';
import { Avatar } from '../common/Avatar';
import { Badge } from '../common/Badge';
import { COLORS, RADIUS, SPACING } from '../../constants';
import { formatUsername, formatChips } from '../../utils';

interface CardObject {
  rank: string;
  suit: string;
}

interface Winner {
  playerId: string;
  amount: number;
  handRank: string;
  handCards: (string | CardObject)[];
}

interface WinnerDisplayProps {
  winners: Winner[];
  onClose: () => void;
}

const Confetti: React.FC = () => {
  const pieces = Array.from({ length: 50 }, (_, i) => ({
    id: i,
    left: Math.random() * 100,
    delay: Math.random() * 2,
    duration: 2 + Math.random() * 2,
  }));

  return (
    <Box sx={{ position: 'absolute', inset: 0, pointerEvents: 'none', overflow: 'hidden' }}>
      {pieces.map((piece) => (
        <Box
          key={piece.id}
          sx={{
            position: 'absolute',
            left: `${piece.left}%`,
            top: -20,
            width: 10,
            height: 10,
            background: `hsl(${Math.random() * 360}, 70%, 60%)`,
            opacity: 0.8,
            borderRadius: '50%',
            '@keyframes fall': {
              '0%': {
                transform: 'translateY(0) rotate(0deg)',
                opacity: 1,
              },
              '100%': {
                transform: `translateY(100vh) rotate(${360 + Math.random() * 360}deg)`,
                opacity: 0,
              },
            },
            animation: `fall ${piece.duration}s ease-in ${piece.delay}s infinite`,
          }}
        />
      ))}
    </Box>
  );
};

export const WinnerDisplay: React.FC<WinnerDisplayProps> = memo(({ winners, onClose }) => {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    if (winners && winners.length > 0) {
      setOpen(true);
      // Auto close after 5 seconds
      const timer = setTimeout(() => {
        handleClose();
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [winners]);

  const handleClose = () => {
    setOpen(false);
    setTimeout(onClose, 300);
  };

  const cardToString = (card: string | CardObject): string => {
    if (typeof card === 'string') return card;
    return `${card.rank}${card.suit}`;
  };

  if (!winners || winners.length === 0) return null;

  const mainWinner = winners[0];
  const bigWin = mainWinner.amount >= 500;

  return (
    <Modal
      open={open}
      onClose={handleClose}
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backdropFilter: 'blur(4px)',
      }}
    >
      <Fade in={open}>
        <Box
          onClick={handleClose}
          sx={{
            position: 'relative',
            width: '90%',
            maxWidth: 600,
            maxHeight: '90vh',
            overflow: 'auto',
            borderRadius: RADIUS.lg,
            background: `linear-gradient(135deg, rgba(0, 0, 0, 0.95) 0%, rgba(20, 20, 20, 0.95) 100%)`,
            border: `3px solid ${COLORS.accent.main}`,
            boxShadow: `0 0 40px ${COLORS.accent.glow}, 0 20px 60px rgba(0, 0, 0, 0.8)`,
            p: 4,
            cursor: 'pointer',
            '&::before': {
              content: '""',
              position: 'absolute',
              inset: 0,
              borderRadius: RADIUS.lg,
              background: `radial-gradient(circle at 50% 0%, ${COLORS.accent.main}20 0%, transparent 70%)`,
              pointerEvents: 'none',
            },
          }}
        >
          {/* Confetti animation for big wins */}
          {bigWin && <Confetti />}

          <Stack spacing={3} alignItems="center" sx={{ position: 'relative', zIndex: 1 }}>
            {/* Trophy icon with animation */}
            <Box
              sx={{
                '@keyframes trophy-bounce': {
                  '0%, 100%': {
                    transform: 'translateY(0) scale(1)',
                  },
                  '50%': {
                    transform: 'translateY(-20px) scale(1.1)',
                  },
                },
                animation: 'trophy-bounce 1s ease-in-out 3',
              }}
            >
              <EmojiEvents
                sx={{
                  fontSize: 100,
                  color: COLORS.accent.main,
                  filter: `drop-shadow(0 0 20px ${COLORS.accent.glow})`,
                }}
              />
            </Box>

            {/* Winner(s) Title */}
            <Typography
              variant="h2"
              sx={{
                color: COLORS.accent.main,
                fontWeight: 900,
                fontSize: { xs: '32px', sm: '40px' },
                textAlign: 'center',
                textShadow: `0 0 20px ${COLORS.accent.glow}`,
                '@keyframes glow-text': {
                  '0%, 100%': {
                    textShadow: `0 0 20px ${COLORS.accent.glow}`,
                  },
                  '50%': {
                    textShadow: `0 0 30px ${COLORS.accent.glow}, 0 0 40px ${COLORS.accent.glow}`,
                  },
                },
                animation: 'glow-text 2s ease-in-out infinite',
              }}
            >
              {winners.length === 1 ? '🎉 WINNER! 🎉' : '🎉 WINNERS! 🎉'}
            </Typography>

            {/* Winner cards */}
            {winners.map((winner, idx) => (
              <Box
                key={idx}
                sx={{
                  width: '100%',
                  p: 3,
                  borderRadius: RADIUS.md,
                  background: idx === 0
                    ? `linear-gradient(135deg, ${COLORS.accent.main}15 0%, ${COLORS.warning.main}15 100%)`
                    : 'rgba(255, 255, 255, 0.05)',
                  border: idx === 0
                    ? `2px solid ${COLORS.accent.main}`
                    : `1px solid ${COLORS.border.main}`,
                  backdropFilter: 'blur(10px)',
                }}
              >
                <Stack spacing={2}>
                  {/* Player info */}
                  <Stack direction="row" alignItems="center" justifyContent="space-between" flexWrap="wrap" gap={2}>
                    <Stack direction="row" spacing={2} alignItems="center">
                      <Avatar
                        username={winner.playerId}
                        size="large"
                      />
                      <Box>
                        <Typography
                          variant="h5"
                          sx={{
                            color: COLORS.text.primary,
                            fontWeight: 700,
                          }}
                        >
                          {formatUsername(winner.playerId.slice(0, 12))}
                        </Typography>
                        {idx === 0 && (
                          <Badge variant="warning" size="small">
                            1ST PLACE
                          </Badge>
                        )}
                      </Box>
                    </Stack>

                    <Box
                      sx={{
                        px: 3,
                        py: 1.5,
                        borderRadius: RADIUS.sm,
                        background: `linear-gradient(135deg, ${COLORS.success.main} 0%, ${COLORS.success.dark} 100%)`,
                        boxShadow: `0 4px 12px ${COLORS.success.glow}`,
                      }}
                    >
                      <Typography
                        variant="h4"
                        sx={{
                          color: COLORS.text.primary,
                          fontWeight: 900,
                          fontFamily: 'monospace',
                        }}
                      >
                        +{formatChips(winner.amount)}
                      </Typography>
                    </Box>
                  </Stack>

                  {/* Hand rank */}
                  <Box sx={{ textAlign: 'center' }}>
                    <Stack direction="row" spacing={1} alignItems="center" justifyContent="center">
                      <Stars sx={{ color: COLORS.accent.main, fontSize: 20 }} />
                      <Typography
                        variant="h5"
                        sx={{
                          color: COLORS.accent.main,
                          fontWeight: 700,
                          textTransform: 'uppercase',
                        }}
                      >
                        {winner.handRank}
                      </Typography>
                      <Stars sx={{ color: COLORS.accent.main, fontSize: 20 }} />
                    </Stack>
                  </Box>

                  {/* Winning cards */}
                  {winner.handCards && winner.handCards.length > 0 && (
                    <Stack direction="row" spacing={1} justifyContent="center" flexWrap="wrap">
                      {winner.handCards.map((card, cardIdx) => (
                        <PlayingCard
                          key={cardIdx}
                          card={cardToString(card)}
                          size="medium"
                          highlight={true}
                        />
                      ))}
                    </Stack>
                  )}
                </Stack>
              </Box>
            ))}

            {/* Close hint */}
            <Typography
              variant="caption"
              sx={{
                color: COLORS.text.secondary,
                fontSize: '12px',
                mt: 2,
              }}
            >
              Click anywhere to continue
            </Typography>
          </Stack>
        </Box>
      </Fade>
    </Modal>
  );
});

WinnerDisplay.displayName = 'WinnerDisplay';

export default WinnerDisplay;
