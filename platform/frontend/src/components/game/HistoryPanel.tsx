import React, { memo } from 'react';
import { Box, Stack, Typography } from '@mui/material';
import { History, PlayArrow, Casino, EmojiEvents, Visibility } from '@mui/icons-material';
import { COLORS, RADIUS } from '../../constants';

// Enhanced history entry to support all event types
interface HistoryEntry {
  id: string;
  eventType?: 'player_action' | 'hand_started' | 'round_advanced' | 'hand_complete' | 'game_complete' | 'showdown';
  playerName?: string;
  action?: string;
  amount?: number;
  timestamp: Date;
  metadata?: any;
}

interface HistoryPanelProps {
  history: HistoryEntry[];
}

export const HistoryPanel: React.FC<HistoryPanelProps> = memo(({ history }) => {
  return (
    <Box
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
      }}
    >
      {/* History list */}
      <Stack
        spacing={0.5}
        sx={{
          flex: 1,
          p: 1,
          overflowY: 'auto',
          '&::-webkit-scrollbar': {
            width: 6,
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
        {history.length === 0 ? (
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%',
              opacity: 0.5,
            }}
          >
            <Typography
              variant="caption"
              sx={{
                color: COLORS.text.secondary,
                fontSize: 11,
              }}
            >
              No actions yet
            </Typography>
          </Box>
        ) : (
          history.map((entry) => <HistoryEventItem key={entry.id} entry={entry} />)
        )}
      </Stack>
    </Box>
  );
});

