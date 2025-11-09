import React, { useEffect, useState, useCallback } from 'react';
import { Box, Stack, Slider, IconButton, Dialog, DialogTitle, DialogContent, DialogActions } from '@mui/material';
import { ArrowBack, ExitToApp } from '@mui/icons-material';
import { useParams, useNavigate } from 'react-router-dom';
import { useWebSocket } from '../contexts/WebSocketContext';
import { useAuth } from '../contexts/AuthContext';
import { useToast } from '../contexts/ToastContext';
import { PokerTable } from '../components/game/PokerTable';
import { WinnerDisplay, GameCompleteDisplay } from '../components/modals';
import { Button } from '../components/common/Button';
import { Badge } from '../components/common/Badge';
import { COLORS, RADIUS, SPACING, GAME } from '../constants';
import { Player, WSMessage } from '../types';

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
  winners?: any[];
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
  const { user } = useAuth();
  const { isConnected, sendMessage, addMessageHandler, removeMessageHandler } = useWebSocket();
  const { showSuccess, showError, showWarning } = useToast();

  const [tableState, setTableState] = useState<TableState | null>(null);
  const [raiseAmount, setRaiseAmount] = useState(0);
  const [showWinners, setShowWinners] = useState(false);
  const [gameComplete, setGameComplete] = useState<GameCompleteData | null>(null);
  const [leaveDialogOpen, setLeaveDialogOpen] = useState(false);

  // Find current user
  const currentUserId = user?.id || tableState?.players?.find(p => p.cards && p.cards.length > 0)?.user_id;
  const currentPlayer = tableState?.players?.find(p => p.user_id === currentUserId);
  const isMyTurn = tableState?.current_turn === currentUserId;

  // Calculate bet limits
  const currentBet = tableState?.current_bet || 0;
  const playerBet = currentPlayer?.current_bet || 0;
  const callAmount = currentBet - playerBet;
  const minRaiseAmount = currentBet > 0 ? currentBet * GAME.MIN_RAISE_MULTIPLIER : GAME.BLINDS.heads_up.big * 2;
  const maxRaiseAmount = currentPlayer?.chips || 1000;

  // Subscribe to table on mount
  useEffect(() => {
    if (isConnected && tableId) {
      sendMessage({
        type: 'subscribe_table',
        payload: { table_id: tableId },
      });
    }
  }, [isConnected, tableId, sendMessage]);

  // Handle WebSocket messages
  useEffect(() => {
    const handleTableState = (message: WSMessage) => {
      const newState = {
        table_id: message.payload.table_id || tableId,
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

      // Check if transitioning from handComplete to a new hand
      if (
        (tableState?.status === 'handComplete' && newState.status === 'waiting') ||
        (tableState?.status === 'handComplete' && newState.status === 'playing' && !newState.winners)
      ) {
        setShowWinners(false);
      }

      setTableState(newState);

      // Show winners modal when hand is complete
      if (newState.status === 'handComplete' && newState.winners && newState.winners.length > 0) {
        setShowWinners(true);
      }
    };

    const handleGameUpdate = (message: WSMessage) => {
      handleTableState(message);
    };

    const handleGameComplete = (message: WSMessage) => {
      setGameComplete({
        winner: message.payload.winner,
        winnerName: message.payload.winnerName,
        finalChips: message.payload.finalChips,
        totalPlayers: message.payload.totalPlayers,
        message: message.payload.message,
      });
      setShowWinners(false);
      showSuccess('Game complete!');
    };

    const handleError = (message: WSMessage) => {
      showError(message.payload.message || 'An error occurred');
    };

    addMessageHandler('table_state', handleTableState);
    addMessageHandler('game_update', handleGameUpdate);
    addMessageHandler('game_complete', handleGameComplete);
    addMessageHandler('error', handleError);

    return () => {
      removeMessageHandler('table_state');
      removeMessageHandler('game_update');
      removeMessageHandler('game_complete');
      removeMessageHandler('error');
    };
  }, [tableId, tableState?.status, addMessageHandler, removeMessageHandler, showSuccess, showError]);

  const handleAction = useCallback((action: string, amount?: number) => {
    sendMessage({
      type: 'game_action',
      payload: { action, amount: amount || 0 },
    });
  }, [sendMessage]);

  const handleLeaveGame = () => {
    navigate('/lobby');
  };

  return (
    <Box
      sx={{
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        background: `linear-gradient(135deg, ${COLORS.background.primary} 0%, ${COLORS.background.secondary} 100%)`,
      }}
    >
      {/* Header */}
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
        <Stack direction="row" spacing={2} alignItems="center">
          <IconButton
            onClick={() => setLeaveDialogOpen(true)}
            sx={{
              color: COLORS.text.secondary,
              '&:hover': { color: COLORS.text.primary },
            }}
          >
            <ArrowBack />
          </IconButton>

          <Badge variant={isConnected ? 'success' : 'danger'} pulse={!isConnected}>
            {isConnected ? 'CONNECTED' : 'DISCONNECTED'}
          </Badge>

          {tableState?.status && (
            <Badge
              variant={tableState.status === 'playing' ? 'primary' : 'secondary'}
              pulse={tableState.status === 'playing'}
            >
              {tableState.status.toUpperCase()}
            </Badge>
          )}
        </Stack>

        <IconButton
          onClick={() => setLeaveDialogOpen(true)}
          sx={{
            color: COLORS.danger.main,
            '&:hover': {
              color: COLORS.danger.light,
              background: `${COLORS.danger.main}20`,
            },
          }}
        >
          <ExitToApp />
        </IconButton>
      </Box>

      {/* Main game area */}
      <Box sx={{ flex: 1, p: 2, overflow: 'hidden' }}>
        <PokerTable tableState={tableState} currentUserId={currentUserId} />
      </Box>

      {/* Action bar */}
      <Box
        sx={{
          px: 3,
          py: 2,
          background: 'rgba(0, 0, 0, 0.4)',
          backdropFilter: 'blur(10px)',
          borderTop: `1px solid ${COLORS.border.main}`,
        }}
      >
        <Stack spacing={2}>
          {/* Raise slider */}
          {isMyTurn && (
            <Box sx={{ px: 2 }}>
              <Slider
                value={raiseAmount}
                onChange={(_, value) => setRaiseAmount(value as number)}
                min={minRaiseAmount}
                max={maxRaiseAmount}
                step={10}
                marks={[
                  { value: minRaiseAmount, label: `Min: $${minRaiseAmount}` },
                  { value: maxRaiseAmount / 2, label: `Pot: $${tableState?.pot || 0}` },
                  { value: maxRaiseAmount, label: `All-in: $${maxRaiseAmount}` },
                ]}
                sx={{
                  color: COLORS.primary.main,
                  '& .MuiSlider-thumb': {
                    background: `linear-gradient(135deg, ${COLORS.primary.main} 0%, ${COLORS.secondary.main} 100%)`,
                    boxShadow: `0 0 12px ${COLORS.primary.glow}`,
                  },
                  '& .MuiSlider-track': {
                    background: `linear-gradient(90deg, ${COLORS.primary.main} 0%, ${COLORS.secondary.main} 100%)`,
                  },
                  '& .MuiSlider-markLabel': {
                    color: COLORS.text.secondary,
                    fontSize: '10px',
                  },
                }}
              />
            </Box>
          )}

          {/* Action buttons */}
          <Stack direction="row" spacing={1.5} justifyContent="center" flexWrap="wrap">
            <Button
              variant="danger"
              onClick={() => handleAction('fold')}
              disabled={!isMyTurn}
              sx={{ minWidth: 100 }}
            >
              FOLD
            </Button>

            <Button
              variant="secondary"
              onClick={() => handleAction('check')}
              disabled={!isMyTurn || currentBet > playerBet}
              sx={{ minWidth: 100 }}
            >
              CHECK
            </Button>

            <Button
              variant="success"
              onClick={() => handleAction('call')}
              disabled={!isMyTurn || callAmount <= 0}
              sx={{ minWidth: 100 }}
            >
              CALL {callAmount > 0 && `$${callAmount}`}
            </Button>

            <Button
              variant="primary"
              onClick={() => handleAction('raise', raiseAmount)}
              disabled={!isMyTurn || raiseAmount < minRaiseAmount}
              sx={{ minWidth: 120 }}
            >
              RAISE ${raiseAmount}
            </Button>

            <Button
              variant="warning"
              onClick={() => handleAction('allin')}
              disabled={!isMyTurn}
              sx={{
                minWidth: 100,
                background: `linear-gradient(135deg, ${COLORS.warning.main} 0%, ${COLORS.warning.dark} 100%)`,
                '&:hover': {
                  background: `linear-gradient(135deg, ${COLORS.warning.light} 0%, ${COLORS.warning.main} 100%)`,
                },
              }}
            >
              ALL-IN
            </Button>
          </Stack>
        </Stack>
      </Box>

      {/* Winner Display Modal */}
      {showWinners && tableState?.winners && tableState.winners.length > 0 && (
        <WinnerDisplay
          winners={tableState.winners}
          onClose={() => setShowWinners(false)}
        />
      )}

      {/* Game Complete Modal */}
      {gameComplete && (
        <GameCompleteDisplay
          winner={gameComplete.winner}
          winnerName={gameComplete.winnerName}
          finalChips={gameComplete.finalChips}
          totalPlayers={gameComplete.totalPlayers}
          message={gameComplete.message}
          currentUserId={currentUserId}
        />
      )}

      {/* Leave game confirmation dialog */}
      <Dialog
        open={leaveDialogOpen}
        onClose={() => setLeaveDialogOpen(false)}
        PaperProps={{
          sx: {
            background: COLORS.background.paper,
            borderRadius: RADIUS.md,
            border: `1px solid ${COLORS.border.main}`,
          },
        }}
      >
        <DialogTitle sx={{ color: COLORS.text.primary }}>Leave Game?</DialogTitle>
        <DialogContent>
          <Box sx={{ color: COLORS.text.secondary }}>
            Are you sure you want to leave this game? You will forfeit your chips if the game is in progress.
          </Box>
        </DialogContent>
        <DialogActions>
          <Button variant="ghost" onClick={() => setLeaveDialogOpen(false)}>
            Cancel
          </Button>
          <Button variant="danger" onClick={handleLeaveGame}>
            Leave Game
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};
