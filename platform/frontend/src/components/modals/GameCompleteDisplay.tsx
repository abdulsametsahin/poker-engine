import React, { useEffect, useState, memo } from 'react';
import { Box, Modal, Fade, Stack, Typography } from '@mui/material';
import { EmojiEvents, Home, Refresh } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { Avatar } from '../common/Avatar';
import { Button } from '../common/Button';
import { COLORS, RADIUS } from '../../constants';
import { formatUsername, formatChipsFull } from '../../utils';

interface GameCompleteDisplayProps {
  winner: string;
  winnerName?: string;
  finalChips: number;
  totalPlayers: number;
  message: string;
  currentUserId?: string;
}

export const GameCompleteDisplay: React.FC<GameCompleteDisplayProps> = memo(({
  winner,
  winnerName,
  finalChips,
  totalPlayers,
  message,
  currentUserId,
}) => {
  const [open, setOpen] = useState(true);
  const navigate = useNavigate();
  const isWinner = currentUserId === winner;

  useEffect(() => {
    setOpen(true);
  }, []);

  const handleReturnToLobby = () => {
    navigate('/lobby');
  };

  return (
    <Modal
      open={open}
      onClose={handleReturnToLobby}
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backdropFilter: 'blur(8px)',
      }}
    >
      <Fade in={open}>
        <Box
          sx={{
            position: 'relative',
            width: '90%',
            maxWidth: 700,
            maxHeight: '90vh',
            overflow: 'auto',
            borderRadius: RADIUS.lg,
            background: isWinner
              ? `linear-gradient(135deg, rgba(16, 185, 129, 0.1) 0%, rgba(0, 0, 0, 0.95) 100%)`
              : `linear-gradient(135deg, rgba(0, 0, 0, 0.95) 0%, rgba(20, 20, 20, 0.95) 100%)`,
            border: `4px solid ${isWinner ? COLORS.success.main : COLORS.warning.main}`,
            boxShadow: isWinner
              ? `0 0 60px ${COLORS.success.glow}, 0 20px 60px rgba(0, 0, 0, 0.8)`
              : `0 0 60px ${COLORS.warning.glow}, 0 20px 60px rgba(0, 0, 0, 0.8)`,
            p: { xs: 3, sm: 5 },
            textAlign: 'center',
            '&::before': {
              content: '""',
              position: 'absolute',
              inset: 0,
              borderRadius: RADIUS.lg,
              background: isWinner
                ? `radial-gradient(circle at 50% 0%, ${COLORS.success.main}20 0%, transparent 70%)`
                : `radial-gradient(circle at 50% 0%, ${COLORS.warning.main}20 0%, transparent 70%)`,
              pointerEvents: 'none',
            },
          }}
        >
          {/* Confetti for winner */}
          {isWinner && (
            <Box sx={{ position: 'absolute', inset: 0, pointerEvents: 'none', overflow: 'hidden' }}>
              {Array.from({ length: 30 }).map((_, i) => (
                <Box
                  key={i}
                  sx={{
                    position: 'absolute',
                    left: `${Math.random() * 100}%`,
                    top: -20,
                    width: 12,
                    height: 12,
                    background: `hsl(${Math.random() * 360}, 70%, 60%)`,
                    borderRadius: '50%',
                    '@keyframes fall': {
                      to: {
                        transform: `translateY(100vh) rotate(${360 + Math.random() * 360}deg)`,
                        opacity: 0,
                      },
                    },
                    animation: `fall ${2 + Math.random() * 2}s ease-in ${Math.random() * 2}s infinite`,
                  }}
                />
              ))}
            </Box>
          )}

          <Stack spacing={4} alignItems="center" sx={{ position: 'relative', zIndex: 1 }}>
            {/* Trophy animation */}
            <Box
              sx={{
                '@keyframes trophy-bounce': {
                  '0%, 100%': {
                    transform: 'translateY(0) scale(1) rotate(0deg)',
                  },
                  '25%': {
                    transform: 'translateY(-30px) scale(1.1) rotate(-5deg)',
                  },
                  '75%': {
                    transform: 'translateY(-30px) scale(1.1) rotate(5deg)',
                  },
                },
                animation: isWinner ? 'trophy-bounce 2s ease-in-out infinite' : 'none',
              }}
            >
              <EmojiEvents
                sx={{
                  fontSize: { xs: 100, sm: 140 },
                  color: isWinner ? COLORS.success.main : COLORS.warning.main,
                  filter: `drop-shadow(0 0 30px ${isWinner ? COLORS.success.glow : COLORS.warning.glow})`,
                }}
              />
            </Box>

            {/* Game Over message */}
            <Typography
              variant="h2"
              sx={{
                color: isWinner ? COLORS.success.main : COLORS.text.primary,
                fontWeight: 900,
                fontSize: { xs: '32px', sm: '48px' },
                textShadow: `0 4px 12px rgba(0, 0, 0, 0.5)`,
              }}
            >
              {message}
            </Typography>

            {/* Podium / Winner info */}
            <Box
              sx={{
                width: '100%',
                p: { xs: 3, sm: 4 },
                borderRadius: RADIUS.md,
                background: isWinner
                  ? `linear-gradient(135deg, ${COLORS.success.main}20 0%, ${COLORS.success.dark}10 100%)`
                  : 'rgba(255, 255, 255, 0.05)',
                border: `2px solid ${isWinner ? COLORS.success.main : COLORS.border.heavy}`,
                backdropFilter: 'blur(10px)',
              }}
            >
              <Stack spacing={3} alignItems="center">
                {/* Winner avatar */}
                <Avatar username={winnerName || winner} size="large" />

                {/* Winner name */}
                <Typography
                  variant="h4"
                  sx={{
                    color: COLORS.text.primary,
                    fontWeight: 700,
                  }}
                >
                  {isWinner ? '🎉 You Won! 🎉' : `${formatUsername(winnerName || winner.slice(0, 12))} Won!`}
                </Typography>

                {/* Statistics */}
                <Stack
                  direction={{ xs: 'column', sm: 'row' }}
                  spacing={3}
                  sx={{ width: '100%' }}
                  justifyContent="center"
                  alignItems="center"
                >
                  {/* Final chips */}
                  <Box sx={{ textAlign: 'center' }}>
                    <Typography
                      variant="caption"
                      sx={{
                        color: COLORS.text.secondary,
                        fontSize: '12px',
                        fontWeight: 600,
                        letterSpacing: '0.1em',
                      }}
                    >
                      FINAL CHIPS
                    </Typography>
                    <Typography
                      variant="h3"
                      sx={{
                        color: COLORS.success.main,
                        fontWeight: 900,
                        fontSize: { xs: '28px', sm: '36px' },
                        fontFamily: 'monospace',
                        mt: 0.5,
                      }}
                    >
                      {formatChipsFull(finalChips)}
                    </Typography>
                  </Box>

                  {/* Divider */}
                  <Box
                    sx={{
                      width: { xs: 100, sm: 2 },
                      height: { xs: 2, sm: 80 },
                      bgcolor: COLORS.border.main,
                    }}
                  />

                  {/* Players defeated */}
                  <Box sx={{ textAlign: 'center' }}>
                    <Typography
                      variant="caption"
                      sx={{
                        color: COLORS.text.secondary,
                        fontSize: '12px',
                        fontWeight: 600,
                        letterSpacing: '0.1em',
                      }}
                    >
                      PLAYERS DEFEATED
                    </Typography>
                    <Typography
                      variant="h3"
                      sx={{
                        color: COLORS.danger.main,
                        fontWeight: 900,
                        fontSize: { xs: '28px', sm: '36px' },
                        mt: 0.5,
                      }}
                    >
                      {totalPlayers - 1}
                    </Typography>
                  </Box>
                </Stack>

                {/* Encouragement message */}
                {isWinner ? (
                  <Typography
                    variant="body1"
                    sx={{
                      color: COLORS.text.secondary,
                      fontSize: '14px',
                      maxWidth: 400,
                    }}
                  >
                    Congratulations! You've eliminated all opponents and won the game! Your poker skills are impressive!
                  </Typography>
                ) : (
                  <Typography
                    variant="body1"
                    sx={{
                      color: COLORS.text.secondary,
                      fontSize: '14px',
                      maxWidth: 400,
                    }}
                  >
                    Better luck next time! Every game is a learning opportunity. Keep practicing and you'll be winning soon!
                  </Typography>
                )}
              </Stack>
            </Box>

            {/* Action buttons */}
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} sx={{ width: '100%', maxWidth: 400 }}>
              <Button
                variant="primary"
                onClick={handleReturnToLobby}
                fullWidth
                sx={{
                  py: 1.5,
                  fontSize: '16px',
                  fontWeight: 700,
                }}
              >
                <Home sx={{ mr: 1 }} />
                RETURN TO LOBBY
              </Button>

              <Button
                variant="secondary"
                onClick={handleReturnToLobby}
                fullWidth
                sx={{
                  py: 1.5,
                  fontSize: '16px',
                  fontWeight: 700,
                }}
              >
                <Refresh sx={{ mr: 1 }} />
                PLAY AGAIN
              </Button>
            </Stack>
          </Stack>
        </Box>
      </Fade>
    </Modal>
  );
});

GameCompleteDisplay.displayName = 'GameCompleteDisplay';

export default GameCompleteDisplay;
