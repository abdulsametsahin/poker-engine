import React, { memo } from 'react';
import { Box, Stack, Typography } from '@mui/material';
import { Badge } from '../common/Badge';
import { PlayingCard } from './PlayingCard';
import ActionTimer from '../ActionTimer';
import { COLORS, TRANSITIONS, RADIUS } from '../../constants';
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
          width: 90,
          height: 60,
          borderRadius: RADIUS.sm,
          border: `1px dashed ${COLORS.border.main}`,
          background: 'rgba(255, 255, 255, 0.02)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          opacity: 0.3,
        }}
      >
        <Typography
          variant="caption"
          sx={{
            color: COLORS.text.disabled,
            fontSize: '9px',
            textAlign: 'center',
          }}
        >
          EMPTY
        </Typography>
      </Box>
    );
  }

  const status = player.folded ? 'folded' : player.all_in ? 'all_in' : 'active';

  return (
    <Box
      sx={{
        width: { xs: 70, sm: 80, md: 90, lg: 100 },
        height: { xs: 50, sm: 55, md: 60, lg: 65 },
        position: 'relative',
        transition: TRANSITIONS.normal,
        opacity: player.folded ? 0.5 : 1,
      }}
    >
      {/* Action Timer */}
      {isActive && actionDeadline && (
        <Box
          sx={{
            position: 'absolute',
            top: -28,
            left: '50%',
            transform: 'translateX(-50%)',
            width: '110%',
            zIndex: 30,
          }}
        >
          <ActionTimer deadline={actionDeadline} totalTime={30} />
        </Box>
      )}

      {/* Main container - minimal design */}
      <Box
        sx={{
          position: 'relative',
          height: '100%',
          borderRadius: RADIUS.sm,
          background: isActive
            ? `linear-gradient(135deg, rgba(124, 58, 237, 0.15) 0%, rgba(6, 182, 212, 0.15) 100%)`
            : 'rgba(0, 0, 0, 0.4)',
          border: isActive
            ? `2px solid ${COLORS.primary.main}`
            : isCurrentUser
            ? `2px solid ${COLORS.accent.main}`
            : `1px solid rgba(255, 255, 255, 0.15)`,
          boxShadow: isActive
            ? `0 0 12px ${COLORS.primary.glow}`
            : 'none',
          px: 1.5,
          py: 1,
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'center',
          gap: 0.5,
          transition: TRANSITIONS.normal,
        }}
      >
        {/* Status badges */}
        {player.folded && (
          <Badge variant="danger" size="small" sx={{ position: 'absolute', top: -8, right: -8, fontSize: '8px' }}>
            FOLD
          </Badge>
        )}
        {player.all_in && !player.folded && (
          <Badge variant="warning" size="small" sx={{ position: 'absolute', top: -8, right: -8, fontSize: '8px' }}>
            ALL IN
          </Badge>
        )}
        {player.is_dealer && (
          <Box
            sx={{
              position: 'absolute',
              top: -10,
              left: -10,
              width: 20,
              height: 20,
              borderRadius: '50%',
              background: COLORS.accent.main,
              border: `2px solid ${COLORS.background.primary}`,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: '10px',
              fontWeight: 'bold',
            }}
          >
            D
          </Box>
        )}

        {/* Username */}
        <Typography
          variant="caption"
          sx={{
            color: isCurrentUser ? COLORS.accent.main : COLORS.text.primary,
            fontSize: '11px',
            fontWeight: isCurrentUser ? 700 : 600,
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
        <Typography
          variant="caption"
          sx={{
            color: COLORS.text.secondary,
            fontSize: '10px',
            fontWeight: 600,
            textAlign: 'center',
          }}
        >
          ${player.chips}
        </Typography>

        {/* Player cards - show only for current user */}
        {isCurrentUser && player.cards && player.cards.length > 0 && (
          <Stack
            direction="row"
            spacing={0.5}
            sx={{
              position: 'absolute',
              bottom: -45,
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
              bottom: -20,
              left: '50%',
              transform: 'translateX(-50%)',
              px: 1,
              py: 0.25,
              borderRadius: RADIUS.sm,
              background: COLORS.info.main,
              boxShadow: `0 2px 6px ${COLORS.info.glow}`,
              zIndex: 5,
            }}
          >
            <Typography
              variant="caption"
              sx={{
                color: COLORS.text.primary,
                fontSize: '9px',
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
              bottom: -18,
              left: '50%',
              transform: 'translateX(-50%)',
              fontSize: '8px',
              color: COLORS.text.disabled,
              textTransform: 'uppercase',
              whiteSpace: 'nowrap',
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
