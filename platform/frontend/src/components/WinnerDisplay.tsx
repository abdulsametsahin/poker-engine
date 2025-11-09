import React, { useEffect, useState } from 'react';
import { Box, Paper, Typography, Stack, Chip, Fade, Modal } from '@mui/material';
import { EmojiEvents } from '@mui/icons-material';
import PlayingCard from './PlayingCard';

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

const WinnerDisplay: React.FC<WinnerDisplayProps> = ({ winners, onClose }) => {
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
    setTimeout(onClose, 300); // Wait for animation
  };

  // Convert card object to string format (e.g., {rank: "A", suit: "s"} -> "As")
  const cardToString = (card: string | CardObject): string => {
    if (typeof card === 'string') {
      return card;
    }
    return `${card.rank}${card.suit}`;
  };

  if (!winners || winners.length === 0) return null;

  return (
    <Modal
      open={open}
      onClose={handleClose}
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <Fade in={open}>
        <Paper
          elevation={24}
          sx={{
            p: 4,
            maxWidth: 600,
            bgcolor: 'background.paper',
            borderRadius: 3,
            border: '4px solid',
            borderColor: 'warning.main',
            position: 'relative',
            overflow: 'hidden',
            '&::before': {
              content: '""',
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              background: 'linear-gradient(45deg, rgba(255,215,0,0.1) 0%, rgba(255,140,0,0.1) 100%)',
              animation: 'shimmer 2s infinite',
            },
            '@keyframes shimmer': {
              '0%': { opacity: 0.5 },
              '50%': { opacity: 1 },
              '100%': { opacity: 0.5 },
            },
          }}
        >
          <Stack spacing={3} alignItems="center" sx={{ position: 'relative', zIndex: 1 }}>
            {/* Trophy Icon */}
            <Box
              sx={{
                animation: 'bounce 1s ease-in-out',
                '@keyframes bounce': {
                  '0%, 100%': { transform: 'translateY(0)' },
                  '50%': { transform: 'translateY(-20px)' },
                },
              }}
            >
              <EmojiEvents sx={{ fontSize: 80, color: 'warning.main' }} />
            </Box>

            {/* Winner(s) Title */}
            <Typography variant="h3" fontWeight="bold" color="warning.main" textAlign="center">
              {winners.length === 1 ? 'Winner!' : 'Winners!'}
            </Typography>

            {/* Winner Cards */}
            {winners.map((winner, idx) => (
              <Paper
                key={idx}
                elevation={4}
                sx={{
                  p: 3,
                  width: '100%',
                  bgcolor: idx === 0 ? 'warning.50' : 'grey.50',
                  border: '2px solid',
                  borderColor: idx === 0 ? 'warning.main' : 'grey.300',
                }}
              >
                <Stack spacing={2}>
                  {/* Player Info */}
                  <Stack direction="row" justifyContent="space-between" alignItems="center">
                    <Typography variant="h6" fontWeight="bold">
                      {winner.playerId.slice(0, 12)}
                    </Typography>
                    <Chip
                      label={`+$${winner.amount}`}
                      color="success"
                      size="medium"
                      sx={{ fontWeight: 'bold', fontSize: 16 }}
                    />
                  </Stack>

                  {/* Hand Rank */}
                  <Typography
                    variant="h5"
                    color="primary"
                    fontWeight="bold"
                    textAlign="center"
                  >
                    {winner.handRank}
                  </Typography>

                  {/* Winning Cards */}
                  {winner.handCards && winner.handCards.length > 0 && (
                    <Stack direction="row" spacing={1} justifyContent="center">
                      {winner.handCards.map((card, cardIdx) => (
                        <PlayingCard key={cardIdx} card={cardToString(card)} size="medium" />
                      ))}
                    </Stack>
                  )}
                </Stack>
              </Paper>
            ))}

            {/* Close hint */}
            <Typography variant="caption" color="text.secondary">
              Click anywhere to continue
            </Typography>
          </Stack>
        </Paper>
      </Fade>
    </Modal>
  );
};

export default WinnerDisplay;
