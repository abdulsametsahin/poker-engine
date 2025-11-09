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
import WinnerDisplay from './WinnerDisplay';
import GameCompleteDisplay from './GameCompleteDisplay';

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

interface GameCompleteData {
  winner: string;
  winnerName?: string;
  finalChips: number;
  totalPlayers: number;
  message: string;
}

export const GameView: React.FC = () => {
  const { tableId } = useParams<{ tableId: string }>();
  const navigate = useNavigate();
  const { isConnected, lastMessage, send } = useWebSocket();
  const [tableState, setTableState] = useState<TableState | null>(null);
  const [raiseAmount, setRaiseAmount] = useState('');
  const [showWinners, setShowWinners] = useState(false);
  const [gameComplete, setGameComplete] = useState<GameCompleteData | null>(null);

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
          // Game is completely over - show game complete modal
          console.log('Game complete detected!', lastMessage.payload);
          setGameComplete({
            winner: lastMessage.payload.winner,
            winnerName: lastMessage.payload.winnerName,
            finalChips: lastMessage.payload.finalChips,
            totalPlayers: lastMessage.payload.totalPlayers,
            message: lastMessage.payload.message,
          });
          setShowWinners(false); // Hide hand winner modal if showing
          break;

        default:
          console.log('Unknown message type:', lastMessage.type);
      }
    }
  }, [lastMessage, tableId, tableState?.status]);

  const handleAction = (action: string, amount?: number) => {
    send({
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
    <Box sx={{ height: '100vh', display: 'flex', flexDirection: 'column', bgcolor: 'grey.100' }}>
      {/* Game Complete Modal - Takes priority over winner modal */}
      {gameComplete ? (
        <GameCompleteDisplay
          winner={gameComplete.winner}
          winnerName={gameComplete.winnerName}
          finalChips={gameComplete.finalChips}
          totalPlayers={gameComplete.totalPlayers}
          message={gameComplete.message}
          currentUserId={currentUserId}
        />
      ) : showWinners && tableState?.winners ? (
        <WinnerDisplay winners={tableState.winners} onClose={handleCloseWinners} />
      ) : null}

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
          <Stack direction="row" spacing={1} alignItems="center">
            <TextField
              type="number"
              value={raiseAmount}
              onChange={(e) => setRaiseAmount(e.target.value)}
              size="small"
              label={`Raise (min: ${minRaiseAmount})`}
              disabled={!isMyTurn}
              sx={{ width: 150 }}
              inputProps={{ min: minRaiseAmount }}
            />
            <Button
              variant="outlined"
              size="small"
              onClick={() => setRaiseAmount(minRaiseAmount.toString())}
              disabled={!isMyTurn}
            >
              Min
            </Button>
            <Button
              variant="outlined"
              size="small"
              onClick={() => setRaiseAmount((tableState?.pot || 0).toString())}
              disabled={!isMyTurn}
            >
              Pot
            </Button>
            <Button
              variant="contained"
              color="secondary"
              onClick={() => handleAction('raise', Number(raiseAmount))}
              disabled={!isMyTurn || !raiseAmount || Number(raiseAmount) < minRaiseAmount}
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
