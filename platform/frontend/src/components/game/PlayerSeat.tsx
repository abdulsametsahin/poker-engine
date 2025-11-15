import React, { memo } from 'react';
import { Box, Stack, Typography } from '@mui/material';
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
  isPaused?: boolean;
}

export const PlayerSeat: React.FC<PlayerSeatProps> = memo(({
  player,
  position,
  isActive,
  isCurrentUser = false,
  actionDeadline,
  isPaused = false,
}) => {
  if (!player) {
    return (
      <Box
        sx={{
          width: 110,
          height: 70,
          borderRadius: RADIUS.md,
          border: `2px dashed ${COLORS.border.main}40`,
          background: 'rgba(255, 255, 255, 0.02)',
          backdropFilter: 'blur(8px)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          opacity: 0.3,
          transition: TRANSITIONS.normal,
          '&:hover': {
            opacity: 0.5,
            borderColor: `${COLORS.border.main}60`,
          },
        }}
      >
        <Typography
          variant="caption"
          sx={{
            color: COLORS.text.disabled,
            fontSize: '10px',
            fontWeight: 600,
            letterSpacing: '0.1em',
            textAlign: 'center',
          }}
        >
          OPEN SEAT
        </Typography>
      </Box>
    );
  }

  return (
    <Box
      sx={{
        width: { xs: 85, sm: 95, md: 110, lg: 120 },
        minHeight: { xs: 65, sm: 70, md: 75, lg: 80 },
        position: 'relative',
        transition: TRANSITIONS.normal,
        filter: player.folded ? 'grayscale(0.7) opacity(0.6)' : 'none',
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
            width: '120%',
            zIndex: 30,
          }}
        >
          <ActionTimer deadline={actionDeadline} totalTime={30} isPaused={isPaused} />
        </Box>
      )}

      {/* Main container - Premium glass morphism design */}
      <Box
        sx={{
          position: 'relative',
          height: '100%',
          borderRadius: RADIUS.md,
          background: isActive
            ? `linear-gradient(145deg, rgba(124, 58, 237, 0.25) 0%, rgba(6, 182, 212, 0.2) 100%)`
            : isCurrentUser
            ? `linear-gradient(145deg, rgba(251, 191, 36, 0.15) 0%, rgba(245, 158, 11, 0.1) 100%)`
            : 'rgba(0, 0, 0, 0.6)',
          backdropFilter: 'blur(16px)',
          border: isActive
            ? `2px solid ${COLORS.primary.main}`
            : isCurrentUser
            ? `2px solid ${COLORS.accent.main}`
            : `1.5px solid rgba(255, 255, 255, 0.12)`,
          boxShadow: isActive
            ? `0 0 20px ${COLORS.primary.glow}, 0 4px 16px rgba(0, 0, 0, 0.5), inset 0 1px 0 rgba(255, 255, 255, 0.15)`
            : isCurrentUser
            ? `0 0 16px ${COLORS.accent.glow}40, 0 4px 12px rgba(0, 0, 0, 0.4), inset 0 1px 0 rgba(255, 255, 255, 0.1)`
            : '0 4px 12px rgba(0, 0, 0, 0.4), inset 0 1px 0 rgba(255, 255, 255, 0.05)',
          px: 2,
          py: 1.25,
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'center',
          gap: 0.5,
          transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
          overflow: 'visible',
          
          // Animated border glow for active player
          ...(isActive && {
            '&::before': {
              content: '""',
              position: 'absolute',
              inset: -2,
              borderRadius: RADIUS.md,
              background: `linear-gradient(135deg, ${COLORS.primary.main}, ${COLORS.secondary.main}, ${COLORS.primary.main})`,
              backgroundSize: '200% 200%',
              animation: 'borderRotate 3s linear infinite',
              zIndex: -1,
              opacity: 0.6,
            },
            '@keyframes borderRotate': {
              '0%': { backgroundPosition: '0% 50%' },
              '50%': { backgroundPosition: '100% 50%' },
              '100%': { backgroundPosition: '0% 50%' },
            },
          }),
          
          '&:hover': {
            transform: player.folded ? 'none' : 'translateY(-2px)',
            boxShadow: isActive
              ? `0 0 24px ${COLORS.primary.glow}, 0 6px 20px rgba(0, 0, 0, 0.6)`
              : isCurrentUser
              ? `0 0 20px ${COLORS.accent.glow}60, 0 6px 16px rgba(0, 0, 0, 0.5)`
              : '0 6px 16px rgba(0, 0, 0, 0.5)',
          },
        }}
      >
        {/* Status badges - Enhanced design */}
        {player.folded && (
          <Box
            sx={{
              position: 'absolute',
              top: -10,
              right: -10,
              px: 1.5,
              py: 0.5,
              borderRadius: RADIUS.sm,
              background: `linear-gradient(135deg, ${COLORS.danger.main}, ${COLORS.danger.dark})`,
              border: `2px solid ${COLORS.background.primary}`,
              boxShadow: `0 0 12px ${COLORS.danger.glow}, 0 2px 8px rgba(0, 0, 0, 0.4)`,
              fontSize: '9px',
              fontWeight: 800,
              letterSpacing: '0.05em',
              color: 'white',
              zIndex: 10,
            }}
          >
            FOLDED
          </Box>
        )}
        {player.all_in && !player.folded && (
          <Box
            sx={{
              position: 'absolute',
              top: -10,
              right: -10,
              px: 1.5,
              py: 0.5,
              borderRadius: RADIUS.sm,
              background: `linear-gradient(135deg, ${COLORS.warning.main}, ${COLORS.warning.dark})`,
              border: `2px solid ${COLORS.background.primary}`,
              boxShadow: `0 0 12px ${COLORS.warning.glow}, 0 2px 8px rgba(0, 0, 0, 0.4)`,
              fontSize: '9px',
              fontWeight: 800,
              letterSpacing: '0.05em',
              color: 'white',
              zIndex: 10,
              animation: 'pulse 1.5s ease-in-out infinite',
              '@keyframes pulse': {
                '0%, 100%': { transform: 'scale(1)' },
                '50%': { transform: 'scale(1.05)' },
              },
            }}
          >
            ALL IN
          </Box>
        )}

        {/* Username with icon */}
        <Stack direction="row" alignItems="center" spacing={0.5} justifyContent="center">
          {isCurrentUser && (
            <Box
              sx={{
                width: 6,
                height: 6,
                borderRadius: '50%',
                background: COLORS.accent.main,
                boxShadow: `0 0 8px ${COLORS.accent.glow}`,
                animation: 'pulse 2s ease-in-out infinite',
              }}
            />
          )}
          <Typography
            variant="caption"
            sx={{
              color: isCurrentUser ? COLORS.accent.light : COLORS.text.primary,
              fontSize: '12px',
              fontWeight: isCurrentUser ? 800 : 700,
              textAlign: 'center',
              maxWidth: '100%',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
              textShadow: '0 1px 4px rgba(0, 0, 0, 0.5)',
              letterSpacing: '0.02em',
            }}
          >
            {formatUsername(player.username || player.user_id.slice(0, 8))}
          </Typography>
        </Stack>

        {/* Chips with icon */}
        <Box
          sx={{
            px: 1.5,
            py: 0.5,
            borderRadius: RADIUS.sm,
            background: 'rgba(0, 0, 0, 0.4)',
            border: `1px solid rgba(255, 255, 255, 0.1)`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 0.5,
          }}
        >
          <Box
            sx={{
              width: 8,
              height: 8,
              borderRadius: '50%',
              background: `linear-gradient(135deg, ${COLORS.accent.main}, ${COLORS.accent.dark})`,
              boxShadow: `0 0 6px ${COLORS.accent.glow}40`,
            }}
          />
          <Typography
            variant="caption"
            sx={{
              color: COLORS.text.primary,
              fontSize: '11px',
              fontWeight: 700,
              fontFamily: 'monospace',
              textShadow: '0 1px 2px rgba(0, 0, 0, 0.5)',
            }}
          >
            ${player.chips.toLocaleString()}
          </Typography>
        </Box>

        {/* Player cards - Enhanced positioning for current user */}
        {isCurrentUser && player.cards && player.cards.length > 0 && (
          <Stack
            direction="row"
            spacing={0.75}
            sx={{
              position: 'absolute',
              bottom: -60,
              left: '50%',
              transform: 'translateX(-50%)',
              zIndex: 20,
              filter: 'drop-shadow(0 4px 12px rgba(0, 0, 0, 0.6))',
            }}
          >
            {player.cards.map((card, idx) => (
              <Box
                key={idx}
                sx={{
                  animation: `cardFloat 2s ease-in-out infinite`,
                  animationDelay: `${idx * 0.2}s`,
                  '@keyframes cardFloat': {
                    '0%, 100%': { transform: 'translateY(0)' },
                    '50%': { transform: 'translateY(-4px)' },
                  },
                }}
              >
                <PlayingCard
                  card={typeof card === 'string' ? card : `${card.rank}${card.suit}`}
                  size="medium"
                  dealAnimation={false}
                />
              </Box>
            ))}
          </Stack>
        )}

        {/* Current bet - Enhanced chip stack design */}
        {player.current_bet > 0 && (
          <Box
            sx={{
              position: 'absolute',
              // Position bet to the right for current user (to avoid card overlap)
              // Position bet below for other players
              ...(isCurrentUser ? {
                right: -70,
                top: '50%',
                transform: 'translateY(-50%)',
              } : {
                bottom: -24,
                left: '50%',
                transform: 'translateX(-50%)',
              }),
              px: 2,
              py: 0.75,
              borderRadius: RADIUS.md,
              background: `linear-gradient(135deg, ${COLORS.info.main}, ${COLORS.info.dark})`,
              border: `2px solid ${COLORS.info.light}40`,
              boxShadow: `0 0 16px ${COLORS.info.glow}, 0 4px 12px rgba(0, 0, 0, 0.5)`,
              zIndex: 15,
              minWidth: 60,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 0.5,
            }}
          >
            <Box
              sx={{
                width: 10,
                height: 10,
                borderRadius: '50%',
                background: 'rgba(255, 255, 255, 0.9)',
                boxShadow: '0 0 8px rgba(255, 255, 255, 0.6)',
              }}
            />
            <Typography
              variant="caption"
              sx={{
                color: 'white',
                fontSize: '11px',
                fontWeight: 800,
                fontFamily: 'monospace',
                textShadow: '0 2px 4px rgba(0, 0, 0, 0.6)',
                letterSpacing: '0.03em',
              }}
            >
              ${player.current_bet}
            </Typography>
          </Box>
        )}

        {/* Last action - Subtle indicator */}
        {player.last_action && !isActive && !player.current_bet && (
          <Typography
            variant="caption"
            sx={{
              position: 'absolute',
              bottom: -20,
              left: '50%',
              transform: 'translateX(-50%)',
              fontSize: '9px',
              fontWeight: 600,
              color: COLORS.text.disabled,
              textTransform: 'uppercase',
              whiteSpace: 'nowrap',
              letterSpacing: '0.05em',
              opacity: 0.7,
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
