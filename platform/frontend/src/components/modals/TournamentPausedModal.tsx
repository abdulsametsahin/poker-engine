import React from 'react';
import { Box, Stack, Typography, Fade, Modal } from '@mui/material';
import { Pause } from '@mui/icons-material';
import { COLORS, RADIUS } from '../../constants';

interface TournamentPausedModalProps {
  open: boolean;
}

const TournamentPausedModal: React.FC<TournamentPausedModalProps> = ({ open }) => {
  return (
    <Modal
      open={open}
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
          sx={{
            maxWidth: 500,
            width: '90%',
            borderRadius: RADIUS.lg,
            background: `linear-gradient(135deg, ${COLORS.background.paper} 0%, ${COLORS.background.tertiary} 100%)`,
            border: `2px solid ${COLORS.warning.main}`,
            boxShadow: `0 8px 32px rgba(0, 0, 0, 0.6)`,
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          {/* Background effect */}
          <Box
            sx={{
              position: 'absolute',
              inset: 0,
              background: `radial-gradient(circle at 50% 0%, ${COLORS.warning.main}15 0%, transparent 50%)`,
              pointerEvents: 'none',
            }}
          />

          <Stack spacing={2} sx={{ p: 4, position: 'relative', zIndex: 1, textAlign: 'center' }}>
            {/* Icon */}
            <Box sx={{ display: 'flex', justifyContent: 'center' }}>
              <Pause
                sx={{
                  fontSize: 64,
                  color: COLORS.warning.main,
                }}
              />
            </Box>

            {/* Title */}
            <Typography
              variant="h5"
              sx={{
                fontWeight: 800,
                color: COLORS.warning.main,
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
              }}
            >
              Tournament Paused
            </Typography>

            {/* Message */}
            <Typography
              variant="body1"
              sx={{
                color: COLORS.text.secondary,
                fontSize: 16,
                lineHeight: 1.6,
              }}
            >
              The tournament has been paused by the organizer.
              <br />
              Please wait while the tournament resumes.
            </Typography>
          </Stack>
        </Box>
      </Fade>
    </Modal>
  );
};

export default TournamentPausedModal;
