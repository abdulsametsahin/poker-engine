import React, { useEffect, useState } from 'react';
import { Box, Stack, Typography, Fade, Modal, Chip } from '@mui/material';
import { EmojiEvents, Home, Refresh } from '@mui/icons-material';
import { PlayingCard } from '../game/PlayingCard';
import { Button } from '../common/Button';
import { COLORS, RADIUS, TRANSITIONS, GAME } from '../../constants';
import { WinnerInfo } from '../../types';

interface WinnerDisplayProps {
  winners: WinnerInfo[];
  pot?: number;
  gameComplete?: boolean;
  gameMode?: string;
  onClose: () => void;
  onPlayAgain?: () => void;
  onReturnToLobby?: () => void;
}

const WinnerDisplay: React.FC<WinnerDisplayProps> = ({ 
  winners, 
  pot, 
  gameComplete = false,
  gameMode,
  onClose,
  onPlayAgain,
  onReturnToLobby,
}) => {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    if (winners && winners.length > 0) {
      setOpen(true);
      // Auto close only if not game complete
      if (!gameComplete) {
        const timer = setTimeout(() => {
          handleClose();
        }, GAME.WINNER_MODAL_DURATION);
        return () => clearTimeout(timer);
      }
    }
  }, [winners, gameComplete]);

  const handleClose = () => {
    setOpen(false);
    setTimeout(onClose, 300);
  };

  const handlePlayAgain = () => {
    if (onPlayAgain) {
      onPlayAgain();
    }
    handleClose();
  };

  const handleReturnToLobby = () => {
    if (onReturnToLobby) {
      onReturnToLobby();
    }
    handleClose();
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

  return (
    <Modal
      open={open}
      onClose={gameComplete ? undefined : handleClose}
      onClick={gameComplete ? undefined : handleClose}
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backdropFilter: 'blur(4px)',
        backgroundColor: 'rgba(0, 0, 0, 0.75)',
      }}
    >
      <Fade in={open} timeout={300}>
        <Box
          onClick={(e) => e.stopPropagation()}
          sx={{
            maxWidth: 600,
            width: '90%',
            borderRadius: RADIUS.lg,
            background: `linear-gradient(135deg, ${COLORS.background.paper} 0%, ${COLORS.background.tertiary} 100%)`,
            border: `2px solid ${gameComplete ? COLORS.accent.main : COLORS.success.main}`,
            boxShadow: `0 8px 32px rgba(0, 0, 0, 0.6)`,
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          {/* Subtle background */}
          <Box
            sx={{
              position: 'absolute',
              inset: 0,
              background: gameComplete
                ? `radial-gradient(circle at 50% 0%, ${COLORS.accent.main}10 0%, transparent 50%)`
                : `radial-gradient(circle at 50% 0%, ${COLORS.success.main}10 0%, transparent 50%)`,
              pointerEvents: 'none',
            }}
          />

          <Stack spacing={2} sx={{ p: 3, position: 'relative', zIndex: 1 }}>
            {/* Header */}
            <Stack direction="row" spacing={2} alignItems="center" justifyContent="center">
              <EmojiEvents
                sx={{
                  fontSize: 32,
                  color: gameComplete ? COLORS.accent.main : COLORS.success.main,
                }}
              />
              <Typography
                variant="h5"
                sx={{
                  fontWeight: 800,
                  color: gameComplete ? COLORS.accent.main : COLORS.success.main,
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                }}
              >
                {gameComplete ? '🎮 Game Complete!' : isMultipleWinners ? 'Split Pot!' : 'Winner!'}
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
              {winners.map((winner, idx) => (
                <Box
                  key={idx}
                  sx={{
                    borderRadius: RADIUS.md,
                    background: `${COLORS.background.secondary}90`,
                    border: `1px solid ${COLORS.border.main}`,
                    p: 2,
                    transition: TRANSITIONS.normal,
                  }}
                >
                  <Stack spacing={1.5}>
                    {/* Player info */}
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                      <Typography
                        variant="body1"
                        sx={{
                          fontWeight: 600,
                          color: COLORS.text.primary,
                          fontSize: 16,
                        }}
                      >
                        {winner.username || winner.playerId.slice(0, 12)}
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
                            highlight={!gameComplete}
                          />
                        ))}
                      </Stack>
                    )}
                  </Stack>
                </Box>
              ))}
            </Stack>

            {/* Game complete actions or auto-close hint */}
            {gameComplete ? (
              <Stack direction="row" spacing={1.5} sx={{ mt: 1 }}>
                <Button
                  variant="primary"
                  onClick={handlePlayAgain}
                  sx={{
                    flex: 1,
                    py: 1.25,
                    fontSize: 14,
                    fontWeight: 700,
                  }}
                  startIcon={<Refresh />}
                >
                  Play Again
                </Button>
                <Button
                  variant="secondary"
                  onClick={handleReturnToLobby}
                  sx={{
                    flex: 1,
                    py: 1.25,
                    fontSize: 14,
                    fontWeight: 700,
                  }}
                  startIcon={<Home />}
                >
                  Lobby
                </Button>
              </Stack>
            ) : (
              <Typography
                variant="caption"
                sx={{
                  textAlign: 'center',
                  color: COLORS.text.secondary,
                  fontSize: 11,
                  mt: 0.5,
                }}
              >
                Click anywhere to continue
              </Typography>
            )}
          </Stack>
        </Box>
      </Fade>
    </Modal>
  );
};

export default WinnerDisplay;
