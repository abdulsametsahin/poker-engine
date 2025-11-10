import React, { useEffect, useState, useCallback } from 'react';
import { Box, Stack, IconButton, Dialog, DialogTitle, DialogContent, DialogActions, TextField, InputAdornment } from '@mui/material';
import { ArrowBack, ExitToApp } from '@mui/icons-material';
import { useParams, useNavigate } from 'react-router-dom';
import { useWebSocket } from '../contexts/WebSocketContext';
import { useAuth } from '../contexts/AuthContext';
import { useToast } from '../contexts/ToastContext';
import { PokerTable } from '../components/game/PokerTable';
import { GameSidebar } from '../components/game/GameSidebar';
import { WinnerDisplay, HandCompleteDisplay } from '../components/modals';
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

export const GameView: React.FC = () => {
  const { tableId } = useParams<{ tableId: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { isConnected, sendMessage, addMessageHandler, removeMessageHandler } = useWebSocket();
  const { showSuccess, showError, showWarning } = useToast();

  const [tableState, setTableState] = useState<TableState | null>(null);
  const [raiseAmount, setRaiseAmount] = useState(0);
  const [showHandComplete, setShowHandComplete] = useState(false);
  const [showGameComplete, setShowGameComplete] = useState(false);
  const [gameMode, setGameMode] = useState<string>('heads_up');
  const [leaveDialogOpen, setLeaveDialogOpen] = useState(false);
  const [history, setHistory] = useState<any[]>([]);
  const [chatMessages, setChatMessages] = useState<any[]>([]);

  // Find current user
  const currentUserId = user?.id || tableState?.players?.find(p => p.cards && p.cards.length > 0)?.user_id;
  const currentPlayer = tableState?.players?.find(p => p.user_id === currentUserId);
  const isMyTurn = tableState?.current_turn === currentUserId;

  // Debug logging
  useEffect(() => {
    if (tableState) {
      console.log('Current User ID:', currentUserId);
      console.log('Current Turn:', tableState.current_turn);
      console.log('Is My Turn:', isMyTurn);
      console.log('Current Player:', currentPlayer);
      console.log('All Players:', tableState.players);
      console.log('Table State:', tableState);
    }
  }, [tableState, currentUserId, isMyTurn, currentPlayer]);

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

      // Track player actions and add to history
      if (tableState && tableState.players) {
        newState.players.forEach((newPlayer: Player) => {
          const oldPlayer = tableState.players.find((p: Player) => p.user_id === newPlayer.user_id);

          // Check if player took a new action (last_action changed)
          if (oldPlayer && newPlayer.last_action && newPlayer.last_action !== oldPlayer.last_action) {
            const actionName = newPlayer.last_action.toLowerCase();
            const playerName = newPlayer.username || 'Player';
            const amount = newPlayer.current_bet && newPlayer.current_bet > (oldPlayer.current_bet || 0)
              ? newPlayer.current_bet - (oldPlayer.current_bet || 0)
              : undefined;

            setHistory(prev => [...prev, {
              id: `${newPlayer.user_id}-${Date.now()}`,
              playerName,
              action: actionName,
              amount,
              timestamp: new Date(),
            }]);
          }
        });
      }

      // Detect game mode based on player count
      const playerCount = newState.players.length;
      if (playerCount === 2) {
        setGameMode('heads_up');
      } else if (playerCount === 3) {
        setGameMode('3_player');
      } else if (playerCount >= 6) {
        setGameMode('6_player');
      }

      // Check if transitioning from handComplete to a new hand (hide modals)
      if (
        (tableState?.status === 'handComplete' && newState.status === 'waiting') ||
        (tableState?.status === 'handComplete' && newState.status === 'playing' && !newState.winners)
      ) {
        setShowHandComplete(false);
        setShowGameComplete(false);
        // Clear history when new hand starts
        setHistory([]);
      }

      setTableState(newState);

      // Show hand complete display when hand is complete (not game complete)
      if (newState.status === 'handComplete' && newState.winners && newState.winners.length > 0) {
        setShowHandComplete(true);
      }
    };

    const handleGameUpdate = (message: WSMessage) => {
      handleTableState(message);
    };

    const handleGameComplete = (message: WSMessage) => {
      // Show game complete modal (different from hand complete)
      showSuccess('Game complete!');
      setShowHandComplete(false); // Hide hand complete if showing
      setShowGameComplete(true);

      // Update table state with game complete info
      if (message.payload) {
        handleTableState(message);
      }
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
    // Note: History will be updated automatically when state changes are received
  }, [sendMessage]);

  const handleSendChatMessage = useCallback((message: string) => {
    // For now, just add to local state
    // TODO: Implement WebSocket chat when backend supports it
    const username = user?.username || 'Anonymous';
    setChatMessages(prev => [...prev, {
      id: Date.now().toString(),
      userId: currentUserId || '',
      username,
      message,
      timestamp: new Date(),
    }]);
  }, [currentUserId, user]);

  const handlePlayAgain = useCallback(() => {
    // Navigate to lobby and automatically join queue with same game mode
    navigate('/lobby', { state: { autoJoinQueue: true, gameMode } });
  }, [navigate, gameMode]);

  const handleReturnToLobby = useCallback(() => {
    navigate('/lobby');
  }, [navigate]);

  const handleLeaveGame = () => {
    navigate('/lobby');
  };

  return (
    <Box
      sx={{
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        position: 'relative',
        background: `linear-gradient(135deg, ${COLORS.background.primary} 0%, ${COLORS.background.secondary} 100%)`,
        overflow: 'hidden',

        // Ambient lighting effects
        '&::before': {
          content: '""',
          position: 'absolute',
          top: '-50%',
          left: '-50%',
          width: '200%',
          height: '200%',
          background: `
            radial-gradient(circle at 30% 30%, rgba(124, 58, 237, 0.15) 0%, transparent 40%),
            radial-gradient(circle at 70% 70%, rgba(6, 182, 212, 0.12) 0%, transparent 40%)
          `,
          pointerEvents: 'none',
          zIndex: 0,
          animation: 'ambientGlow 20s ease-in-out infinite',
        },

        // Subtle grain texture
        '&::after': {
          content: '""',
          position: 'absolute',
          inset: 0,
          backgroundImage: `
            repeating-linear-gradient(0deg, transparent, transparent 1px, rgba(255,255,255,0.01) 1px, rgba(255,255,255,0.01) 2px),
            repeating-linear-gradient(90deg, transparent, transparent 1px, rgba(255,255,255,0.01) 1px, rgba(255,255,255,0.01) 2px)
          `,
          pointerEvents: 'none',
          zIndex: 0,
        },

        '@keyframes ambientGlow': {
          '0%, 100%': {
            transform: 'translate(0, 0) scale(1)',
            opacity: 0.6,
          },
          '50%': {
            transform: 'translate(5%, 5%) scale(1.1)',
            opacity: 0.8,
          },
        },

        '& > *': {
          position: 'relative',
          zIndex: 1,
        },
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

      {/* Main game area - Full width for circular table */}
      <Box
        sx={{
          flex: 1,
          p: 3,
          overflow: 'hidden',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          position: 'relative',
        }}
      >
        <PokerTable tableState={tableState} currentUserId={currentUserId} />
      </Box>

      {/* Floating sidebar (renders in top-right corner) */}
      <GameSidebar
        history={history}
        messages={chatMessages}
        currentUserId={currentUserId}
        onSendMessage={handleSendChatMessage}
      />

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
          {/* Raise amount input */}
          {isMyTurn && (
            <Box sx={{ px: 2, maxWidth: 300, mx: 'auto' }}>
              <TextField
                type="number"
                label="Raise Amount"
                value={raiseAmount}
                onChange={(e) => setRaiseAmount(Number(e.target.value))}
                InputProps={{
                  startAdornment: <InputAdornment position="start">$</InputAdornment>,
                }}
                inputProps={{
                  min: minRaiseAmount,
                  max: maxRaiseAmount,
                  step: 10,
                }}
                helperText={`Min: $${minRaiseAmount} • Max: $${maxRaiseAmount}`}
                fullWidth
                size="small"
                sx={{
                  '& .MuiOutlinedInput-root': {
                    color: COLORS.text.primary,
                    '& fieldset': {
                      borderColor: COLORS.border.main,
                    },
                    '&:hover fieldset': {
                      borderColor: COLORS.primary.main,
                    },
                    '&.Mui-focused fieldset': {
                      borderColor: COLORS.primary.main,
                    },
                  },
                  '& .MuiInputLabel-root': {
                    color: COLORS.text.secondary,
                  },
                  '& .MuiFormHelperText-root': {
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

      {/* Hand Complete Display - Side panel with auto-hide */}
      {showHandComplete && tableState?.winners && tableState.winners.length > 0 && !showGameComplete && (
        <HandCompleteDisplay
          winners={tableState.winners}
          pot={tableState.pot}
          currentUserId={currentUserId}
          onClose={() => setShowHandComplete(false)}
        />
      )}

      {/* Game Complete Display - Full modal */}
      {showGameComplete && tableState?.winners && tableState.winners.length > 0 && (
        <WinnerDisplay
          winners={tableState.winners}
          pot={tableState.pot}
          gameComplete={true}
          gameMode={gameMode}
          onClose={() => setShowGameComplete(false)}
          onPlayAgain={handlePlayAgain}
          onReturnToLobby={handleReturnToLobby}
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
