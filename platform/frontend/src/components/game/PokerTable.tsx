import React, { memo, useMemo } from 'react';
import { Box, Stack, Typography } from '@mui/material';
import { PlayerSeat } from './PlayerSeat';
import { PlayingCard } from './PlayingCard';
import { ShowdownDisplay } from './ShowdownDisplay';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { OvalTableSVG } from './OvalTableSVG';
import { DealerButton } from './DealerButton';
import { COLORS, RADIUS, SPACING, TRANSITIONS } from '../../constants';
import { Player, WinnerInfo } from '../../types';
import { getBettingRoundName } from '../../utils';

interface TableState {
  table_id?: string;
  players: Player[];
  community_cards?: string[];
  pot?: number;
  current_turn?: string;
  status?: string;
  betting_round?: string;
  current_bet?: number;
  action_deadline?: string;
  winners?: WinnerInfo[];
}

interface PokerTableProps {
  tableState: TableState | null;
  currentUserId?: string;
}

export const PokerTable: React.FC<PokerTableProps> = memo(({
  tableState,
  currentUserId,
}) => {
  const {
    players = [],
    community_cards = [],
    pot = 0,
    current_turn,
    status,
    betting_round,
    current_bet = 0,
    action_deadline,
    winners = [],
  } = tableState || {};

  // Calculate positions for oval perimeter layout
  // Current user is always positioned at the bottom, others arranged clockwise
  const { getPlayerPosition, getDealerButtonPosition } = useMemo(() => {
    const currentUserIndex = players.findIndex(p => p?.user_id === currentUserId);

    const getPlayerPosition = (index: number, total: number) => {
      // Calculate offset so current user is at bottom (90 degrees)
      let adjustedIndex = index;
      if (currentUserIndex !== -1) {
        adjustedIndex = (index - currentUserIndex + total) % total;
      }

      // Start from bottom (π/2) and go counter-clockwise
      const angle = (adjustedIndex / total) * 2 * Math.PI + Math.PI / 2;

      // Adjusted radius to match oval table perimeter
      // SVG viewBox is 1200x800, felt area is ~480x310, so positions at ~50-52% radius
      const radiusX = 42; // Horizontal spread (slightly tighter)
      const radiusY = 38; // Vertical spread (slightly larger for oval)

      return {
        left: `${50 + radiusX * Math.cos(angle)}%`,
        top: `${50 + radiusY * Math.sin(angle)}%`,
        transform: 'translate(-50%, -50%)',
      };
    };

    // Position dealer button near the dealer (offset slightly inward from seat)
    const getDealerButtonPosition = (dealerIndex: number, total: number) => {
      let adjustedIndex = dealerIndex;
      if (currentUserIndex !== -1) {
        adjustedIndex = (dealerIndex - currentUserIndex + total) % total;
      }

      const angle = (adjustedIndex / total) * 2 * Math.PI + Math.PI / 2;
      const radiusX = 35; // Closer to center than player seat
      const radiusY = 31;

      return {
        left: `${50 + radiusX * Math.cos(angle)}%`,
        top: `${50 + radiusY * Math.sin(angle)}%`,
      };
    };

    return { getPlayerPosition, getDealerButtonPosition };
  }, [players, currentUserId]);

  // Find dealer index
  const dealerIndex = players.findIndex(p => p?.is_dealer);

  if (!tableState) {
    return (
      <Box
        sx={{
          height: '100%',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <LoadingSpinner fullScreen={false} />
      </Box>
    );
  }

  return (
    <Box
      sx={{
        height: '100%',
        width: '100%',
        position: 'relative',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'transparent',
        overflow: 'visible',
      }}
    >
      {/* SVG Oval Poker Table - Maintains 3:2 aspect ratio */}
      <Box
        sx={{
          position: 'absolute',
          width: '100%',
          height: '100%',
          maxWidth: '1200px',
          maxHeight: '800px',
          top: '50%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          // Maintain aspect ratio
          aspectRatio: '3 / 2',
          '@media (max-width: 1200px)': {
            width: '95%',
            height: 'auto',
          },
          '@media (max-width: 768px)': {
            width: '98%',
            height: 'auto',
          },
        }}
      >
        <OvalTableSVG />
      </Box>

      {/* Center area */}
      <Box
        sx={{
          position: 'absolute',
          top: '50%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: 2,
          zIndex: 10,
        }}
      >
        {/* Pot */}
        <Box
          sx={{
            px: { xs: 1.5, sm: 2, md: 2.5 },
            py: { xs: 0.75, sm: 1, md: 1.25 },
            borderRadius: RADIUS.sm,
            background: 'rgba(0, 0, 0, 0.75)',
            backdropFilter: 'blur(16px)',
            border: `2px solid ${COLORS.accent.main}`,
            boxShadow: `
              0 4px 16px rgba(0, 0, 0, 0.6), 
              0 0 24px ${COLORS.accent.glow},
              inset 0 1px 0 rgba(255, 255, 255, 0.1)
            `,
            minWidth: { xs: 80, sm: 90, md: 100 },
            textAlign: 'center',
            position: 'relative',
            overflow: 'hidden',
            // Animated shimmer effect
            '&::before': {
              content: '""',
              position: 'absolute',
              top: 0,
              left: '-100%',
              width: '100%',
              height: '100%',
              background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.1), transparent)',
              animation: 'shimmer 3s infinite',
            },
            '@keyframes shimmer': {
              '0%': { left: '-100%' },
              '100%': { left: '100%' },
            },
          }}
        >
          <Typography
            variant="caption"
            sx={{
              color: COLORS.text.secondary,
              fontSize: { xs: '8px', sm: '9px', md: '10px' },
              fontWeight: 700,
              letterSpacing: '0.12em',
              textShadow: '0 2px 4px rgba(0,0,0,0.5)',
            }}
          >
            POT
          </Typography>
          <Typography
            variant="h4"
            sx={{
              color: COLORS.accent.main,
              fontWeight: 900,
              fontSize: { xs: '18px', sm: '20px', md: '24px' },
              lineHeight: 1.1,
              fontFamily: 'monospace',
              textShadow: `0 0 16px ${COLORS.accent.glow}, 0 2px 8px rgba(0,0,0,0.8)`,
              letterSpacing: '0.05em',
            }}
          >
            ${pot}
          </Typography>
        </Box>

        {/* Community cards */}
        {community_cards.length > 0 && (
          <Stack
            direction="row"
            spacing={{ xs: 0.5, sm: 0.65, md: 0.75 }}
            sx={{
              px: { xs: 1, sm: 1.25, md: 1.5 },
              py: { xs: 0.75, sm: 0.85, md: 1 },
              borderRadius: RADIUS.sm,
              background: 'rgba(0, 0, 0, 0.65)',
              backdropFilter: 'blur(16px)',
              border: `1.5px solid ${COLORS.border.heavy}`,
              boxShadow: '0 4px 16px rgba(0, 0, 0, 0.6), inset 0 1px 0 rgba(255, 255, 255, 0.05)',
            }}
          >
            {community_cards.map((card, idx) => (
              <PlayingCard
                key={idx}
                card={card}
                size="medium"
                dealAnimation={idx >= 0} // All cards have deal animation
              />
            ))}
          </Stack>
        )}

        {/* Betting round indicator */}
        {status === 'playing' && betting_round && (
          <Box
            sx={{
              px: { xs: 1.5, sm: 2, md: 2.5 },
              py: { xs: 0.5, sm: 0.65, md: 0.75 },
              borderRadius: RADIUS.sm,
              background: `linear-gradient(135deg, ${COLORS.primary.main}50 0%, ${COLORS.secondary.main}50 100%)`,
              backdropFilter: 'blur(16px)',
              border: `1.5px solid ${COLORS.primary.main}`,
              boxShadow: `
                0 4px 12px rgba(0, 0, 0, 0.5), 
                0 0 16px ${COLORS.primary.glow},
                inset 0 1px 0 rgba(255, 255, 255, 0.15)
              `,
              position: 'relative',
              overflow: 'hidden',
              // Animated pulse effect
              animation: 'pulse-glow 2s ease-in-out infinite',
              '@keyframes pulse-glow': {
                '0%, 100%': {
                  boxShadow: `0 4px 12px rgba(0, 0, 0, 0.5), 0 0 16px ${COLORS.primary.glow}, inset 0 1px 0 rgba(255, 255, 255, 0.15)`,
                },
                '50%': {
                  boxShadow: `0 4px 16px rgba(0, 0, 0, 0.6), 0 0 24px ${COLORS.primary.glow}, inset 0 1px 0 rgba(255, 255, 255, 0.2)`,
                },
              },
            }}
          >
            <Typography
              variant="caption"
              sx={{
                color: COLORS.primary.light,
                fontSize: { xs: '9px', sm: '10px', md: '11px' },
                fontWeight: 800,
                letterSpacing: '0.1em',
                textShadow: '0 2px 4px rgba(0,0,0,0.5)',
              }}
            >
              {getBettingRoundName(betting_round)}
            </Typography>
            {current_bet > 0 && (
              <Typography
                variant="caption"
                sx={{
                  color: COLORS.text.secondary,
                  fontSize: { xs: '7px', sm: '8px', md: '9px' },
                  display: 'block',
                  mt: 0.2,
                  textShadow: '0 1px 2px rgba(0,0,0,0.5)',
                }}
              >
                Current Bet: ${current_bet}
              </Typography>
            )}
          </Box>
        )}
      </Box>

      {/* Players in circular arrangement */}
      {players.map((player, index) => {
        const position = getPlayerPosition(index, players.length);
        return (
          <Box
            key={player?.user_id || index}
            sx={{
              position: 'absolute',
              ...position,
              zIndex: current_turn === player?.user_id ? 20 : 15,
            }}
          >
            <PlayerSeat
              player={player}
              position={player?.seat_number || index}
              isActive={current_turn === player?.user_id}
              isCurrentUser={currentUserId === player?.user_id}
              actionDeadline={current_turn === player?.user_id ? action_deadline : undefined}
            />
          </Box>
        );
      })}

      {/* Dealer Button */}
      {dealerIndex !== -1 && players.length > 1 && (
        <DealerButton
          position={getDealerButtonPosition(dealerIndex, players.length)}
        />
      )}

      {/* Status messages */}
      {status === 'waiting' && (
        <Box
          sx={{
            position: 'absolute',
            top: 16,
            left: '50%',
            transform: 'translateX(-50%)',
            px: 3,
            py: 1,
            borderRadius: RADIUS.md,
            background: 'rgba(0, 0, 0, 0.7)',
            backdropFilter: 'blur(10px)',
            border: `1px solid ${COLORS.info.main}`,
            boxShadow: `0 0 12px ${COLORS.info.glow}`,
            zIndex: 5,
          }}
        >
          <Typography
            variant="body2"
            sx={{
              color: COLORS.info.main,
              fontSize: '13px',
              fontWeight: 600,
            }}
          >
            ⏳ Waiting for game to start... ({players.length} player{players.length !== 1 ? 's' : ''})
          </Typography>
        </Box>
      )}

      {status === 'handComplete' && (
        <Box
          sx={{
            position: 'absolute',
            top: 16,
            left: '50%',
            transform: 'translateX(-50%)',
            px: 3,
            py: 1,
            borderRadius: RADIUS.md,
            background: 'rgba(0, 0, 0, 0.7)',
            backdropFilter: 'blur(10px)',
            border: `1px solid ${COLORS.success.main}`,
            boxShadow: `0 0 12px ${COLORS.success.glow}`,
            zIndex: 5,
          }}
        >
          <Typography
            variant="body2"
            sx={{
              color: COLORS.success.main,
              fontSize: '13px',
              fontWeight: 600,
            }}
          >
            ✓ Hand complete! Starting next round...
          </Typography>
        </Box>
      )}

      {/* Showdown Display - shows all player hands at bottom */}
      <ShowdownDisplay
        players={players}
        winners={winners}
        show={status === 'handComplete' && winners.length > 0}
      />
    </Box>
  );
});

PokerTable.displayName = 'PokerTable';

export default PokerTable;
