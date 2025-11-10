import React, { memo } from 'react';
import { Box, Stack, Typography } from '@mui/material';
import { History } from '@mui/icons-material';
import { COLORS, RADIUS } from '../../constants';

interface HistoryEntry {
  id: string;
  playerName: string;
  action: string;
  amount?: number;
  timestamp: Date;
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
        borderRadius: RADIUS.md,
        background: `linear-gradient(135deg, ${COLORS.background.paper}f5 0%, ${COLORS.background.tertiary}f5 100%)`,
        backdropFilter: 'blur(10px)',
        border: `1px solid ${COLORS.border.main}`,
        overflow: 'hidden',
      }}
    >
      {/* Header */}
      <Box
        sx={{
          px: 2,
          py: 1.5,
          borderBottom: `1px solid ${COLORS.border.main}`,
          display: 'flex',
          alignItems: 'center',
          gap: 1,
        }}
      >
        <History sx={{ fontSize: 18, color: COLORS.primary.main }} />
        <Typography
          variant="body2"
          sx={{
            color: COLORS.text.primary,
            fontWeight: 700,
            fontSize: 13,
            letterSpacing: '0.05em',
            textTransform: 'uppercase',
          }}
        >
          History
        </Typography>
      </Box>

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
          history.map((entry) => (
            <Box
              key={entry.id}
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
                  {entry.playerName}
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
                  color: getActionColor(entry.action),
                  fontSize: 10,
                  fontWeight: 700,
                  textTransform: 'uppercase',
                }}
              >
                {entry.action}
                {entry.amount !== undefined && ` $${entry.amount}`}
              </Typography>
            </Box>
          ))
        )}
      </Stack>
    </Box>
  );
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
