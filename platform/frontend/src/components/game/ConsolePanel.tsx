import React, { memo, useEffect, useRef } from 'react';
import { Box, Stack, Typography } from '@mui/material';
import { Terminal } from '@mui/icons-material';
import { COLORS, RADIUS } from '../../constants';

interface LogEntry {
  id: string;
  timestamp: Date;
  level: 'info' | 'warning' | 'error' | 'success' | 'debug';
  category: string;
  message: string;
}

interface ConsolePanelProps {
  logs: LogEntry[];
  maxLogs?: number;
}

export const ConsolePanel: React.FC<ConsolePanelProps> = memo(({ logs, maxLogs = 100 }) => {
  const scrollRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new logs arrive
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [logs]);

  // Keep only the last maxLogs entries
  const displayLogs = logs.slice(-maxLogs);

  return (
    <Box
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
        bgcolor: '#1a1a1a',
        fontFamily: 'monospace',
      }}
    >
      {/* Console output */}
      <Stack
        ref={scrollRef}
        spacing={0}
        sx={{
          flex: 1,
          p: 1,
          overflowY: 'auto',
          overflowX: 'hidden',
          '&::-webkit-scrollbar': {
            width: 6,
          },
          '&::-webkit-scrollbar-track': {
            background: '#0a0a0a',
            borderRadius: RADIUS.sm,
          },
          '&::-webkit-scrollbar-thumb': {
            background: '#333',
            borderRadius: RADIUS.sm,
            '&:hover': {
              background: '#444',
            },
          },
        }}
      >
        {displayLogs.length === 0 ? (
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%',
              opacity: 0.5,
              gap: 1,
            }}
          >
            <Terminal sx={{ fontSize: 32, color: '#666' }} />
            <Typography
              variant="caption"
              sx={{
                color: '#666',
                fontSize: 11,
                fontFamily: 'monospace',
              }}
            >
              Waiting for logs...
            </Typography>
          </Box>
        ) : (
          displayLogs.map((log) => (
            <Box
              key={log.id}
              sx={{
                py: 0.5,
                px: 1,
                borderBottom: '1px solid #222',
                '&:hover': {
                  bgcolor: '#252525',
                },
              }}
            >
              <Stack direction="row" spacing={1} alignItems="flex-start">
                {/* Timestamp */}
                <Typography
                  variant="caption"
                  sx={{
                    color: '#666',
                    fontSize: 9,
                    fontFamily: 'monospace',
                    minWidth: 60,
                    flexShrink: 0,
                  }}
                >
                  {log.timestamp.toLocaleTimeString([], {
                    hour: '2-digit',
                    minute: '2-digit',
                    second: '2-digit',
                  })}
                </Typography>

                {/* Level indicator */}
                <Box
                  sx={{
                    width: 6,
                    height: 6,
                    borderRadius: '50%',
                    bgcolor: getLevelColor(log.level),
                    mt: 0.5,
                    flexShrink: 0,
                  }}
                />

                {/* Category */}
                <Typography
                  variant="caption"
                  sx={{
                    color: getCategoryColor(log.category),
                    fontSize: 10,
                    fontFamily: 'monospace',
                    fontWeight: 600,
                    minWidth: 100,
                    flexShrink: 0,
                  }}
                >
                  [{log.category}]
                </Typography>

                {/* Message */}
                <Typography
                  variant="caption"
                  sx={{
                    color: getLogLevelTextColor(log.level),
                    fontSize: 10,
                    fontFamily: 'monospace',
                    wordBreak: 'break-word',
                    flex: 1,
                  }}
                >
                  {log.message}
                </Typography>
              </Stack>
            </Box>
          ))
        )}
      </Stack>
    </Box>
  );
});

function getLevelColor(level: string): string {
  switch (level) {
    case 'error':
      return '#ff4444';
    case 'warning':
      return '#ffaa00';
    case 'success':
      return '#00ff00';
    case 'info':
      return '#00aaff';
    case 'debug':
      return '#aaaaaa';
    default:
      return '#888888';
  }
}

function getLogLevelTextColor(level: string): string {
  switch (level) {
    case 'error':
      return '#ff6666';
    case 'warning':
      return '#ffcc44';
    case 'success':
      return '#66ff66';
    case 'info':
      return '#66ccff';
    case 'debug':
      return '#cccccc';
    default:
      return '#aaaaaa';
  }
}

function getCategoryColor(category: string): string {
  switch (category.toUpperCase()) {
    case 'ACTION':
      return '#00ffaa';
    case 'ENGINE_EVENT':
      return '#aa88ff';
    case 'TOURNAMENT':
      return '#ffaa00';
    case 'CASH_GAME':
      return '#00aaff';
    default:
      return '#88aaff';
  }
}

ConsolePanel.displayName = 'ConsolePanel';
