import React, { useEffect, useState } from 'react';
import {
  Box,
  Button,
  Stack,
  TextField,
  Chip,
} from '@mui/material';
import { useParams, useNavigate } from 'react-router-dom';
import { useWebSocket } from '../hooks/useWebSocket';
import PokerTable from './PokerTable';

interface Player {
  user_id: string;
  seat_number: number;
  chips: number;
  status: string;
  bet?: number;
  cards?: string[];
}

interface TableState {
  table_id?: string;
  players: Player[];
  community_cards?: string[];
  pot?: number;
  current_turn?: string;
  status?: string;
  betting_round?: string;
  current_bet?: number;
  action_deadline?: string;
}

export const GameView: React.FC = () => {
  const { tableId } = useParams<{ tableId: string }>();
  const navigate = useNavigate();
  const { isConnected, lastMessage, send } = useWebSocket();
  const [tableState, setTableState] = useState<TableState | null>(null);
  const [raiseAmount, setRaiseAmount] = useState(0);

  useEffect(() => {
    if (isConnected && tableId) {
      send({
        type: 'subscribe_table',
        payload: { table_id: tableId },
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isConnected, tableId]);

  useEffect(() => {
    if (lastMessage) {
      switch (lastMessage.type) {
        case 'table_state':
        case 'game_update':
          setTableState({
            table_id: lastMessage.payload.table_id || tableId,
            players: lastMessage.payload.players || [],
            community_cards: lastMessage.payload.community_cards || [],
            pot: lastMessage.payload.pot || 0,
            current_turn: lastMessage.payload.current_turn,
            status: lastMessage.payload.status,
            betting_round: lastMessage.payload.betting_round,
            current_bet: lastMessage.payload.current_bet,
            action_deadline: lastMessage.payload.action_deadline,
          });
          break;
      }
    }
  }, [lastMessage, tableId]);

  const handleAction = (action: string, amount?: number) => {
    send({
      type: 'game_action',
      payload: { action, amount: amount || 0 },
    });
  };

  // Find current user ID by finding the player with cards (backend only sends cards to the player)
  const currentUserId = tableState?.players?.find(p => p.cards && p.cards.length > 0)?.user_id;

  // Check if it's the current user's turn
  const isMyTurn = tableState?.current_turn === currentUserId;

  return (
    <Box sx={{ height: '100vh', display: 'flex', flexDirection: 'column', bgcolor: 'grey.100' }}>
      {/* Main Game Area */}
      <Box sx={{ flex: 1, overflow: 'auto', p: 2 }}>
        <PokerTable tableState={tableState} />
      </Box>

      {/* Action Buttons */}
      <Box
        sx={{
          bgcolor: 'background.paper',
          borderTop: '2px solid',
          borderColor: 'divider',
          p: 2,
        }}
      >
        <Stack direction="row" spacing={2} justifyContent="space-between" alignItems="center">
          {/* Connection Status & Back Button */}
          <Stack direction="row" spacing={1}>
            <Button
              variant="outlined"
              onClick={() => navigate('/lobby')}
              size="small"
            >
              ← Back to Lobby
            </Button>
            <Chip
              label={isConnected ? 'Connected' : 'Disconnected'}
              color={isConnected ? 'success' : 'error'}
              size="small"
            />
          </Stack>

          {/* Game Actions */}
          <Stack direction="row" spacing={1} sx={{ flex: 1, maxWidth: 600, mx: 'auto' }}>
            <Button
              variant="contained"
              color="error"
              onClick={() => handleAction('fold')}
              disabled={!isMyTurn}
              sx={{ flex: 1 }}
            >
              Fold
            </Button>
            <Button
              variant="outlined"
              onClick={() => handleAction('check')}
              disabled={!isMyTurn}
              sx={{ flex: 1 }}
            >
              Check
            </Button>
            <Button
              variant="contained"
              color="primary"
              onClick={() => handleAction('call')}
              disabled={!isMyTurn}
              sx={{ flex: 1 }}
            >
              Call
            </Button>
          </Stack>

          {/* Raise Controls */}
          <Stack direction="row" spacing={1}>
            <TextField
              type="number"
              value={raiseAmount}
              onChange={(e) => setRaiseAmount(Number(e.target.value))}
              size="small"
              label="Amount"
              disabled={!isMyTurn}
              sx={{ width: 120 }}
            />
            <Button
              variant="contained"
              color="secondary"
              onClick={() => handleAction('raise', raiseAmount)}
              disabled={!isMyTurn}
            >
              Raise
            </Button>
            <Button
              variant="contained"
              color="warning"
              onClick={() => handleAction('allin')}
              disabled={!isMyTurn}
            >
              All-In
            </Button>
          </Stack>
        </Stack>
      </Box>
    </Box>
  );
};
