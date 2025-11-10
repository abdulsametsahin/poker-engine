import React, { memo } from 'react';
import { Box, Stack, Typography, Fade } from '@mui/material';
import { PlayingCard } from './PlayingCard';
import { COLORS, RADIUS, TRANSITIONS } from '../../constants';
import { Player, WinnerInfo } from '../../types';
import { formatUsername } from '../../utils';

interface ShowdownDisplayProps {
  players: Player[];
  winners: WinnerInfo[];
  show: boolean;
}

export const ShowdownDisplay: React.FC<ShowdownDisplayProps> = memo(({ players, winners, show }) => {
  if (!show || !players || players.length === 0) return null;

  // Filter out folded players and those without cards
  const playersInShowdown = players.filter(p => !p.folded && p.cards && p.cards.length > 0);
  
  if (playersInShowdown.length === 0) return null;

  // Check if a player is a winner
  const isWinner = (playerId: string) => {
    return winners.some(w => w.playerId === playerId);
  };

  return (
    <Fade in={show} timeout={800}>
      <Box
        sx={{
          position: 'absolute',
          bottom: 20,
          left: '50%',
          transform: 'translateX(-50%)',
          zIndex: 100,
          maxWidth: '90%',
          width: 'auto',
        }}
      >
        <Stack
          direction="row"
          spacing={2}
          sx={{
            px: 3,
            py: 2,
            borderRadius: RADIUS.lg,
            background: `linear-gradient(135deg, ${COLORS.background.paper}f5 0%, ${COLORS.background.tertiary}f5 100%)`,
            backdropFilter: 'blur(20px)',
            border: `2px solid ${COLORS.border.heavy}`,
            boxShadow: `0 8px 32px rgba(0, 0, 0, 0.6), 0 0 20px ${COLORS.primary.glow}`,
            overflowX: 'auto',
            '&::-webkit-scrollbar': {
              height: 6,
            },
            '&::-webkit-scrollbar-track': {
              background: COLORS.background.secondary,
              borderRadius: RADIUS.sm,
            },
            '&::-webkit-scrollbar-thumb': {
              background: COLORS.primary.main,
              borderRadius: RADIUS.sm,
            },
          }}
        >
          {playersInShowdown.map((player, idx) => {
            const playerIsWinner = isWinner(player.user_id);
            
            return (
              <Fade key={player.user_id} in={show} timeout={1000 + idx * 150}>
                <Box
                  sx={{
                    minWidth: 140,
                    borderRadius: RADIUS.md,
                    p: 1.5,
                    background: playerIsWinner
                      ? `linear-gradient(135deg, ${COLORS.success.main}20 0%, ${COLORS.success.dark}15 100%)`
                      : `linear-gradient(135deg, ${COLORS.danger.main}10 0%, ${COLORS.background.secondary}80 100%)`,
                    border: `2px solid ${playerIsWinner ? COLORS.success.main : COLORS.danger.main}`,
                    transition: TRANSITIONS.normal,
                    boxShadow: playerIsWinner
                      ? `0 4px 16px ${COLORS.success.glow}`
                      : `0 2px 8px ${COLORS.danger.main}40`,
                    '@keyframes winner-glow': {
                      '0%, 100%': {
                        boxShadow: `0 4px 16px ${COLORS.success.glow}`,
                      },
                      '50%': {
                        boxShadow: `0 6px 24px ${COLORS.success.glow}, 0 0 12px ${COLORS.success.glow}`,
                      },
                    },
                    animation: playerIsWinner ? 'winner-glow 2s ease-in-out infinite' : 'none',
                  }}
                >
                  <Stack spacing={1} alignItems="center">
                    {/* Player name */}
                    <Typography
                      variant="caption"
                      sx={{
                        color: playerIsWinner ? COLORS.success.main : COLORS.text.primary,
                        fontWeight: 700,
                        fontSize: 11,
                        textAlign: 'center',
                        maxWidth: '100%',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                      }}
                    >
                      {playerIsWinner && '👑 '}
                      {formatUsername(player.username || player.user_id.slice(0, 8))}
                    </Typography>

                    {/* Player cards */}
                    <Stack direction="row" spacing={0.5}>
                      {player.cards?.map((card, cardIdx) => {
                        const cardStr = typeof card === 'string' ? card : `${card.rank}${card.suit}`;
                        return (
                          <Box
                            key={cardIdx}
                            sx={{
                              '@keyframes card-flip': {
                                '0%': {
                                  transform: 'rotateY(90deg)',
                                  opacity: 0,
                                },
                                '100%': {
                                  transform: 'rotateY(0deg)',
                                  opacity: 1,
                                },
                              },
                              animation: `card-flip 0.4s ease-out ${0.3 + idx * 0.1 + cardIdx * 0.1}s backwards`,
                            }}
                          >
                            <PlayingCard
                              card={cardStr}
                              size="small"
                              highlight={playerIsWinner}
                            />
                          </Box>
                        );
                      })}
                    </Stack>

                    {/* Winner/Loser indicator */}
                    <Typography
                      variant="caption"
                      sx={{
                        color: playerIsWinner ? COLORS.success.main : COLORS.danger.main,
                        fontWeight: 900,
                        fontSize: 10,
                        textTransform: 'uppercase',
                        letterSpacing: '0.05em',
                      }}
                    >
                      {playerIsWinner ? 'Winner' : 'Loser'}
                    </Typography>
                  </Stack>
                </Box>
              </Fade>
            );
          })}
        </Stack>
      </Box>
    </Fade>
  );
});

ShowdownDisplay.displayName = 'ShowdownDisplay';

export default ShowdownDisplay;
