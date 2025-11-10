import React, { useEffect, useState } from 'react';
import { Box, Stack, Typography, Fade, Slide, Chip } from '@mui/material';
import { EmojiEvents } from '@mui/icons-material';
import { PlayingCard } from '../game/PlayingCard';
import { COLORS, RADIUS, TRANSITIONS, GAME } from '../../constants';
import { WinnerInfo } from '../../types';

interface HandCompleteDisplayProps {
  winners: WinnerInfo[];
  pot?: number;
  currentUserId?: string;
  onClose: () => void;
}

const HandCompleteDisplay: React.FC<HandCompleteDisplayProps> = ({
  winners,
  pot,
  currentUserId,
  onClose,
}) => {
  const [show, setShow] = useState(false);
  const [countdown, setCountdown] = useState(Math.floor(GAME.HAND_COMPLETE_DELAY / 1000));

  useEffect(() => {
    if (winners && winners.length > 0) {
      setShow(true);
      setCountdown(Math.floor(GAME.HAND_COMPLETE_DELAY / 1000));

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

      // Auto hide after HAND_COMPLETE_DELAY
      const timer = setTimeout(() => {
        handleClose();
      }, GAME.HAND_COMPLETE_DELAY);

      return () => {
        clearTimeout(timer);
        clearInterval(countdownInterval);
      };
    }
  }, [winners]);

  const handleClose = () => {
    setShow(false);
    setTimeout(onClose, 300);
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
  const isWinner = winners.some(w => w.playerId === currentUserId);

  return (
    <Slide direction="left" in={show} timeout={300}>
      <Box
        sx={{
          position: 'fixed',
          right: 20,
          top: '50%',
          transform: 'translateY(-50%)',
          zIndex: 1000,
          maxWidth: 320,
          width: '90%',
        }}
      >
        <Fade in={show} timeout={300}>
          <Box
            onClick={handleClose}
            sx={{
              borderRadius: RADIUS.lg,
              background: `linear-gradient(135deg, ${COLORS.background.paper}f5 0%, ${COLORS.background.tertiary}f5 100%)`,
              border: `2px solid ${isWinner ? COLORS.success.main : COLORS.primary.main}`,
              boxShadow: `0 8px 32px rgba(0, 0, 0, 0.6), 0 0 20px ${isWinner ? COLORS.success.glow : COLORS.primary.glow}`,
              overflow: 'hidden',
              cursor: 'pointer',
              transition: TRANSITIONS.normal,
              '&:hover': {
                transform: 'scale(1.02)',
              },
            }}
          >
            {/* Subtle background */}
            <Box
              sx={{
                position: 'absolute',
                inset: 0,
                background: isWinner
                  ? `radial-gradient(circle at 50% 0%, ${COLORS.success.main}15 0%, transparent 50%)`
                  : `radial-gradient(circle at 50% 0%, ${COLORS.primary.main}10 0%, transparent 50%)`,
                pointerEvents: 'none',
              }}
            />

            <Stack spacing={1.5} sx={{ p: 2.5, position: 'relative', zIndex: 1 }}>
              {/* Header */}
              <Stack direction="row" spacing={1.5} alignItems="center" justifyContent="center">
                <EmojiEvents
                  sx={{
                    fontSize: 24,
                    color: isWinner ? COLORS.success.main : COLORS.primary.main,
                  }}
                />
                <Typography
                  variant="h6"
                  sx={{
                    fontWeight: 800,
                    color: isWinner ? COLORS.success.main : COLORS.primary.main,
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                    fontSize: 16,
                  }}
                >
                  {isMultipleWinners ? 'Split Pot!' : 'Hand Complete'}
                </Typography>
              </Stack>

              {/* Pot amount */}
              {pot !== undefined && (
                <Typography
                  variant="body2"
                  sx={{
                    textAlign: 'center',
                    color: COLORS.text.secondary,
                    fontFamily: 'monospace',
                    fontSize: 13,
                  }}
                >
                  Pot: <span style={{ color: COLORS.success.main, fontWeight: 700 }}>${totalWinnings}</span>
                </Typography>
              )}

              {/* Winners */}
              <Stack spacing={1}>
                {winners.map((winner, idx) => {
                  const isCurrentUserWinner = winner.playerId === currentUserId;
                  return (
                    <Box
                      key={idx}
                      sx={{
                        borderRadius: RADIUS.md,
                        background: isCurrentUserWinner
                          ? `${COLORS.success.main}15`
                          : `${COLORS.background.secondary}90`,
                        border: `1px solid ${isCurrentUserWinner ? COLORS.success.main : COLORS.border.main}`,
                        p: 1.5,
                        transition: TRANSITIONS.normal,
                      }}
                    >
                      <Stack spacing={1}>
                        {/* Player info */}
                        <Stack direction="row" justifyContent="space-between" alignItems="center">
                          <Typography
                            variant="body2"
                            sx={{
                              fontWeight: 600,
                              color: isCurrentUserWinner ? COLORS.success.main : COLORS.text.primary,
                              fontSize: 13,
                            }}
                          >
                            {isCurrentUserWinner ? '🎉 You Won!' : (winner.username || winner.playerId.slice(0, 12))}
                          </Typography>

                          <Chip
                            label={`+$${winner.amount}`}
                            sx={{
                              background: COLORS.success.main,
                              color: COLORS.text.primary,
                              fontWeight: 700,
                              fontSize: 11,
                              height: 22,
                              px: 1,
                            }}
                          />
                        </Stack>

                        {/* Hand rank */}
                        <Typography
                          variant="caption"
                          sx={{
                            textAlign: 'center',
                            color: COLORS.accent.main,
                            fontWeight: 600,
                            textTransform: 'uppercase',
                            fontSize: 11,
                            letterSpacing: '0.05em',
                          }}
                        >
                          {winner.handRank}
                        </Typography>

                        {/* Winning cards */}
                        {winner.handCards && winner.handCards.length > 0 && (
                          <Stack
                            direction="row"
                            spacing={0.5}
                            justifyContent="center"
                            flexWrap="wrap"
                          >
                            {winner.handCards.map((card, cardIdx) => (
                              <PlayingCard
                                key={cardIdx}
                                card={cardToString(card)}
                                size="small"
                                highlight={isCurrentUserWinner}
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
                  fontSize: 10,
                  mt: 0.5,
                }}
              >
                Click to dismiss • Auto-closing in {countdown}s
              </Typography>
            </Stack>
          </Box>
        </Fade>
      </Box>
    </Slide>
  );
};

export default HandCompleteDisplay;
