import React, { memo, useState } from 'react';
import { Box, Tabs, Tab, IconButton, Badge as MuiBadge, Tooltip } from '@mui/material';
import { History, Chat, Close, ChevronLeft } from '@mui/icons-material';
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
  const [activeTab, setActiveTab] = useState(1); // Default to Chat
  const [isExpanded, setIsExpanded] = useState(false);

  // Count unread messages (simple heuristic - in production, track properly)
  const unreadCount = 0; // You can implement proper unread tracking

  return (
    <>
      {/* Floating toggle buttons (when collapsed) */}
      {!isExpanded && (
        <Box
          sx={{
            position: 'fixed',
            top: 16,
            right: 16,
            display: 'flex',
            flexDirection: 'column',
            gap: 1,
            zIndex: 100,
          }}
        >
          <Tooltip title="Chat" placement="left">
            <IconButton
              onClick={() => {
                setActiveTab(1);
                setIsExpanded(true);
              }}
              sx={{
                width: 48,
                height: 48,
                background: 'rgba(0, 0, 0, 0.7)',
                backdropFilter: 'blur(10px)',
                border: `1px solid ${COLORS.info.main}`,
                boxShadow: `0 4px 12px rgba(0, 0, 0, 0.5), 0 0 12px ${COLORS.info.glow}`,
                transition: TRANSITIONS.normal,
                '&:hover': {
                  background: `linear-gradient(135deg, ${COLORS.info.main}40 0%, ${COLORS.info.dark}40 100%)`,
                  transform: 'scale(1.05)',
                  boxShadow: `0 6px 16px rgba(0, 0, 0, 0.6), 0 0 16px ${COLORS.info.glow}`,
                },
              }}
            >
              <MuiBadge badgeContent={unreadCount} color="error">
                <Chat sx={{ color: COLORS.info.main, fontSize: 20 }} />
              </MuiBadge>
            </IconButton>
          </Tooltip>

          <Tooltip title="History" placement="left">
            <IconButton
              onClick={() => {
                setActiveTab(0);
                setIsExpanded(true);
              }}
              sx={{
                width: 48,
                height: 48,
                background: 'rgba(0, 0, 0, 0.7)',
                backdropFilter: 'blur(10px)',
                border: `1px solid ${COLORS.secondary.main}`,
                boxShadow: `0 4px 12px rgba(0, 0, 0, 0.5), 0 0 12px ${COLORS.secondary.glow}`,
                transition: TRANSITIONS.normal,
                '&:hover': {
                  background: `linear-gradient(135deg, ${COLORS.secondary.main}40 0%, ${COLORS.secondary.dark}40 100%)`,
                  transform: 'scale(1.05)',
                  boxShadow: `0 6px 16px rgba(0, 0, 0, 0.6), 0 0 16px ${COLORS.secondary.glow}`,
                },
              }}
            >
              <History sx={{ color: COLORS.secondary.main, fontSize: 20 }} />
            </IconButton>
          </Tooltip>
        </Box>
      )}

      {/* Expanded panel */}
      {isExpanded && (
        <>
          {/* Backdrop */}
          <Box
            onClick={() => setIsExpanded(false)}
            sx={{
              position: 'fixed',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              background: 'rgba(0, 0, 0, 0.5)',
              backdropFilter: 'blur(4px)',
              zIndex: 99,
              animation: 'fadeIn 0.2s ease-out',
              '@keyframes fadeIn': {
                from: { opacity: 0 },
                to: { opacity: 1 },
              },
            }}
          />

          {/* Floating panel */}
          <Box
            onClick={(e) => e.stopPropagation()}
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
              animation: 'slideIn 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
              '@keyframes slideIn': {
                from: {
                  transform: 'translateX(100%)',
                  opacity: 0,
                },
                to: {
                  transform: 'translateX(0)',
                  opacity: 1,
                },
              },
            }}
          >
            {/* Header with close button */}
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

              <IconButton
                onClick={() => setIsExpanded(false)}
                size="small"
                sx={{
                  color: COLORS.text.secondary,
                  '&:hover': {
                    color: COLORS.text.primary,
                    background: 'rgba(255, 255, 255, 0.1)',
                  },
                }}
              >
                <ChevronLeft fontSize="small" />
              </IconButton>
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
      )}
    </>
  );
});

GameSidebar.displayName = 'GameSidebar';
