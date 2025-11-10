import React, { memo, useState } from 'react';
import { Box, Stack, Typography, TextField, IconButton, InputAdornment } from '@mui/material';
import { Chat, Send } from '@mui/icons-material';
import { COLORS, RADIUS } from '../../constants';

interface ChatMessage {
  id: string;
  userId: string;
  username: string;
  message: string;
  timestamp: Date;
}

interface ChatPanelProps {
  messages: ChatMessage[];
  currentUserId?: string;
  onSendMessage: (message: string) => void;
}

export const ChatPanel: React.FC<ChatPanelProps> = memo(({ messages, currentUserId, onSendMessage }) => {
  const [inputMessage, setInputMessage] = useState('');

  const handleSend = () => {
    if (inputMessage.trim()) {
      onSendMessage(inputMessage.trim());
      setInputMessage('');
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <Box
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
      }}
    >
      {/* Messages list */}
      <Stack
        spacing={1}
        sx={{
          flex: 1,
          p: 1.5,
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
        {messages.length === 0 ? (
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
              No messages yet
            </Typography>
          </Box>
        ) : (
          messages.map((msg) => {
            const isCurrentUser = msg.userId === currentUserId;
            return (
              <Box
                key={msg.id}
                sx={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: isCurrentUser ? 'flex-end' : 'flex-start',
                }}
              >
                <Box
                  sx={{
                    maxWidth: '80%',
                    px: 1.5,
                    py: 1,
                    borderRadius: RADIUS.sm,
                    background: isCurrentUser
                      ? COLORS.info.main
                      : `${COLORS.background.secondary}cc`,
                    border: `1px solid ${isCurrentUser ? COLORS.info.main : COLORS.border.main}`,
                  }}
                >
                  {!isCurrentUser && (
                    <Typography
                      variant="caption"
                      sx={{
                        color: COLORS.accent.main,
                        fontSize: 10,
                        fontWeight: 700,
                        display: 'block',
                        mb: 0.5,
                      }}
                    >
                      {msg.username}
                    </Typography>
                  )}
                  <Typography
                    variant="body2"
                    sx={{
                      color: isCurrentUser ? COLORS.text.primary : COLORS.text.primary,
                      fontSize: 12,
                      wordBreak: 'break-word',
                    }}
                  >
                    {msg.message}
                  </Typography>
                  <Typography
                    variant="caption"
                    sx={{
                      color: isCurrentUser ? COLORS.text.secondary : COLORS.text.disabled,
                      fontSize: 9,
                      display: 'block',
                      mt: 0.25,
                    }}
                  >
                    {msg.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                  </Typography>
                </Box>
              </Box>
            );
          })
        )}
      </Stack>

      {/* Input */}
      <Box
        sx={{
          px: 1.5,
          py: 1,
          borderTop: `1px solid ${COLORS.border.main}`,
        }}
      >
        <TextField
          fullWidth
          size="small"
          placeholder="Type a message..."
          value={inputMessage}
          onChange={(e) => setInputMessage(e.target.value)}
          onKeyPress={handleKeyPress}
          InputProps={{
            endAdornment: (
              <InputAdornment position="end">
                <IconButton
                  size="small"
                  onClick={handleSend}
                  disabled={!inputMessage.trim()}
                  sx={{
                    color: COLORS.info.main,
                    '&:disabled': {
                      color: COLORS.text.disabled,
                    },
                  }}
                >
                  <Send sx={{ fontSize: 18 }} />
                </IconButton>
              </InputAdornment>
            ),
          }}
          sx={{
            '& .MuiOutlinedInput-root': {
              color: COLORS.text.primary,
              background: `${COLORS.background.secondary}80`,
              '& fieldset': {
                borderColor: COLORS.border.main,
              },
              '&:hover fieldset': {
                borderColor: COLORS.info.main,
              },
              '&.Mui-focused fieldset': {
                borderColor: COLORS.info.main,
              },
            },
          }}
        />
      </Box>
    </Box>
  );
});

ChatPanel.displayName = 'ChatPanel';
