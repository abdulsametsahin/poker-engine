import React from 'react';
import { Box, Stack, Typography, Fade, Modal, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material';
import { EmojiEvents, Home, Refresh } from '@mui/icons-material';
import { Button } from '../common/Button';
import { COLORS, RADIUS } from '../../constants';

interface PlayerStanding {
  playerId: string;
  username?: string;
  chips: number;
  profit: number;
  isCurrentUser?: boolean;
}

interface GameCompleteModalProps {
  show: boolean;
  standings: PlayerStanding[];
  onPlayAgain: () => void;
  onReturnToLobby: () => void;
}

const GameCompleteModal: React.FC<GameCompleteModalProps> = ({
  show,
  standings,
  onPlayAgain,
  onReturnToLobby,
}) => {
  if (!show) return null;

  // Sort standings by chips (descending)
  const sortedStandings = [...standings].sort((a, b) => b.chips - a.chips);

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
            border: `2px solid ${COLORS.accent.main}`,
            boxShadow: `0 12px 40px rgba(0, 0, 0, 0.7), 0 0 30px ${COLORS.accent.glow}`,
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          {/* Subtle background */}
          <Box
            sx={{
              position: 'absolute',
              inset: 0,
              background: `radial-gradient(circle at 50% 0%, ${COLORS.accent.main}15 0%, transparent 60%)`,
              pointerEvents: 'none',
            }}
          />

          <Stack spacing={3} sx={{ p: 4, position: 'relative', zIndex: 1 }}>
            {/* Header */}
            <Stack spacing={2} alignItems="center">
              <EmojiEvents
                sx={{
                  fontSize: 64,
                  color: COLORS.accent.main,
                  animation: 'rotate 2s linear infinite',
                  '@keyframes rotate': {
                    '0%': { transform: 'rotate(0deg)' },
                    '100%': { transform: 'rotate(360deg)' },
                  },
                }}
              />

              <Typography
                variant="h4"
                sx={{
                  fontWeight: 800,
                  color: COLORS.accent.main,
                  textTransform: 'uppercase',
                  letterSpacing: '0.1em',
                  textAlign: 'center',
                }}
              >
                🎮 Game Complete!
              </Typography>

              <Typography
                variant="body1"
                sx={{
                  color: COLORS.text.secondary,
                  textAlign: 'center',
                  fontSize: 14,
                }}
              >
                Thanks for playing! Here are the final results.
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
                Final Standings
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
                        Rank
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
                      <TableCell
                        align="right"
                        sx={{
                          color: COLORS.text.secondary,
                          fontWeight: 700,
                          fontSize: 12,
                          borderBottom: `1px solid ${COLORS.border.main}`,
                        }}
                      >
                        Profit/Loss
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
                          {idx === 0 ? '🥇' : idx === 1 ? '🥈' : idx === 2 ? '🥉' : `#${idx + 1}`}
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
                        <TableCell
                          align="right"
                          sx={{
                            color: player.profit > 0 ? COLORS.success.main : player.profit < 0 ? COLORS.danger.main : COLORS.text.secondary,
                            fontWeight: 700,
                            fontFamily: 'monospace',
                            fontSize: 14,
                            borderBottom: idx === sortedStandings.length - 1 ? 'none' : `1px solid ${COLORS.border.main}`,
                          }}
                        >
                          {player.profit > 0 ? '+' : ''}${player.profit.toLocaleString()}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            </Box>

            {/* Action Buttons */}
            <Stack direction="row" spacing={2}>
              <Button
                variant="primary"
                onClick={onPlayAgain}
                sx={{
                  flex: 1,
                  py: 1.5,
                  fontSize: 16,
                  fontWeight: 700,
                }}
                startIcon={<Refresh />}
              >
                Play Again
              </Button>
              <Button
                variant="secondary"
                onClick={onReturnToLobby}
                sx={{
                  flex: 1,
                  py: 1.5,
                  fontSize: 16,
                  fontWeight: 700,
                }}
                startIcon={<Home />}
              >
                Return to Lobby
              </Button>
            </Stack>
          </Stack>
        </Box>
      </Fade>
    </Modal>
  );
};

export default GameCompleteModal;
