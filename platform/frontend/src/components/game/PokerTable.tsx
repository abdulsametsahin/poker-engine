import React, { memo } from 'react';
import { Box, Stack, Typography } from '@mui/material';
import { PlayerSeat } from './PlayerSeat';
import { PlayingCard } from './PlayingCard';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { COLORS, RADIUS, SPACING, TRANSITIONS } from '../../constants';
import { Player } from '../../types';
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
}

interface PokerTableProps {
  tableState: TableState | null;
  currentUserId?: string;
}

export const PokerTable: React.FC<PokerTableProps> = memo(({
  tableState,
  currentUserId,
}) => {
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

  const {
    players = [],
    community_cards = [],
    pot = 0,
    current_turn,
    status,
    betting_round,
    current_bet = 0,
    action_deadline,
  } = tableState;

  // Calculate positions for circular layout
  const getPlayerPosition = (index: number, total: number) => {
    // Arrange players in a circle
    const angle = (index / total) * 2 * Math.PI - Math.PI / 2; // Start from top
    const radiusX = 45; // Horizontal spread
    const radiusY = 35; // Vertical spread

    return {
      left: `${50 + radiusX * Math.cos(angle)}%`,
      top: `${50 + radiusY * Math.sin(angle)}%`,
      transform: 'translate(-50%, -50%)',
    };
  };

  return (
    <Box
      sx={{
        height: '100%',
        position: 'relative',
        borderRadius: RADIUS.lg,
        background: `linear-gradient(135deg, rgba(11, 107, 62, 0.4) 0%, rgba(6, 78, 59, 0.3) 100%)`,
        border: `3px solid ${COLORS.success.main}40`,
        overflow: 'hidden',
        '&::before': {
          content: '""',
          position: 'absolute',
          inset: 0,
          background: 'radial-gradient(ellipse at center, rgba(16, 185, 129, 0.08) 0%, transparent 70%)',
          pointerEvents: 'none',
        },
      }}
    >
      {/* Felt texture pattern */}
      <Box
        sx={{
          position: 'absolute',
          inset: 0,
          backgroundImage: `
            repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(0,0,0,0.03) 2px, rgba(0,0,0,0.03) 4px),
            repeating-linear-gradient(90deg, transparent, transparent 2px, rgba(0,0,0,0.03) 2px, rgba(0,0,0,0.03) 4px)
          `,
          pointerEvents: 'none',
        }}
      />

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
            px: 3,
            py: 1.5,
            borderRadius: RADIUS.md,
            background: 'rgba(0, 0, 0, 0.6)',
            backdropFilter: 'blur(10px)',
            border: `2px solid ${COLORS.accent.main}`,
            boxShadow: `0 4px 12px rgba(0, 0, 0, 0.5), 0 0 20px ${COLORS.accent.glow}`,
            minWidth: 120,
            textAlign: 'center',
          }}
        >
          <Typography
            variant="caption"
            sx={{
              color: COLORS.text.secondary,
              fontSize: '10px',
              fontWeight: 600,
              letterSpacing: '0.1em',
            }}
          >
            POT
          </Typography>
          <Typography
            variant="h4"
            sx={{
              color: COLORS.accent.main,
              fontWeight: 900,
              fontSize: '28px',
              lineHeight: 1.2,
              fontFamily: 'monospace',
            }}
          >
            ${pot}
          </Typography>
        </Box>

        {/* Community cards */}
        {community_cards.length > 0 && (
          <Stack
            direction="row"
            spacing={1}
            sx={{
              px: 2,
              py: 1.5,
              borderRadius: RADIUS.md,
              background: 'rgba(0, 0, 0, 0.5)',
              backdropFilter: 'blur(10px)',
              border: `1px solid ${COLORS.border.heavy}`,
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
              px: 2.5,
              py: 1,
              borderRadius: RADIUS.md,
              background: `linear-gradient(135deg, ${COLORS.primary.main}40 0%, ${COLORS.secondary.main}40 100%)`,
              backdropFilter: 'blur(10px)',
              border: `1px solid ${COLORS.primary.main}`,
              boxShadow: `0 0 12px ${COLORS.primary.glow}`,
            }}
          >
            <Typography
              variant="caption"
              sx={{
                color: COLORS.primary.light,
                fontSize: '11px',
                fontWeight: 700,
                letterSpacing: '0.1em',
              }}
            >
              {getBettingRoundName(betting_round)}
            </Typography>
            {current_bet > 0 && (
              <Typography
                variant="caption"
                sx={{
                  color: COLORS.text.secondary,
                  fontSize: '9px',
                  display: 'block',
                  mt: 0.25,
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
    </Box>
  );
});

PokerTable.displayName = 'PokerTable';

export default PokerTable;
