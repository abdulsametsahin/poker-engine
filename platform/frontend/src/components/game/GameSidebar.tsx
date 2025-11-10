import React, { memo, useState } from 'react';
import { Box, Tabs, Tab } from '@mui/material';
import { History, Chat } from '@mui/icons-material';
import { HistoryPanel } from './HistoryPanel';
import { ChatPanel } from './ChatPanel';
import { COLORS, RADIUS } from '../../constants';

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
  const [activeTab, setActiveTab] = useState(1); // Default to Chat

  return (
    <Box
      sx={{
        width: 280,
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        borderRadius: RADIUS.md,
        background: 'rgba(0, 0, 0, 0.3)',
        backdropFilter: 'blur(10px)',
        border: `1px solid ${COLORS.border.main}`,
        overflow: 'hidden',
      }}
    >
      {/* Tabs */}
      <Tabs
        value={activeTab}
        onChange={(_, newValue) => setActiveTab(newValue)}
        sx={{
          minHeight: 40,
          borderBottom: `1px solid ${COLORS.border.main}`,
          '& .MuiTabs-indicator': {
            backgroundColor: COLORS.info.main,
          },
        }}
      >
        <Tab
          icon={<History sx={{ fontSize: 16 }} />}
          label="History"
          sx={{
            minHeight: 40,
            fontSize: 11,
            fontWeight: 600,
            color: COLORS.text.secondary,
            '&.Mui-selected': {
              color: COLORS.info.main,
            },
          }}
        />
        <Tab
          icon={<Chat sx={{ fontSize: 16 }} />}
          label="Chat"
          sx={{
            minHeight: 40,
            fontSize: 11,
            fontWeight: 600,
            color: COLORS.text.secondary,
            '&.Mui-selected': {
              color: COLORS.info.main,
            },
          }}
        />
      </Tabs>

      {/* Panel content */}
      <Box sx={{ flex: 1, overflow: 'hidden', p: 1 }}>
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
