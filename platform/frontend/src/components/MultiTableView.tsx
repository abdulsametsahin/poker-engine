import React, { useState, useEffect, useCallback } from 'react';
import { Box, Stack, IconButton, Typography, Tabs, Tab } from '@mui/material';
import { Add, Close, Fullscreen, FullscreenExit } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useWebSocket } from '../contexts/WebSocketContext';
import { PokerTable } from '../components/game/PokerTable';
import { COLORS, RADIUS, SPACING } from '../constants';
import { Player, WSMessage } from '../types';

interface TableState {
  table_id: string;
  players: Player[];
  community_cards?: string[];
  pot?: number;
  current_turn?: string;
  status?: string;
  betting_round?: string;
  current_bet?: number;
  action_deadline?: string;
  winners?: any[];
}

interface MultiTableViewProps {
  currentUserId?: string;
  onTableAction?: (tableId: string, action: string, amount?: number) => void;
}

export const MultiTableView: React.FC<MultiTableViewProps> = ({ currentUserId, onTableAction }) => {
  const navigate = useNavigate();
  const { isConnected, addMessageHandler, removeMessageHandler } = useWebSocket();
  
  const [tables, setTables] = useState<Map<string, TableState>>(new Map());
  const [activeTableId, setActiveTableId] = useState<string | null>(null);
  const [fullscreenTableId, setFullscreenTableId] = useState<string | null>(null);

  // Handle WebSocket messages for all subscribed tables
  useEffect(() => {
    const handleTableState = (message: WSMessage<any>) => {
      const tableId = message.payload.table_id;
      if (!tableId) return;

      const newState: TableState = {
        table_id: tableId,
        players: message.payload.players || [],
        community_cards: message.payload.community_cards || [],
        pot: message.payload.pot || 0,
        current_turn: message.payload.current_turn,
        status: message.payload.status,
        betting_round: message.payload.betting_round,
        current_bet: message.payload.current_bet,
        action_deadline: message.payload.action_deadline,
        winners: message.payload.winners,
      };

      setTables(prev => new Map(prev).set(tableId, newState));
    };

    const handleGameUpdate = (message: WSMessage<any>) => {
      handleTableState(message);
    };

    const cleanup1 = addMessageHandler('table_state', handleTableState);
    const cleanup2 = addMessageHandler('game_update', handleGameUpdate);

    return () => {
      cleanup1();
      cleanup2();
    };
  }, [addMessageHandler]);

  const handleAddTable = useCallback(() => {
    // Navigate to lobby to join a new table
    navigate('/lobby');
  }, [navigate]);

  const handleCloseTable = useCallback((tableId: string) => {
    setTables(prev => {
      const newTables = new Map(prev);
      newTables.delete(tableId);
      return newTables;
    });
    
    // If closing the active table, switch to another
    if (activeTableId === tableId) {
      const remainingTables = Array.from(tables.keys()).filter(id => id !== tableId);
      setActiveTableId(remainingTables.length > 0 ? remainingTables[0] : null);
    }
  }, [activeTableId, tables]);

  const handleToggleFullscreen = useCallback((tableId: string) => {
    setFullscreenTableId(prev => prev === tableId ? null : tableId);
  }, []);

  const handleTabChange = useCallback((_: React.SyntheticEvent, newValue: string) => {
    setActiveTableId(newValue);
  }, []);

  const handleOpenSingleTable = useCallback((tableId: string) => {
    navigate(`/game/${tableId}`);
  }, [navigate]);

  const tablesArray = Array.from(tables.entries());

  // If no tables, show empty state
  if (tablesArray.length === 0) {
    return (
      <Box
        sx={{
          height: '100vh',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          background: `linear-gradient(135deg, ${COLORS.background.primary} 0%, ${COLORS.background.secondary} 100%)`,
        }}
      >
        <Typography variant="h5" sx={{ color: COLORS.text.secondary, mb: 2 }}>
          No active tables
        </Typography>
        <IconButton
          onClick={handleAddTable}
          sx={{
            width: 60,
            height: 60,
            bgcolor: COLORS.primary.main,
            color: COLORS.text.primary,
            '&:hover': {
              bgcolor: COLORS.primary.light,
            },
          }}
        >
          <Add sx={{ fontSize: 32 }} />
        </IconButton>
      </Box>
    );
  }

  // Fullscreen mode - show only one table
  if (fullscreenTableId) {
    const tableState = tables.get(fullscreenTableId);
    if (!tableState) return null;

    return (
      <Box
        sx={{
          height: '100vh',
          display: 'flex',
          flexDirection: 'column',
          background: `linear-gradient(135deg, ${COLORS.background.primary} 0%, ${COLORS.background.secondary} 100%)`,
        }}
      >
        {/* Fullscreen header */}
        <Box
          sx={{
            px: 3,
            py: 1.5,
            background: 'rgba(0, 0, 0, 0.4)',
            backdropFilter: 'blur(10px)',
            borderBottom: `1px solid ${COLORS.border.main}`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <Typography variant="h6" sx={{ color: COLORS.text.primary }}>
            Table {fullscreenTableId}
          </Typography>
          <IconButton
            onClick={() => setFullscreenTableId(null)}
            sx={{
              color: COLORS.text.secondary,
              '&:hover': { color: COLORS.text.primary },
            }}
          >
            <FullscreenExit />
          </IconButton>
        </Box>

        {/* Table view */}
        <Box sx={{ flex: 1, p: 3, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <PokerTable tableState={tableState} currentUserId={currentUserId} />
        </Box>
      </Box>
    );
  }

  // Multi-table grid view
  return (
    <Box
      sx={{
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        background: `linear-gradient(135deg, ${COLORS.background.primary} 0%, ${COLORS.background.secondary} 100%)`,
      }}
    >
      {/* Header with tabs */}
      <Box
        sx={{
          px: 3,
          py: 1.5,
          background: 'rgba(0, 0, 0, 0.4)',
          backdropFilter: 'blur(10px)',
          borderBottom: `1px solid ${COLORS.border.main}`,
        }}
      >
        <Stack direction="row" spacing={2} alignItems="center">
          <Tabs
            value={activeTableId || false}
            onChange={handleTabChange}
            sx={{
              flex: 1,
              minHeight: 40,
              '& .MuiTab-root': {
                color: COLORS.text.secondary,
                minHeight: 40,
                '&.Mui-selected': {
                  color: COLORS.primary.main,
                },
              },
              '& .MuiTabs-indicator': {
                backgroundColor: COLORS.primary.main,
              },
            }}
          >
            {tablesArray.map(([tableId]) => (
              <Tab
                key={tableId}
                label={`Table ${tableId.substring(0, 8)}`}
                value={tableId}
              />
            ))}
          </Tabs>

          <IconButton
            onClick={handleAddTable}
            sx={{
              color: COLORS.primary.main,
              '&:hover': { color: COLORS.primary.light },
            }}
          >
            <Add />
          </IconButton>
        </Stack>
      </Box>

      {/* Table grid */}
      <Box
        sx={{
          flex: 1,
          p: 2,
          display: 'grid',
          gridTemplateColumns: tablesArray.length === 1 ? '1fr' : tablesArray.length === 2 ? 'repeat(2, 1fr)' : 'repeat(auto-fit, minmax(500px, 1fr))',
          gap: 2,
          overflow: 'auto',
        }}
      >
        {tablesArray.map(([tableId, tableState]) => (
          <Box
            key={tableId}
            sx={{
              position: 'relative',
              borderRadius: RADIUS.md,
              border: `2px solid ${activeTableId === tableId ? COLORS.primary.main : COLORS.border.main}`,
              background: 'rgba(0, 0, 0, 0.3)',
              overflow: 'hidden',
              minHeight: 300,
              display: 'flex',
              flexDirection: 'column',
              transition: 'all 0.2s ease',
              '&:hover': {
                borderColor: COLORS.primary.light,
              },
            }}
          >
            {/* Table header */}
            <Box
              sx={{
                px: 2,
                py: 1,
                background: 'rgba(0, 0, 0, 0.5)',
                borderBottom: `1px solid ${COLORS.border.main}`,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
              }}
            >
              <Typography variant="body2" sx={{ color: COLORS.text.primary, fontWeight: 600 }}>
                Table {tableId.substring(0, 8)}
              </Typography>
              <Stack direction="row" spacing={0.5}>
                <IconButton
                  size="small"
                  onClick={() => handleOpenSingleTable(tableId)}
                  sx={{
                    color: COLORS.text.secondary,
                    '&:hover': { color: COLORS.text.primary },
                  }}
                >
                  <Fullscreen fontSize="small" />
                </IconButton>
                <IconButton
                  size="small"
                  onClick={() => handleCloseTable(tableId)}
                  sx={{
                    color: COLORS.danger.main,
                    '&:hover': { color: COLORS.danger.light },
                  }}
                >
                  <Close fontSize="small" />
                </IconButton>
              </Stack>
            </Box>

            {/* Table content */}
            <Box
              sx={{
                flex: 1,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                p: 2,
              }}
              onClick={() => setActiveTableId(tableId)}
            >
              <Box
                sx={{
                  width: '100%',
                  height: '100%',
                  transform: 'scale(0.8)',
                  transformOrigin: 'center',
                }}
              >
                <PokerTable tableState={tableState} currentUserId={currentUserId} />
              </Box>
            </Box>
          </Box>
        ))}
      </Box>
    </Box>
  );
};

export default MultiTableView;
