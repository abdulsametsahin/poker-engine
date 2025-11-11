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
    <>
      {/* Always visible panel */}
      <Box
        sx={{
          position: 'fixed',
          top: 16,
          right: 16,
          bottom: 16,
          width: 320,
          maxWidth: 'calc(100vw - 32px)',
          display: 'flex',
          flexDirection: 'column',
          borderRadius: RADIUS.lg,
          background: 'rgba(15, 15, 15, 0.95)',
          backdropFilter: 'blur(20px)',
          border: `2px solid ${COLORS.primary.main}`,
          boxShadow: `
            0 20px 60px rgba(0, 0, 0, 0.8),
            0 0 30px ${COLORS.primary.glow}
          `,
          overflow: 'hidden',
          zIndex: 100,
        }}
      >
        {/* Header with tabs */}
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            px: 1,
            py: 0.5,
            borderBottom: `1px solid ${COLORS.border.main}`,
            background: `linear-gradient(135deg, ${COLORS.primary.main}20 0%, ${COLORS.secondary.main}20 100%)`,
          }}
        >
          <Tabs
            value={activeTab}
            onChange={(_, newValue) => setActiveTab(newValue)}
            sx={{
              minHeight: 36,
              flex: 1,
              '& .MuiTabs-indicator': {
                backgroundColor: COLORS.info.main,
              },
            }}
          >
            <Tab
              icon={<History sx={{ fontSize: 16 }} />}
              label="History"
              sx={{
                minHeight: 36,
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
                minHeight: 36,
                fontSize: 11,
                fontWeight: 600,
                color: COLORS.text.secondary,
                '&.Mui-selected': {
                  color: COLORS.info.main,
                },
              }}
            />
          </Tabs>
        </Box>

        {/* Panel content */}
        <Box sx={{ flex: 1, overflow: 'hidden', p: 1.5 }}>
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
    </>
  );
});

GameSidebar.displayName = 'GameSidebar';
