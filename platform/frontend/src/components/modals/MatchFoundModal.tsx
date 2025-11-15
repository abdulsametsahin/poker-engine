import React, { useState, useEffect } from 'react';
import { Box, Typography, LinearProgress, Modal, Fade } from '@mui/material';
import { COLORS, RADIUS, GAME } from '../../constants';

interface MatchFoundModalProps {
  open: boolean;
  gameMode: string;
  startDeadline?: string;
  onCountdownComplete: () => void;
}

export const MatchFoundModal: React.FC<MatchFoundModalProps> = ({
  open,
  gameMode,
  startDeadline,
  onCountdownComplete,
}) => {
  const [countdown, setCountdown] = useState(10);
  const [progress, setProgress] = useState(100);

  useEffect(() => {
    if (!open || !startDeadline) {
      setCountdown(10);
      setProgress(100);
      return;
    }

    // Parse the deadline from the server (ISO string)
    const deadlineTime = new Date(startDeadline).getTime();
    // Get total countdown duration from env or default to 10 seconds
    const totalTime = parseInt(process.env.REACT_APP_MATCHMAKING_COUNTDOWN_SECONDS || '10', 10);

    const updateTimer = () => {
      const now = Date.now();
      const remaining = Math.max(0, deadlineTime - now);
      const remainingSeconds = Math.ceil(remaining / 1000);

      setCountdown(remainingSeconds);
      setProgress((remainingSeconds / totalTime) * 100);

      if (remainingSeconds <= 0) {
        onCountdownComplete();
      }
    };

    // Update immediately
    updateTimer();

    // Update every 100ms for smooth countdown
    const interval = setInterval(updateTimer, 100);

    return () => clearInterval(interval);
  }, [open, startDeadline, onCountdownComplete]);

  const getGameModeName = (mode: string) => {
    switch (mode) {
      case 'heads_up':
        return 'Heads-Up';
      case '3_player':
        return '3-Player';
      case '6_player':
        return '6-Player';
      default:
        return mode;
    }
  };

  return (
    <Modal
      open={open}
      closeAfterTransition
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <Fade in={open}>
        <Box
          sx={{
            position: 'relative',
            width: 480,
            maxWidth: '90vw',
            borderRadius: RADIUS.lg,
            background: `linear-gradient(135deg, rgba(15, 15, 15, 0.98) 0%, rgba(30, 30, 30, 0.98) 100%)`,
            backdropFilter: 'blur(20px)',
            border: `3px solid ${COLORS.success.main}`,
            boxShadow: `
              0 20px 60px rgba(0, 0, 0, 0.9),
              0 0 40px ${COLORS.success.glow},
              inset 0 2px 4px rgba(255, 255, 255, 0.1)
            `,
            p: 4,
            outline: 'none',

            // Pulsing animation
            animation: 'matchPulse 2s ease-in-out infinite',
            '@keyframes matchPulse': {
              '0%, 100%': {
                boxShadow: `
                  0 20px 60px rgba(0, 0, 0, 0.9),
                  0 0 40px ${COLORS.success.glow},
                  inset 0 2px 4px rgba(255, 255, 255, 0.1)
                `,
              },
              '50%': {
                boxShadow: `
                  0 20px 60px rgba(0, 0, 0, 0.9),
                  0 0 60px ${COLORS.success.glow},
                  inset 0 2px 4px rgba(255, 255, 255, 0.1)
                `,
              },
            },
          }}
        >
          {/* Success icon */}
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'center',
              mb: 3,
            }}
          >
            <Box
              sx={{
                width: 80,
                height: 80,
                borderRadius: '50%',
                background: `linear-gradient(135deg, ${COLORS.success.main} 0%, ${COLORS.success.dark} 100%)`,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                boxShadow: `0 8px 24px ${COLORS.success.glow}`,
                animation: 'checkBounce 0.6s ease-out',
                '@keyframes checkBounce': {
                  '0%': {
                    transform: 'scale(0)',
                  },
                  '50%': {
                    transform: 'scale(1.1)',
                  },
                  '100%': {
                    transform: 'scale(1)',
                  },
                },
              }}
            >
              <Typography sx={{ fontSize: '48px' }}>✓</Typography>
            </Box>
          </Box>

          {/* Title */}
          <Typography
            variant="h4"
            sx={{
              textAlign: 'center',
              fontWeight: 900,
              fontSize: '32px',
              background: `linear-gradient(135deg, ${COLORS.success.light} 0%, ${COLORS.success.main} 100%)`,
              backgroundClip: 'text',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              mb: 1,
            }}
          >
            Match Found!
          </Typography>

          {/* Game mode */}
          <Typography
            sx={{
              textAlign: 'center',
              color: COLORS.text.secondary,
              fontSize: '16px',
              fontWeight: 600,
              mb: 4,
            }}
          >
            {getGameModeName(gameMode)} Game
          </Typography>

          {/* Countdown */}
          <Box
            sx={{
              textAlign: 'center',
              mb: 3,
            }}
          >
            <Typography
              sx={{
                fontSize: '72px',
                fontWeight: 900,
                color: COLORS.text.primary,
                lineHeight: 1,
                fontFamily: 'monospace',
                mb: 1,
              }}
            >
              {countdown}
            </Typography>
            <Typography
              sx={{
                color: COLORS.text.secondary,
                fontSize: '14px',
                fontWeight: 600,
              }}
            >
              Starting game...
            </Typography>
          </Box>

          {/* Progress bar */}
          <LinearProgress
            variant="determinate"
            value={progress}
            sx={{
              height: 8,
              borderRadius: RADIUS.sm,
              backgroundColor: 'rgba(16, 185, 129, 0.2)',
              '& .MuiLinearProgress-bar': {
                borderRadius: RADIUS.sm,
                background: `linear-gradient(90deg, ${COLORS.success.main} 0%, ${COLORS.success.light} 100%)`,
                boxShadow: `0 0 12px ${COLORS.success.glow}`,
              },
            }}
          />
        </Box>
      </Fade>
    </Modal>
  );
};
