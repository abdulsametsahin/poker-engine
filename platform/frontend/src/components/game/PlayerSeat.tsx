import React, { memo } from 'react';
import { Box, Stack, Typography } from '@mui/material';
import { Avatar } from '../common/Avatar';
import { Chip } from '../common/Chip';
import { Badge } from '../common/Badge';
import { PlayingCard } from './PlayingCard';
import ActionTimer from '../ActionTimer';
import { COLORS, SPACING, TRANSITIONS, RADIUS } from '../../constants';
import { Player } from '../../types';
import { formatUsername } from '../../utils';

interface PlayerSeatProps {
  player: Player | null;
  position: number;
  isActive: boolean;
  isCurrentUser?: boolean;
  actionDeadline?: string;
}

export const PlayerSeat: React.FC<PlayerSeatProps> = memo(({
  player,
  position,
  isActive,
  isCurrentUser = false,
  actionDeadline,
}) => {
  if (!player) {
    return (
      <Box
        sx={{
          width: 120,
          height: 160,
          borderRadius: RADIUS.md,
          border: `2px dashed ${COLORS.border.main}`,
          background: 'rgba(255, 255, 255, 0.02)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          opacity: 0.4,
        }}
      >
        <Typography
          variant="caption"
          sx={{
            color: COLORS.text.disabled,
            fontSize: '10px',
            textAlign: 'center',
          }}
        >
          SEAT {position}
          <br />
          EMPTY
        </Typography>
      </Box>
    );
  }

  const status = player.folded ? 'folded' : player.all_in ? 'all_in' : 'active';

  return (
    <Box
      sx={{
        width: 120,
        height: 160,
        position: 'relative',
        transition: TRANSITIONS.normal,
        opacity: player.folded ? 0.5 : 1,
      }}
    >
      {/* Glow effect when active */}
      {isActive && (
        <Box
          sx={{
            position: 'absolute',
            inset: -4,
            borderRadius: RADIUS.md,
            background: `linear-gradient(135deg, ${COLORS.primary.main} 0%, ${COLORS.secondary.main} 100%)`,
            opacity: 0.3,
            filter: 'blur(8px)',
            '@keyframes pulse-glow': {
              '0%, 100%': {
                opacity: 0.3,
              },
              '50%': {
                opacity: 0.6,
              },
            },
            animation: 'pulse-glow 2s ease-in-out infinite',
          }}
        />
      )}

      {/* Main container */}
      <Box
        sx={{
          position: 'relative',
          height: '100%',
          borderRadius: RADIUS.md,
          background: isActive
            ? `linear-gradient(135deg, rgba(124, 58, 237, 0.15) 0%, rgba(6, 182, 212, 0.15) 100%)`
            : 'rgba(255, 255, 255, 0.05)',
          backdropFilter: 'blur(10px)',
          border: isActive
            ? `2px solid ${COLORS.primary.main}`
            : isCurrentUser
            ? `2px solid ${COLORS.accent.main}`
            : `1px solid ${COLORS.border.main}`,
          boxShadow: isActive
            ? `0 0 20px ${COLORS.primary.glow}`
            : isCurrentUser
            ? `0 0 12px ${COLORS.accent.glow}`
            : 'none',
          p: 1.5,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: 1,
        }}
      >
        {/* Action Timer */}
        {isActive && actionDeadline && (
          <Box
            sx={{
              position: 'absolute',
              top: -32,
              left: '50%',
              transform: 'translateX(-50%)',
              width: '110%',
              zIndex: 30,
            }}
          >
            <ActionTimer deadline={actionDeadline} totalTime={30} />
          </Box>
        )}

        {/* Status badges */}
        {player.folded && (
          <Badge variant="danger" size="small" sx={{ position: 'absolute', top: 4, right: 4 }}>
            FOLD
          </Badge>
        )}
        {player.all_in && !player.folded && (
          <Badge variant="warning" size="small" sx={{ position: 'absolute', top: 4, right: 4 }}>
            ALL-IN
          </Badge>
        )}
        {isCurrentUser && !isActive && !player.folded && (
          <Badge variant="primary" size="small" sx={{ position: 'absolute', top: 4, left: 4 }}>
            YOU
          </Badge>
        )}

        {/* Avatar with dealer button */}
        <Box sx={{ position: 'relative' }}>
          <Avatar
            username={player.username || player.user_id}
            size="medium"
            online={isActive}
            dealer={player.is_dealer}
          />
        </Box>

        {/* Username */}
        <Typography
          variant="caption"
          sx={{
            color: COLORS.text.primary,
            fontSize: '11px',
            fontWeight: 600,
            textAlign: 'center',
            maxWidth: '100%',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          {formatUsername(player.username || player.user_id.slice(0, 8))}
        </Typography>

        {/* Chips */}
        <Chip amount={player.chips} variant="default" size="small" />

        {/* Player cards - show only for current user */}
        {isCurrentUser && player.cards && player.cards.length > 0 && (
          <Stack
            direction="row"
            spacing={0.5}
            sx={{
              position: 'absolute',
              bottom: -50,
              left: '50%',
              transform: 'translateX(-50%)',
              zIndex: 10,
            }}
          >
            {player.cards.map((card, idx) => (
              <PlayingCard
                key={idx}
                card={typeof card === 'string' ? card : `${card.rank}${card.suit}`}
                size="small"
                dealAnimation={false}
              />
            ))}
          </Stack>
        )}

        {/* Current bet */}
        {player.current_bet > 0 && (
          <Box
            sx={{
              position: 'absolute',
              bottom: -24,
              left: '50%',
              transform: 'translateX(-50%)',
              px: 1,
              py: 0.5,
              borderRadius: RADIUS.sm,
              background: `linear-gradient(135deg, ${COLORS.info.main} 0%, ${COLORS.info.dark} 100%)`,
              border: `1px solid ${COLORS.info.main}`,
              boxShadow: `0 2px 8px ${COLORS.info.glow}`,
              zIndex: 5,
            }}
          >
            <Typography
              variant="caption"
              sx={{
                color: COLORS.text.primary,
                fontSize: '10px',
                fontWeight: 700,
              }}
            >
              ${player.current_bet}
            </Typography>
          </Box>
        )}

        {/* Last action */}
        {player.last_action && !isActive && (
          <Typography
            variant="caption"
            sx={{
              position: 'absolute',
              bottom: 4,
              fontSize: '9px',
              color: COLORS.text.secondary,
              textTransform: 'uppercase',
            }}
          >
            {player.last_action}
          </Typography>
        )}
      </Box>
    </Box>
  );
});

PlayerSeat.displayName = 'PlayerSeat';
