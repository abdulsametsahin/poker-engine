import React, { useEffect, useState } from 'react';
import { Box, Stack, Typography, Fade, Modal, Chip } from '@mui/material';
import { EmojiEvents } from '@mui/icons-material';
import { PlayingCard } from '../game/PlayingCard';
import { COLORS, RADIUS, TRANSITIONS } from '../../constants';
import { WinnerInfo } from '../../types';

interface HandResultsModalProps {
  winners: WinnerInfo[];
  pot?: number;
  currentUserId?: string;
  duration?: number; // Duration in ms before auto-close
  onClose: () => void;
  show: boolean;
}

const HandResultsModal: React.FC<HandResultsModalProps> = ({
  winners,
  pot,
  currentUserId,
  duration = 5000,
  onClose,
  show,
}) => {
  const [countdown, setCountdown] = useState(Math.floor(duration / 1000));

  useEffect(() => {
    if (show && winners && winners.length > 0) {
      setCountdown(Math.floor(duration / 1000));

      // Countdown timer
      const countdownInterval = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) {
            clearInterval(countdownInterval);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);

      // Auto hide after duration
      const timer = setTimeout(() => {
        onClose();
      }, duration);

      return () => {
        clearTimeout(timer);
        clearInterval(countdownInterval);
      };
    }
  }, [show, winners, duration, onClose]);

  const handleClose = () => {
    onClose();
  };

  // Convert card object to string format
  const cardToString = (card: any): string => {
    if (typeof card === 'string') {
      return card;
    }
    return `${card.rank}${card.suit}`;
  };

  if (!winners || winners.length === 0) return null;

  const isMultipleWinners = winners.length > 1;
  const totalWinnings = winners.reduce((sum, w) => sum + w.amount, 0);
  const isCurrentUserWinner = winners.some(w => w.playerId === currentUserId);

  return (
    <Modal
      open={show}
      onClose={handleClose}
      onClick={handleClose}
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backdropFilter: 'blur(4px)',
        backgroundColor: 'rgba(0, 0, 0, 0.7)',
      }}
    >
      <Fade in={show} timeout={300}>
        <Box
          onClick={(e) => e.stopPropagation()}
          sx={{
            maxWidth: 600,
            width: '90%',
            borderRadius: RADIUS.lg,
            background: `linear-gradient(135deg, ${COLORS.background.paper} 0%, ${COLORS.background.tertiary} 100%)`,
            border: `2px solid ${isCurrentUserWinner ? COLORS.success.main : COLORS.primary.main}`,
            boxShadow: `0 8px 32px rgba(0, 0, 0, 0.6), 0 0 20px ${isCurrentUserWinner ? COLORS.success.glow : COLORS.primary.glow}`,
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          {/* Subtle background */}
          <Box
            sx={{
              position: 'absolute',
              inset: 0,
              background: isCurrentUserWinner
                ? `radial-gradient(circle at 50% 0%, ${COLORS.success.main}15 0%, transparent 50%)`
                : `radial-gradient(circle at 50% 0%, ${COLORS.primary.main}10 0%, transparent 50%)`,
              pointerEvents: 'none',
            }}
          />

          <Stack spacing={2} sx={{ p: 3, position: 'relative', zIndex: 1 }}>
            {/* Header */}
            <Stack direction="row" spacing={2} alignItems="center" justifyContent="center">
              <EmojiEvents
                sx={{
                  fontSize: 36,
                  color: isCurrentUserWinner ? COLORS.success.main : COLORS.primary.main,
                  animation: isCurrentUserWinner ? 'pulse 1.5s ease-in-out infinite' : 'none',
                  '@keyframes pulse': {
                    '0%, 100%': {
                      transform: 'scale(1)',
                    },
                    '50%': {
                      transform: 'scale(1.1)',
                    },
                  },
                }}
              />
              <Typography
                variant="h5"
                sx={{
                  fontWeight: 800,
                  color: isCurrentUserWinner ? COLORS.success.main : COLORS.primary.main,
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                {isMultipleWinners ? 'Split Pot!' : isCurrentUserWinner ? '🎉 You Won!' : 'Hand Complete'}
              </Typography>
            </Stack>

            {/* Pot amount */}
            {pot !== undefined && (
              <Typography
                variant="body1"
                sx={{
                  textAlign: 'center',
                  color: COLORS.text.secondary,
                  fontFamily: 'monospace',
                  fontSize: 16,
                }}
              >
                Pot: <span style={{ color: COLORS.success.main, fontWeight: 700 }}>${totalWinnings}</span>
              </Typography>
            )}

            {/* Winners */}
            <Stack spacing={1.5}>
              {winners.map((winner, idx) => {
                const isThisUserWinner = winner.playerId === currentUserId;
                return (
                  <Box
                    key={idx}
                    sx={{
                      borderRadius: RADIUS.md,
                      background: isThisUserWinner
                        ? `${COLORS.success.main}20`
                        : `${COLORS.background.secondary}90`,
                      border: `2px solid ${isThisUserWinner ? COLORS.success.main : COLORS.border.main}`,
                      p: 2,
                      transition: TRANSITIONS.normal,
                      boxShadow: isThisUserWinner ? `0 0 15px ${COLORS.success.glow}` : 'none',
                    }}
                  >
                    <Stack spacing={1.5}>
                      {/* Player info */}
                      <Stack direction="row" justifyContent="space-between" alignItems="center">
                        <Typography
                          variant="body1"
                          sx={{
                            fontWeight: 700,
                            color: isThisUserWinner ? COLORS.success.main : COLORS.text.primary,
                            fontSize: 16,
                          }}
                        >
                          {isThisUserWinner ? '🏆 You' : (winner.username || winner.playerId.slice(0, 12))}
                        </Typography>

                        <Chip
                          label={`+$${winner.amount}`}
                          sx={{
                            background: COLORS.success.main,
                            color: COLORS.text.primary,
                            fontWeight: 700,
                            fontSize: 14,
                            height: 28,
                          }}
                        />
                      </Stack>

                      {/* Hand rank */}
                      {winner.handRank && (
                        <Typography
                          variant="body2"
                          sx={{
                            textAlign: 'center',
                            color: COLORS.accent.main,
                            fontWeight: 600,
                            textTransform: 'uppercase',
                            fontSize: 13,
                            letterSpacing: '0.05em',
                          }}
                        >
                          {winner.handRank}
                        </Typography>
                      )}

                      {/* Winning cards */}
                      {winner.handCards && winner.handCards.length > 0 && (
                        <Stack
                          direction="row"
                          spacing={0.75}
                          justifyContent="center"
                          flexWrap="wrap"
                        >
                          {winner.handCards.map((card, cardIdx) => (
                            <PlayingCard
                              key={cardIdx}
                              card={cardToString(card)}
                              size="small"
                              highlight={isThisUserWinner}
                            />
                          ))}
                        </Stack>
                      )}
                    </Stack>
                  </Box>
                );
              })}
            </Stack>

            {/* Auto-close hint with countdown */}
            <Typography
              variant="caption"
              sx={{
                textAlign: 'center',
                color: COLORS.text.secondary,
                fontSize: 11,
                mt: 0.5,
              }}
            >
              Click to dismiss • Auto-closing in {countdown}s
            </Typography>
          </Stack>
        </Box>
      </Fade>
    </Modal>
  );
};

export default HandResultsModal;
