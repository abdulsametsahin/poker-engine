import React, { useEffect, useState } from 'react';
import {
  Box,
  Button,
  Stack,
  TextField,
  Chip,
  Typography,
} from '@mui/material';
import { useParams, useNavigate } from 'react-router-dom';
import { useWebSocket } from '../contexts/WebSocketContext';
import PokerTable from './PokerTable';
import WinnerDisplay from './WinnerDisplay';

interface Player {
  user_id: string;
  seat_number: number;
  chips: number;
  status: string;
  bet?: number;
  cards?: string[];
}

interface CardObject {
  rank: string;
  suit: string;
}

interface Winner {
  playerId: string;
  amount: number;
  handRank: string;
  handCards: (string | CardObject)[];
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
  winners?: Winner[];
}

export const GameView: React.FC = () => {
  const { tableId } = useParams<{ tableId: string }>();
  const navigate = useNavigate();
  const { isConnected, lastMessage, sendMessage } = useWebSocket();
  const [tableState, setTableState] = useState<TableState | null>(null);
  const [raiseAmount, setRaiseAmount] = useState('');
  const [showWinners, setShowWinners] = useState(false);

  useEffect(() => {
    if (isConnected && tableId) {
      sendMessage({
        type: 'subscribe_table',
        payload: { table_id: tableId },
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isConnected, tableId]);

  useEffect(() => {
    if (lastMessage) {
      console.log('Received message:', lastMessage.type, lastMessage.payload);

      switch (lastMessage.type) {
        case 'table_state':
        case 'game_update':
          const newState = {
            table_id: lastMessage.payload.table_id || tableId,
            players: lastMessage.payload.players || [],
            community_cards: lastMessage.payload.community_cards || [],
            pot: lastMessage.payload.pot || 0,
            current_turn: lastMessage.payload.current_turn,
            status: lastMessage.payload.status,
            betting_round: lastMessage.payload.betting_round,
            current_bet: lastMessage.payload.current_bet,
            action_deadline: lastMessage.payload.action_deadline,
            winners: lastMessage.payload.winners,
          };

          // Check if game ended or new hand started (went back to waiting or playing without winners)
          if (
            (tableState?.status === 'handComplete' && newState.status === 'waiting') ||
            (tableState?.status === 'handComplete' && newState.status === 'playing' && !newState.winners)
          ) {
            // New hand started or game returned to waiting, hide winners modal
            setShowWinners(false);
          }

          setTableState(newState);

          // Show winners modal when hand is complete and we have winners
          if (newState.status === 'handComplete' && newState.winners && newState.winners.length > 0) {
            setShowWinners(true);
          }
          break;

        case 'game_complete':
          // Game is completely over - navigate to lobby after showing success message
          console.log('Game complete detected!', lastMessage.payload);
          setTimeout(() => {
            navigate('/lobby');
          }, 3000);
          break;

        default:
          console.log('Unknown message type:', lastMessage.type);
      }
    }
  }, [lastMessage, tableId, tableState?.status, navigate]);

  const handleAction = (action: string, amount?: number) => {
    sendMessage({
      type: 'game_action',
      payload: { action, amount: amount || 0 },
    });
  };

  // Find current user ID by finding the player with cards (backend only sends cards to the player)
  const currentUserId = tableState?.players?.find(p => p.cards && p.cards.length > 0)?.user_id;
  const currentPlayer = tableState?.players?.find(p => p.user_id === currentUserId);

  // Check if it's the current user's turn
  const isMyTurn = tableState?.current_turn === currentUserId;

  // Calculate minimum raise amount
  // Minimum raise = current bet + min raise amount (which equals the last raise)
  // For simplicity, we'll use current bet + big blind as minimum
  const currentBet = tableState?.current_bet || 0;
  const minRaiseAmount = currentBet * 2 || 20; // At minimum, double the current bet

  // Get player's current bet
  const playerBet = currentPlayer?.bet || 0;

  const handleCloseWinners = () => {
    setShowWinners(false);
  };

  return (
    <Box sx={{
      height: '100vh',
      display: 'flex',
      flexDirection: 'column',
      bgcolor: '#0f1419',
      backgroundImage: 'radial-gradient(circle at 20% 30%, rgba(16, 185, 129, 0.05) 0%, transparent 50%), radial-gradient(circle at 80% 70%, rgba(99, 102, 241, 0.05) 0%, transparent 50%)',
    }}>
      {/* Winner Modal */}
      {showWinners && tableState?.winners ? (
        <WinnerDisplay winners={tableState.winners} onClose={handleCloseWinners} />
      ) : null}

      {/* Compact Header */}
      <Box sx={{
        bgcolor: 'rgba(17, 24, 39, 0.95)',
        borderBottom: '1px solid rgba(255, 255, 255, 0.05)',
        backdropFilter: 'blur(10px)',
        px: 2,
        py: 1,
      }}>
        <Stack direction="row" justifyContent="space-between" alignItems="center">
          <Stack direction="row" spacing={1.5} alignItems="center">
            <Button
              onClick={() => navigate('/lobby')}
              size="small"
              sx={{
                color: 'rgba(255, 255, 255, 0.7)',
                fontSize: '12px',
                minWidth: 'auto',
                px: 1.5,
                py: 0.5,
                '&:hover': { color: '#fff', bgcolor: 'rgba(255, 255, 255, 0.05)' }
              }}
            >
              ← BACK TO LOBBY
            </Button>
            <Box sx={{
              px: 1.5,
              py: 0.5,
              borderRadius: 1,
              bgcolor: isConnected ? 'rgba(16, 185, 129, 0.15)' : 'rgba(239, 68, 68, 0.15)',
              border: `1px solid ${isConnected ? 'rgba(16, 185, 129, 0.3)' : 'rgba(239, 68, 68, 0.3)'}`,
            }}>
              <Typography variant="caption" sx={{ color: isConnected ? '#10b981' : '#ef4444', fontSize: '11px', fontWeight: 600 }}>
                {isConnected ? 'Connected' : 'Disconnected'}
              </Typography>
            </Box>
          </Stack>

          <Typography variant="caption" sx={{ color: 'rgba(255, 255, 255, 0.4)', fontSize: '11px' }}>
            {tableState?.table_id?.slice(0, 16)}...
          </Typography>

          <Chip
            label={tableState?.status || 'waiting'}
            size="small"
            sx={{
              bgcolor: tableState?.status === 'playing' ? 'rgba(16, 185, 129, 0.15)' : 'rgba(156, 163, 175, 0.15)',
              color: tableState?.status === 'playing' ? '#10b981' : '#9ca3af',
              border: `1px solid ${tableState?.status === 'playing' ? 'rgba(16, 185, 129, 0.3)' : 'rgba(156, 163, 175, 0.3)'}`,
              fontSize: '11px',
              height: '22px',
              fontWeight: 600,
            }}
          />
        </Stack>
      </Box>

      {/* Main Game Area */}
      <Box sx={{ flex: 1, overflow: 'hidden', p: 1.5 }}>
        <PokerTable tableState={tableState} />
      </Box>

      {/* Compact Action Buttons */}
      <Box
        sx={{
          bgcolor: 'rgba(17, 24, 39, 0.95)',
          borderTop: '1px solid rgba(255, 255, 255, 0.05)',
          backdropFilter: 'blur(10px)',
          px: 2,
          py: 1.5,
        }}
      >
        <Stack direction="row" spacing={1} justifyContent="center" alignItems="center" sx={{ maxWidth: 1200, mx: 'auto' }}>
          {/* Primary Actions */}
          <Button
            variant="contained"
            onClick={() => handleAction('fold')}
            disabled={!isMyTurn}
            sx={{
              minWidth: 90,
              height: 42,
              bgcolor: 'rgba(239, 68, 68, 0.15)',
              color: '#ef4444',
              border: '1px solid rgba(239, 68, 68, 0.3)',
              fontWeight: 700,
              fontSize: '13px',
              '&:hover': { bgcolor: 'rgba(239, 68, 68, 0.25)', borderColor: 'rgba(239, 68, 68, 0.5)' },
              '&:disabled': { bgcolor: 'rgba(75, 85, 99, 0.1)', color: 'rgba(156, 163, 175, 0.3)', border: '1px solid rgba(75, 85, 99, 0.2)' }
            }}
          >
            FOLD
          </Button>

          <Button
            variant="outlined"
            onClick={() => handleAction('check')}
            disabled={!isMyTurn}
            sx={{
              minWidth: 90,
              height: 42,
              bgcolor: 'rgba(59, 130, 246, 0.05)',
              color: '#60a5fa',
              border: '1px solid rgba(59, 130, 246, 0.3)',
              fontWeight: 700,
              fontSize: '13px',
              '&:hover': { bgcolor: 'rgba(59, 130, 246, 0.15)', borderColor: 'rgba(59, 130, 246, 0.5)' },
              '&:disabled': { bgcolor: 'rgba(75, 85, 99, 0.1)', color: 'rgba(156, 163, 175, 0.3)', border: '1px solid rgba(75, 85, 99, 0.2)' }
            }}
          >
            CHECK
          </Button>

          <Button
            variant="contained"
            onClick={() => handleAction('call')}
            disabled={!isMyTurn}
            sx={{
              minWidth: 90,
              height: 42,
              bgcolor: 'rgba(16, 185, 129, 0.15)',
              color: '#10b981',
              border: '1px solid rgba(16, 185, 129, 0.3)',
              fontWeight: 700,
              fontSize: '13px',
              '&:hover': { bgcolor: 'rgba(16, 185, 129, 0.25)', borderColor: 'rgba(16, 185, 129, 0.5)' },
              '&:disabled': { bgcolor: 'rgba(75, 85, 99, 0.1)', color: 'rgba(156, 163, 175, 0.3)', border: '1px solid rgba(75, 85, 99, 0.2)' }
            }}
          >
            CALL
          </Button>

          <Box sx={{ width: 1, height: 32, bgcolor: 'rgba(255, 255, 255, 0.05)', mx: 0.5 }} />

          {/* Raise Amount Input */}
          <TextField
            type="number"
            value={raiseAmount}
            onChange={(e) => setRaiseAmount(e.target.value)}
            size="small"
            placeholder={`Min: ${minRaiseAmount}`}
            disabled={!isMyTurn}
            sx={{
              width: 120,
              '& .MuiOutlinedInput-root': {
                height: 42,
                bgcolor: 'rgba(31, 41, 55, 0.5)',
                color: '#fff',
                fontSize: '13px',
                fontWeight: 600,
                '& fieldset': { borderColor: 'rgba(75, 85, 99, 0.3)' },
                '&:hover fieldset': { borderColor: 'rgba(99, 102, 241, 0.5)' },
                '&.Mui-focused fieldset': { borderColor: '#6366f1' },
                '&.Mui-disabled': { bgcolor: 'rgba(31, 41, 55, 0.3)' }
              },
              '& input': { textAlign: 'center', padding: '0 8px' }
            }}
            inputProps={{ min: minRaiseAmount }}
          />

          <Button
            size="small"
            onClick={() => setRaiseAmount(minRaiseAmount.toString())}
            disabled={!isMyTurn}
            sx={{
              minWidth: 50,
              height: 42,
              bgcolor: 'rgba(75, 85, 99, 0.2)',
              color: 'rgba(255, 255, 255, 0.7)',
              border: '1px solid rgba(75, 85, 99, 0.3)',
              fontSize: '11px',
              fontWeight: 700,
              '&:hover': { bgcolor: 'rgba(75, 85, 99, 0.3)' },
              '&:disabled': { color: 'rgba(156, 163, 175, 0.3)' }
            }}
          >
            MIN
          </Button>

          <Button
            size="small"
            onClick={() => setRaiseAmount((tableState?.pot || 0).toString())}
            disabled={!isMyTurn}
            sx={{
              minWidth: 50,
              height: 42,
              bgcolor: 'rgba(75, 85, 99, 0.2)',
              color: 'rgba(255, 255, 255, 0.7)',
              border: '1px solid rgba(75, 85, 99, 0.3)',
              fontSize: '11px',
              fontWeight: 700,
              '&:hover': { bgcolor: 'rgba(75, 85, 99, 0.3)' },
              '&:disabled': { color: 'rgba(156, 163, 175, 0.3)' }
            }}
          >
            POT
          </Button>

          <Button
            variant="contained"
            onClick={() => handleAction('raise', Number(raiseAmount))}
            disabled={!isMyTurn || !raiseAmount || Number(raiseAmount) < minRaiseAmount}
            sx={{
              minWidth: 90,
              height: 42,
              bgcolor: 'rgba(99, 102, 241, 0.15)',
              color: '#6366f1',
              border: '1px solid rgba(99, 102, 241, 0.3)',
              fontWeight: 700,
              fontSize: '13px',
              '&:hover': { bgcolor: 'rgba(99, 102, 241, 0.25)', borderColor: 'rgba(99, 102, 241, 0.5)' },
              '&:disabled': { bgcolor: 'rgba(75, 85, 99, 0.1)', color: 'rgba(156, 163, 175, 0.3)', border: '1px solid rgba(75, 85, 99, 0.2)' }
            }}
          >
            RAISE
          </Button>

          <Button
            variant="contained"
            onClick={() => handleAction('allin')}
            disabled={!isMyTurn}
            sx={{
              minWidth: 90,
              height: 42,
              bgcolor: 'linear-gradient(135deg, rgba(251, 191, 36, 0.15) 0%, rgba(245, 158, 11, 0.15) 100%)',
              color: '#fbbf24',
              border: '1px solid rgba(251, 191, 36, 0.3)',
              fontWeight: 700,
              fontSize: '13px',
              '&:hover': {
                bgcolor: 'linear-gradient(135deg, rgba(251, 191, 36, 0.25) 0%, rgba(245, 158, 11, 0.25) 100%)',
                borderColor: 'rgba(251, 191, 36, 0.5)'
              },
              '&:disabled': { bgcolor: 'rgba(75, 85, 99, 0.1)', color: 'rgba(156, 163, 175, 0.3)', border: '1px solid rgba(75, 85, 99, 0.2)' }
            }}
          >
            ALL-IN
          </Button>
        </Stack>
      </Box>
    </Box>
  );
};
