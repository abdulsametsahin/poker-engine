import React, { useEffect, useState } from 'react';
import { Box, Paper, Typography, Stack, Button, Fade, Modal } from '@mui/material';
import { EmojiEvents, Home } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';

interface GameCompleteDisplayProps {
  winner: string;
  winnerName?: string;
  finalChips: number;
  totalPlayers: number;
  message: string;
  currentUserId?: string;
}

const GameCompleteDisplay: React.FC<GameCompleteDisplayProps> = ({
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
      }}
    >
      <Fade in={open}>
        <Paper
          elevation={24}
          sx={{
            p: 5,
            maxWidth: 600,
            bgcolor: isWinner ? 'success.50' : 'background.paper',
            borderRadius: 3,
            border: '6px solid',
            borderColor: isWinner ? 'success.main' : 'warning.main',
            position: 'relative',
            overflow: 'hidden',
            textAlign: 'center',
            '&::before': {
              content: '""',
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              background: isWinner
                ? 'linear-gradient(45deg, rgba(76,175,80,0.1) 0%, rgba(139,195,74,0.1) 100%)'
                : 'linear-gradient(45deg, rgba(255,152,0,0.1) 0%, rgba(255,193,7,0.1) 100%)',
              animation: 'shimmer 3s infinite',
            },
            '@keyframes shimmer': {
              '0%': { opacity: 0.3 },
              '50%': { opacity: 0.6 },
              '100%': { opacity: 0.3 },
            },
          }}
        >
          <Stack spacing={4} alignItems="center" sx={{ position: 'relative', zIndex: 1 }}>
            {/* Trophy Icon */}
            <Box
              sx={{
                animation: isWinner ? 'bounce 1.5s ease-in-out infinite' : 'none',
                '@keyframes bounce': {
                  '0%, 100%': { transform: 'translateY(0) scale(1)' },
                  '50%': { transform: 'translateY(-30px) scale(1.1)' },
                },
              }}
            >
              <EmojiEvents
                sx={{
                  fontSize: 120,
                  color: isWinner ? 'success.main' : 'warning.main',
                }}
              />
            </Box>

            {/* Game Over Title */}
            <Typography
              variant="h2"
              fontWeight="bold"
              color={isWinner ? 'success.main' : 'text.primary'}
              sx={{
                textShadow: '2px 2px 4px rgba(0,0,0,0.2)',
              }}
            >
              {message}
            </Typography>

            {/* Winner Information */}
            <Paper
              elevation={4}
              sx={{
                p: 4,
                width: '100%',
                bgcolor: isWinner ? 'success.100' : 'grey.100',
                border: '3px solid',
                borderColor: isWinner ? 'success.main' : 'grey.300',
              }}
            >
              <Stack spacing={2}>
                <Typography variant="h4" fontWeight="bold" color="primary">
                  {isWinner ? '🎉 You Won! 🎉' : `${winnerName || winner.slice(0, 12)} Won!`}
                </Typography>

                <Stack direction="row" spacing={3} justifyContent="center" alignItems="center">
                  <Box>
                    <Typography variant="caption" color="text.secondary">
                      Final Chips
                    </Typography>
                    <Typography variant="h3" fontWeight="bold" color="success.main">
                      ${finalChips.toLocaleString()}
                    </Typography>
                  </Box>

                  <Box sx={{ width: 2, height: 60, bgcolor: 'divider' }} />

                  <Box>
                    <Typography variant="caption" color="text.secondary">
                      Players Defeated
                    </Typography>
                    <Typography variant="h3" fontWeight="bold" color="error.main">
                      {totalPlayers - 1}
                    </Typography>
                  </Box>
                </Stack>

                {isWinner && (
                  <Typography variant="body1" color="text.secondary" sx={{ mt: 2 }}>
                    Congratulations! You've eliminated all opponents and won the game!
                  </Typography>
                )}

                {!isWinner && (
                  <Typography variant="body1" color="text.secondary" sx={{ mt: 2 }}>
                    Better luck next time! Keep practicing and you'll win soon!
                  </Typography>
                )}
              </Stack>
            </Paper>

            {/* Return to Lobby Button */}
            <Button
              variant="contained"
              size="large"
              startIcon={<Home />}
              onClick={handleReturnToLobby}
              sx={{
                px: 4,
                py: 1.5,
                fontSize: 18,
                fontWeight: 'bold',
                background: 'linear-gradient(45deg, #2196F3 30%, #21CBF3 90%)',
                boxShadow: '0 3px 5px 2px rgba(33, 203, 243, .3)',
              }}
            >
              Return to Lobby
            </Button>
          </Stack>
        </Paper>
      </Fade>
    </Modal>
  );
};

export default GameCompleteDisplay;
