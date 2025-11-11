import React, { memo, useState } from 'react';
import { Box, Tabs, Tab } from '@mui/material';
import { History, Chat } from '@mui/icons-material';
import { HistoryPanel } from './HistoryPanel';
import { ChatPanel } from './ChatPanel';
import { COLORS, RADIUS, TRANSITIONS } from '../../constants';

interface HistoryEntry {
  id: string;
  playerName: string;
  action: string;
  amount?: number;
  timestamp: Date;
}

interface ChatMessage {
  id: string;
  userId: string;
  username: string;
  message: string;
  timestamp: Date;
}

interface GameSidebarProps {
  history: HistoryEntry[];
  messages: ChatMessage[];
  currentUserId?: string;
  onSendMessage: (message: string) => void;
}

export const GameSidebar: React.FC<GameSidebarProps> = memo(({
  history,
  messages,
  currentUserId,
  onSendMessage,
}) => {
  const [activeTab, setActiveTab] = useState(0); // Default to History

  // Count unread messages (simple heuristic - in production, track properly)
  const unreadCount = 0; // You can implement proper unread tracking

  return (
    <Box
      sx={{
        width: 340,
        minWidth: 340,
        display: 'flex',
        flexDirection: 'column',
        borderLeft: `1px solid ${COLORS.border.main}`,
        background: 'rgba(10, 10, 10, 0.8)',
        backdropFilter: 'blur(10px)',
        overflow: 'hidden',
      }}
    >
      {/* Header with tabs */}
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          px: 2,
          py: 1.5,
          borderBottom: `1px solid ${COLORS.border.main}`,
          background: `linear-gradient(135deg, ${COLORS.primary.main}15 0%, ${COLORS.secondary.main}15 100%)`,
        }}
      >
        <Tabs
          value={activeTab}
          onChange={(_, newValue) => setActiveTab(newValue)}
          sx={{
            minHeight: 40,
            width: '100%',
            '& .MuiTabs-indicator': {
              backgroundColor: COLORS.primary.main,
              height: 3,
            },
          }}
        >
          <Tab
            icon={<History sx={{ fontSize: 18 }} />}
            label="History"
            sx={{
              minHeight: 40,
              fontSize: 12,
              fontWeight: 600,
              color: COLORS.text.secondary,
              transition: TRANSITIONS.fast,
              '&.Mui-selected': {
                color: COLORS.primary.main,
              },
              '&:hover': {
                color: COLORS.text.primary,
              },
            }}
          />
          <Tab
            icon={<Chat sx={{ fontSize: 18 }} />}
            label="Chat"
            sx={{
              minHeight: 40,
              fontSize: 12,
              fontWeight: 600,
              color: COLORS.text.secondary,
              transition: TRANSITIONS.fast,
              '&.Mui-selected': {
                color: COLORS.primary.main,
              },
              '&:hover': {
                color: COLORS.text.primary,
              },
            }}
          />
        </Tabs>
      </Box>

      {/* Panel content */}
      <Box
        sx={{
          flex: 1,
          overflow: 'hidden',
          p: 2,
        }}
      >
        {activeTab === 0 ? (
          <HistoryPanel history={history} />
        ) : (
          <ChatPanel
            messages={messages}
            currentUserId={currentUserId}
            onSendMessage={onSendMessage}
          />
        )}
      </Box>
    </Box>
  );
});

GameSidebar.displayName = 'GameSidebar';
