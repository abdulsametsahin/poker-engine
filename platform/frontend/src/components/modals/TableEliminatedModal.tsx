import React, { useEffect, useState } from 'react';
import { Box, Stack, Typography, Fade, Modal, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material';
import { EmojiEvents, ExitToApp, SentimentVeryDissatisfied } from '@mui/icons-material';
import { Button } from '../common/Button';
import { COLORS, RADIUS } from '../../constants';

interface PlayerStanding {
  playerId: string;
  username?: string;
  chips: number;
  position: number;
  isCurrentUser?: boolean;
}

interface TableEliminatedModalProps {
  show: boolean;
  isEliminated: boolean;
  tableStandings: PlayerStanding[];
  onReturnToTournament: () => void;
  countdown?: number; // Countdown in seconds before enabling button
}

const TableEliminatedModal: React.FC<TableEliminatedModalProps> = ({
  show,
  isEliminated,
  tableStandings,
  onReturnToTournament,
  countdown: initialCountdown = 5,
}) => {
  const [countdown, setCountdown] = useState(initialCountdown);
  const [buttonEnabled, setButtonEnabled] = useState(false);

  useEffect(() => {
    if (show) {
      setCountdown(initialCountdown);
      setButtonEnabled(false);

      // Countdown timer
      const countdownInterval = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) {
            clearInterval(countdownInterval);
            setButtonEnabled(true);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);

      return () => {
        clearInterval(countdownInterval);
      };
    }
  }, [show, initialCountdown]);

  if (!show) return null;

  // Sort standings by position
  const sortedStandings = [...tableStandings].sort((a, b) => a.position - b.position);

  return (
    <Modal
      open={show}
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backdropFilter: 'blur(6px)',
        backgroundColor: 'rgba(0, 0, 0, 0.85)',
      }}
    >
      <Fade in={show} timeout={500}>
        <Box
          sx={{
            maxWidth: 700,
            width: '90%',
            borderRadius: RADIUS.lg,
            background: `linear-gradient(135deg, ${COLORS.background.paper} 0%, ${COLORS.background.tertiary} 100%)`,
            border: `2px solid ${isEliminated ? COLORS.danger.main : COLORS.accent.main}`,
            boxShadow: `0 12px 40px rgba(0, 0, 0, 0.7), 0 0 30px ${isEliminated ? COLORS.danger.glow : COLORS.accent.glow}`,
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          {/* Subtle background */}
          <Box
            sx={{
              position: 'absolute',
              inset: 0,
              background: isEliminated
                ? `radial-gradient(circle at 50% 0%, ${COLORS.danger.main}15 0%, transparent 60%)`
                : `radial-gradient(circle at 50% 0%, ${COLORS.accent.main}15 0%, transparent 60%)`,
              pointerEvents: 'none',
            }}
          />

          <Stack spacing={3} sx={{ p: 4, position: 'relative', zIndex: 1 }}>
            {/* Header */}
            <Stack spacing={2} alignItems="center">
              {isEliminated ? (
                <SentimentVeryDissatisfied
                  sx={{
                    fontSize: 64,
                    color: COLORS.danger.main,
                    animation: 'shake 0.5s ease-in-out',
                    '@keyframes shake': {
                      '0%, 100%': { transform: 'translateX(0)' },
                      '10%, 30%, 50%, 70%, 90%': { transform: 'translateX(-5px)' },
                      '20%, 40%, 60%, 80%': { transform: 'translateX(5px)' },
                    },
                  }}
                />
              ) : (
                <EmojiEvents
                  sx={{
                    fontSize: 64,
                    color: COLORS.accent.main,
                    animation: 'bounce 1s ease-in-out infinite',
                    '@keyframes bounce': {
                      '0%, 100%': { transform: 'translateY(0)' },
                      '50%': { transform: 'translateY(-10px)' },
                    },
                  }}
                />
              )}

              <Typography
                variant="h4"
                sx={{
                  fontWeight: 800,
                  color: isEliminated ? COLORS.danger.main : COLORS.accent.main,
                  textTransform: 'uppercase',
                  letterSpacing: '0.1em',
                  textAlign: 'center',
                }}
              >
                {isEliminated ? 'You\'ve Been Eliminated' : 'Table Complete!'}
              </Typography>

              <Typography
                variant="body1"
                sx={{
                  color: COLORS.text.secondary,
                  textAlign: 'center',
                  fontSize: 14,
                }}
              >
                {isEliminated
                  ? 'Better luck next time! Thanks for playing.'
                  : 'Great job! The tournament continues.'}
              </Typography>
            </Stack>

            {/* Final Standings Table */}
            <Box>
              <Typography
                variant="h6"
                sx={{
                  fontWeight: 700,
                  color: COLORS.text.primary,
                  mb: 1.5,
                  textAlign: 'center',
                }}
              >
                Final Table Standings
              </Typography>

              <TableContainer
                sx={{
                  borderRadius: RADIUS.md,
                  background: `${COLORS.background.secondary}80`,
                  border: `1px solid ${COLORS.border.main}`,
                  maxHeight: 300,
                  overflowY: 'auto',
                }}
              >
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell
                        sx={{
                          color: COLORS.text.secondary,
                          fontWeight: 700,
                          fontSize: 12,
                          borderBottom: `1px solid ${COLORS.border.main}`,
                        }}
                      >
                        Position
                      </TableCell>
                      <TableCell
                        sx={{
                          color: COLORS.text.secondary,
                          fontWeight: 700,
                          fontSize: 12,
                          borderBottom: `1px solid ${COLORS.border.main}`,
                        }}
                      >
                        Player
                      </TableCell>
                      <TableCell
                        align="right"
                        sx={{
                          color: COLORS.text.secondary,
                          fontWeight: 700,
                          fontSize: 12,
                          borderBottom: `1px solid ${COLORS.border.main}`,
                        }}
                      >
                        Chips
                      </TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {sortedStandings.map((player, idx) => (
                      <TableRow
                        key={player.playerId}
                        sx={{
                          background: player.isCurrentUser
                            ? `${COLORS.primary.main}20`
                            : 'transparent',
                          borderLeft: player.isCurrentUser
                            ? `3px solid ${COLORS.primary.main}`
                            : 'none',
                        }}
                      >
                        <TableCell
                          sx={{
                            color: COLORS.text.primary,
                            fontWeight: player.isCurrentUser ? 700 : 500,
                            fontSize: 14,
                            borderBottom: idx === sortedStandings.length - 1 ? 'none' : `1px solid ${COLORS.border.main}`,
                          }}
                        >
                          {player.position === 1 ? '🥇' : player.position === 2 ? '🥈' : player.position === 3 ? '🥉' : `#${player.position}`}
                        </TableCell>
                        <TableCell
                          sx={{
                            color: player.isCurrentUser ? COLORS.primary.main : COLORS.text.primary,
                            fontWeight: player.isCurrentUser ? 700 : 500,
                            fontSize: 14,
                            borderBottom: idx === sortedStandings.length - 1 ? 'none' : `1px solid ${COLORS.border.main}`,
                          }}
                        >
                          {player.isCurrentUser ? '🎮 You' : (player.username || player.playerId.slice(0, 12))}
                        </TableCell>
                        <TableCell
                          align="right"
                          sx={{
                            color: COLORS.accent.main,
                            fontWeight: 600,
                            fontFamily: 'monospace',
                            fontSize: 14,
                            borderBottom: idx === sortedStandings.length - 1 ? 'none' : `1px solid ${COLORS.border.main}`,
                          }}
                        >
                          ${player.chips.toLocaleString()}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            </Box>

            {/* Return Button */}
            <Box>
              <Button
                variant="primary"
                onClick={onReturnToTournament}
                disabled={!buttonEnabled}
                sx={{
                  width: '100%',
                  py: 1.5,
                  fontSize: 16,
                  fontWeight: 700,
                  opacity: buttonEnabled ? 1 : 0.5,
                  cursor: buttonEnabled ? 'pointer' : 'not-allowed',
                }}
                startIcon={<ExitToApp />}
              >
                {buttonEnabled
                  ? 'RETURN TO TOURNAMENT'
                  : `RETURN TO TOURNAMENT (${countdown}s)`}
              </Button>

              {!buttonEnabled && (
                <Typography
                  variant="caption"
                  sx={{
                    textAlign: 'center',
                    color: COLORS.text.secondary,
                    fontSize: 11,
                    mt: 1,
                    display: 'block',
                  }}
                >
                  Please review the final standings
                </Typography>
              )}
            </Box>
          </Stack>
        </Box>
      </Fade>
    </Modal>
  );
};

export default TableEliminatedModal;