// Component to render individual history events
const HistoryEventItem: React.FC<{ entry: HistoryEntry }> = memo(({ entry }) => {
  // Handle different event types
  const eventType = entry.eventType || (entry.action ? 'player_action' : 'unknown');

  switch (eventType) {
    case 'hand_started':
      return (
        <Box
          sx={{
            px: 1.5,
            py: 1,
            borderRadius: RADIUS.sm,
            background: `${COLORS.info.main}15`,
            border: `1px solid ${COLORS.info.main}40`,
          }}
        >
          <Stack direction="row" spacing={0.5} alignItems="center">
            <PlayArrow sx={{ fontSize: 12, color: COLORS.info.main }} />
            <Typography
              variant="caption"
              sx={{
                color: COLORS.info.main,
                fontSize: 10,
                fontWeight: 700,
                textTransform: 'uppercase',
              }}
            >
              New Hand #{entry.metadata?.hand_number || ''}
            </Typography>
          </Stack>
        </Box>
      );

    case 'round_advanced':
      const round = entry.metadata?.new_round || entry.metadata?.round || 'unknown';
      const cards = entry.metadata?.community_cards || [];
      // Format cards properly - handle both object and string formats
      const formatCard = (card: any): string => {
        if (typeof card === 'string') return card;
        if (typeof card === 'object' && card.rank && card.suit) {
          return `${card.rank}${card.suit}`;
        }
        return '';
      };
      const formattedCards = Array.isArray(cards)
        ? cards.map(formatCard).filter(Boolean).join(' ')
        : '';

      return (
        <Box
          sx={{
            px: 1.5,
            py: 1,
            borderRadius: RADIUS.sm,
            background: `${COLORS.warning.main}15`,
            border: `1px solid ${COLORS.warning.main}40`,
          }}
        >
          <Stack direction="row" spacing={0.5} alignItems="center">
            <Casino sx={{ fontSize: 12, color: COLORS.warning.main }} />
            <Typography
              variant="caption"
              sx={{
                color: COLORS.warning.main,
                fontSize: 10,
                fontWeight: 700,
                textTransform: 'uppercase',
              }}
            >
              {round}
            </Typography>
          </Stack>
          {formattedCards && (
            <Typography
              variant="caption"
              sx={{
                color: COLORS.text.secondary,
                fontSize: 9,
                ml: 2,
              }}
            >
              {formattedCards}
            </Typography>
          )}
        </Box>
      );

    case 'showdown':
      return (
        <Box
          sx={{
            px: 1.5,
            py: 1,
            borderRadius: RADIUS.sm,
            background: `${COLORS.accent.main}15`,
            border: `1px solid ${COLORS.accent.main}40`,
          }}
        >
          <Stack direction="row" spacing={0.5} alignItems="center">
            <Visibility sx={{ fontSize: 12, color: COLORS.accent.main }} />
            <Typography
              variant="caption"
              sx={{
                color: COLORS.accent.main,
                fontSize: 10,
                fontWeight: 700,
                textTransform: 'uppercase',
              }}
            >
              Showdown
            </Typography>
          </Stack>
        </Box>
      );

    case 'hand_complete':
      const winners = entry.metadata?.winners || [];
      const winnerNames = winners.map((w: any) => w.player_name || w.playerName).filter(Boolean);
      const pot = entry.metadata?.final_pot || entry.metadata?.pot || entry.amount;

      return (
        <Box
          sx={{
            px: 1.5,
            py: 1,
            borderRadius: RADIUS.sm,
            background: `${COLORS.success.main}15`,
            border: `1px solid ${COLORS.success.main}40`,
          }}
        >
          <Stack direction="row" spacing={0.5} alignItems="center">
            <EmojiEvents sx={{ fontSize: 12, color: COLORS.success.main }} />
            <Typography
              variant="caption"
              sx={{
                color: COLORS.success.main,
                fontSize: 10,
                fontWeight: 700,
              }}
            >
              {winnerNames.length > 0
                ? `${winnerNames.join(', ')} won ${pot ? `$${pot}` : ''}`
                : `Hand Complete ${pot ? `($${pot})` : ''}`
              }
            </Typography>
          </Stack>
        </Box>
      );

    case 'game_complete':
      const winnerName = entry.metadata?.winner_name || 'Player';
      const finalChips = entry.metadata?.final_chips || 0;

      return (
        <Box
          sx={{
            px: 1.5,
            py: 1,
            borderRadius: RADIUS.sm,
            background: `linear-gradient(135deg, ${COLORS.success.main}20 0%, ${COLORS.warning.main}20 100%)`,
            border: `2px solid ${COLORS.success.main}`,
          }}
        >
          <Stack direction="row" spacing={0.5} alignItems="center">
            <EmojiEvents sx={{ fontSize: 14, color: COLORS.success.main }} />
            <Typography
              variant="caption"
              sx={{
                color: COLORS.success.main,
                fontSize: 11,
                fontWeight: 800,
                textTransform: 'uppercase',
              }}
            >
              Game Complete!
            </Typography>
          </Stack>
          <Typography
            variant="caption"
            sx={{
              color: COLORS.text.primary,
              fontSize: 10,
              ml: 2,
              fontWeight: 600,
            }}
          >
            {winnerName} wins with ${finalChips}
          </Typography>
        </Box>
      );

    case 'player_action':
    default:
      // Standard player action (fold, call, raise, check, allin)
      return (
        <Box
          sx={{
            px: 1.5,
            py: 1,
            borderRadius: RADIUS.sm,
            background: `${COLORS.background.secondary}80`,
            border: `1px solid ${COLORS.border.main}`,
          }}
        >
          <Stack direction="row" justifyContent="space-between" alignItems="center">
            <Typography
              variant="caption"
              sx={{
                color: COLORS.text.primary,
                fontSize: 11,
                fontWeight: 600,
              }}
            >
              {entry.playerName || entry.metadata?.player_name || 'Player'}
            </Typography>
            <Typography
              variant="caption"
              sx={{
                color: COLORS.text.secondary,
                fontSize: 9,
              }}
            >
              {entry.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            </Typography>
          </Stack>
          <Typography
            variant="caption"
            sx={{
              color: getActionColor(entry.action || ''),
              fontSize: 10,
              fontWeight: 700,
              textTransform: 'uppercase',
            }}
          >
            {entry.action || 'action'}
            {entry.amount !== undefined && entry.amount > 0 && ` $${entry.amount}`}
          </Typography>
        </Box>
      );
  }
});

function getActionColor(action: string): string {
  switch (action.toLowerCase()) {
    case 'fold':
      return COLORS.danger.main;
    case 'call':
      return COLORS.success.main;
    case 'raise':
      return COLORS.warning.main;
    case 'check':
      return COLORS.info.main;
    case 'all_in':
    case 'allin':
      return COLORS.accent.main;
    default:
      return COLORS.text.secondary;
  }
}

HistoryPanel.displayName = 'HistoryPanel';
HistoryEventItem.displayName = 'HistoryEventItem';
