import React, { useEffect, useState, useCallback } from 'react';
import { Box, Stack, IconButton, Dialog, DialogTitle, DialogContent, DialogActions, TextField, InputAdornment, Typography } from '@mui/material';
import { ArrowBack, ExitToApp, Terminal, Pause } from '@mui/icons-material';
import { useParams, useNavigate } from 'react-router-dom';
import { useWebSocket } from '../contexts/WebSocketContext';
import { useAuth } from '../contexts/AuthContext';
import { useToast } from '../contexts/ToastContext';
import { PokerTable } from '../components/game/PokerTable';
import { GameSidebar } from '../components/game/GameSidebar';
import { ConsolePanel } from '../components/game/ConsolePanel';
import { TableSwitcher } from '../components/game/TableSwitcher';
import { WinnerDisplay, HandCompleteDisplay } from '../components/modals';
import { Button } from '../components/common/Button';
import { Badge } from '../components/common/Badge';
import { COLORS, RADIUS, SPACING, GAME } from '../constants';
import { Player, WSMessage } from '../types';
import { addActiveTable, updateTableActivity, removeActiveTable } from '../utils/tableManager';

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
  const [consoleOpen, setConsoleOpen] = useState(false);
  const [consoleLogs, setConsoleLogs] = useState<any[]>([]);
  const [history, setHistory] = useState<any[]>(() => {
    // Load history from localStorage on mount
    try {
      const savedHistory = localStorage.getItem(`game_history_${tableId}`);
      return savedHistory ? JSON.parse(savedHistory, (key, value) => {
        // Restore Date objects
        if (key === 'timestamp') return new Date(value);
        return value;
      }) : [];
    } catch (error) {
      console.error('Failed to load history from localStorage:', error);
      return [];
    }
  });
  const [chatMessages, setChatMessages] = useState<any[]>([]);

  // Find current user
  const currentUserId = user?.id || tableState?.players?.find(p => p.cards && p.cards.length > 0)?.user_id;
  const currentPlayer = tableState?.players?.find(p => p.user_id === currentUserId);

  // Helper function to add console logs
  const addConsoleLog = useCallback((category: string, message: string, level: 'info' | 'warning' | 'error' | 'success' | 'debug' = 'info') => {
    setConsoleLogs(prev => [
      ...prev,
      {
        id: `${Date.now()}-${Math.random()}`,
        timestamp: new Date(),
        level,
        category,
        message,
      },
    ]);
  }, []);
  const isMyTurn = tableState?.current_turn === currentUserId;

  // Save history to localStorage whenever it changes
  useEffect(() => {
    if (tableId && history.length > 0) {
      try {
        localStorage.setItem(`game_history_${tableId}`, JSON.stringify(history));
      } catch (error) {
        console.error('Failed to save history to localStorage:', error);
      }
    }
  }, [history, tableId]);

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

  // Initialize raise amount to min bet when it becomes player's turn
  React.useEffect(() => {
    if (isMyTurn && tableState?.status === 'playing') {
      setRaiseAmount(minRaiseAmount);
    }
  }, [isMyTurn, minRaiseAmount, tableState?.status]);

  // Subscribe to table on mount and track active table
  useEffect(() => {
    if (isConnected && tableId) {
      sendMessage({
        type: 'subscribe_table',
        payload: { table_id: tableId },
      });
      
      // Add to active tables list
      addActiveTable(tableId);
    }
    
    // Cleanup: remove table from active list when leaving
    return () => {
      if (tableId) {
        removeActiveTable(tableId);
      }
    };
  }, [isConnected, tableId, sendMessage]);
  
  // Update table activity periodically
  useEffect(() => {
    if (!tableId) return;
    
    const interval = setInterval(() => {
      updateTableActivity(tableId);
    }, 30000); // Update every 30 seconds
    
    return () => clearInterval(interval);
  }, [tableId]);

  // Handle WebSocket messages
  useEffect(() => {
    const handleTableState = (message: WSMessage) => {
      // Log the message receipt
      addConsoleLog('WEBSOCKET', `Received ${message.type} message`, 'debug');

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

            // Use last_action_amount from backend if available, otherwise undefined
            const amount = newPlayer.last_action_amount !== undefined && newPlayer.last_action_amount > 0
              ? newPlayer.last_action_amount
              : undefined;

            // Check if this exact action was already added in the last 500ms to prevent duplicates
            setHistory(prev => {
              const now = Date.now();
              const recentDuplicate = prev.some(entry =>
                entry.playerName === playerName &&
                entry.action === actionName &&
                entry.amount === amount &&
                now - new Date(entry.timestamp).getTime() < 500
              );

              if (recentDuplicate) {
                return prev; // Skip duplicate
              }

              // Log to console only if not a duplicate
              const amountStr = amount ? ` $${amount}` : '';
              addConsoleLog('ACTION', `${playerName} ${actionName}${amountStr}`, 'info');

              return [...prev, {
                id: `${newPlayer.user_id}-${Date.now()}`,
                playerName,
                action: actionName,
                amount,
                timestamp: new Date(),
              }];
            });
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
        // Don't clear history - keep it persistent across hands
      }

      // Log game state changes
      if (tableState?.status !== newState.status) {
        addConsoleLog('GAME_STATE', `Status changed: ${tableState?.status || 'none'} → ${newState.status}`, 'info');

        // Log hand start
        if (newState.status === 'playing' && tableState?.status !== 'playing') {
          addConsoleLog('HAND_START', 'New hand started', 'success');
        }
      }

      if (tableState?.betting_round !== newState.betting_round && newState.betting_round) {
        const communityCards = newState.community_cards || [];
        const cardsStr = communityCards.length > 0 ? ` - Cards: ${communityCards.join(' ')}` : '';
        addConsoleLog('GAME_STATE', `Betting round: ${newState.betting_round}${cardsStr}`, 'info');
      }

      setTableState(newState);

      // Show hand complete display when hand is complete (not game complete)
      if (newState.status === 'handComplete' && newState.winners && newState.winners.length > 0) {
        const winnersStr = newState.winners.map((w: any) =>
          `${w.playerName} (${w.handRank}) - ${w.amount} chips`
        ).join(', ');
        const communityCards = newState.community_cards || [];
        const cardsStr = communityCards.length > 0 ? ` - Board: ${communityCards.join(' ')}` : '';
        addConsoleLog('HAND_COMPLETE', `Winners: ${winnersStr}${cardsStr}`, 'success');
        addConsoleLog('HAND_COMPLETE', `Pot: ${newState.pot} chips`, 'info');
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

      // Clear history from localStorage when game ends
      if (tableId) {
        try {
          localStorage.removeItem(`game_history_${tableId}`);
        } catch (error) {
          console.error('Failed to clear history from localStorage:', error);
        }
      }

      // Update table state with game complete info
      if (message.payload) {
        handleTableState(message);
      }
    };

    const handleError = (message: WSMessage) => {
      showError(message.payload.message || 'An error occurred');
    };

    const handleTournamentPaused = (message: WSMessage) => {
      addConsoleLog('TOURNAMENT', 'Tournament paused - Game on hold', 'warning');
      showWarning('Tournament has been paused. Game is on hold.');
    };

    const handleTournamentResumed = (message: WSMessage) => {
      addConsoleLog('TOURNAMENT', 'Tournament resumed - Game continuing', 'success');
      showSuccess('Tournament has been resumed. Game continues!');
    };

    addMessageHandler('table_state', handleTableState);
    addMessageHandler('game_update', handleGameUpdate);
    addMessageHandler('game_complete', handleGameComplete);
    addMessageHandler('error', handleError);
    addMessageHandler('tournament_paused', handleTournamentPaused);
    addMessageHandler('tournament_resumed', handleTournamentResumed);

    return () => {
      removeMessageHandler('table_state');
      removeMessageHandler('game_update');
      removeMessageHandler('game_complete');
      removeMessageHandler('error');
      removeMessageHandler('tournament_paused');
      removeMessageHandler('tournament_resumed');
    };
  }, [tableId, tableState?.status, tableState?.betting_round, tableState?.players, addMessageHandler, removeMessageHandler, showSuccess, showError, showWarning, addConsoleLog]);

  const handleAction = useCallback((action: string, amount?: number) => {
    const amountStr = amount ? ` $${amount}` : '';
    addConsoleLog('ACTION', `You ${action}${amountStr}`, 'success');

    sendMessage({
      type: 'game_action',
      payload: { action, amount: amount || 0 },
    });
    // Note: History will be updated automatically when state changes are received
  }, [sendMessage, addConsoleLog]);

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

  return (
    <Box
      sx={{
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        background: `linear-gradient(135deg, ${COLORS.background.primary} 0%, ${COLORS.background.secondary} 100%)`,
        overflow: 'hidden',
        position: 'relative',

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
          position: 'relative',
          zIndex: 10,
        }}
      >
        <Stack direction="row" spacing={2} alignItems="center">
          <IconButton
            onClick={() => navigate('/lobby')}
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

        <Stack direction="row" spacing={1}>
          <TableSwitcher />
          
          <IconButton
            onClick={() => setConsoleOpen(true)}
            sx={{
              color: COLORS.info.main,
              '&:hover': {
                color: COLORS.info.light,
                background: `${COLORS.info.main}20`,
              },
            }}
            title="View Console Logs"
          >
            <Terminal />
          </IconButton>

          <IconButton
            onClick={() => navigate('/lobby')}
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
        </Stack>
      </Box>

      {/* Main content area - Two column layout */}
      <Box
        sx={{
          flex: 1,
          display: 'flex',
          overflow: 'hidden',
          position: 'relative',
          zIndex: 1,
        }}
      >
        {/* Left side - Poker table */}
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

          {/* Paused Overlay - Game on Hold */}
          {tableState?.status === 'paused' && (
            <Box
              sx={{
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                backgroundColor: 'rgba(0, 0, 0, 0.7)',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                zIndex: 1000,
              }}
            >
              <Pause sx={{ fontSize: 80, color: COLORS.warning.main, mb: 2 }} />
              <Typography variant="h4" sx={{ color: 'white', mb: 1, fontWeight: 'bold' }}>
                Game on Hold
              </Typography>
              <Typography variant="body1" sx={{ color: COLORS.text.secondary }}>
                Tournament is currently paused. Waiting for resume...
              </Typography>
            </Box>
          )}
        </Box>

        {/* Right side - Sidebar */}
        <GameSidebar
          history={history}
          messages={chatMessages}
          currentUserId={currentUserId}
          onSendMessage={handleSendChatMessage}
        />
      </Box>

      {/* Action bar - Only show when playing and it's my turn */}
      {tableState?.status === 'playing' && isMyTurn && (
        <Box
          sx={{
            px: 3,
            py: 1.5,
            background: 'rgba(0, 0, 0, 0.6)',
            backdropFilter: 'blur(10px)',
            borderTop: `1px solid ${COLORS.border.main}`,
            position: 'relative',
            zIndex: 10,
          }}
        >
          <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
            {/* Primary action buttons */}
            <Button
              variant="danger"
              onClick={() => handleAction('fold')}
              sx={{
                minWidth: 70,
                height: 36,
                fontSize: '12px',
                px: 2,
              }}
            >
              FOLD
            </Button>

            <Button
              variant="secondary"
              onClick={() => handleAction('check')}
              disabled={currentBet > playerBet}
              sx={{
                minWidth: 70,
                height: 36,
                fontSize: '12px',
                px: 2,
              }}
            >
              CHECK
            </Button>

            <Button
              variant="success"
              onClick={() => handleAction('call')}
              disabled={callAmount <= 0}
              sx={{
                minWidth: 70,
                height: 36,
                fontSize: '12px',
                px: 2,
              }}
            >
              CALL {callAmount > 0 && `$${callAmount}`}
            </Button>

            {/* Divider */}
            <Box sx={{ width: 1, height: 24, bgcolor: COLORS.border.main, mx: 0.5 }} />

            {/* Raise amount input - compact and on the left */}
            <TextField
              type="number"
              value={raiseAmount}
              onChange={(e) => setRaiseAmount(Number(e.target.value))}
              InputProps={{
                startAdornment: <InputAdornment position="start" sx={{ mr: 0.5 }}>$</InputAdornment>,
              }}
              inputProps={{
                min: minRaiseAmount,
                max: maxRaiseAmount,
                step: 10,
              }}
              size="small"
              sx={{
                width: 100,
                '& .MuiOutlinedInput-root': {
                  height: 36,
                  color: COLORS.text.primary,
                  fontSize: '13px',
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
                '& input': {
                  py: 0.75,
                  px: 1,
                },
              }}
            />

            {/* Raise button */}
            <Button
              variant="primary"
              onClick={() => handleAction('raise', raiseAmount)}
              disabled={raiseAmount < minRaiseAmount}
              sx={{
                minWidth: 70,
                height: 36,
                fontSize: '12px',
                px: 2,
              }}
            >
              RAISE
            </Button>

            {/* All-in button */}
            <Button
              variant="warning"
              onClick={() => handleAction('allin')}
              sx={{
                minWidth: 70,
                height: 36,
                fontSize: '12px',
                px: 2,
                background: `linear-gradient(135deg, ${COLORS.warning.main} 0%, ${COLORS.warning.dark} 100%)`,
                '&:hover': {
                  background: `linear-gradient(135deg, ${COLORS.warning.light} 0%, ${COLORS.warning.main} 100%)`,
                },
              }}
            >
              ALL-IN
            </Button>

            {/* Helper text */}
            <Typography
              variant="caption"
              sx={{
                color: COLORS.text.secondary,
                fontSize: '10px',
                ml: 2,
              }}
            >
              Min: ${minRaiseAmount} • Max: ${maxRaiseAmount}
            </Typography>
          </Stack>
        </Box>
      )}

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
      {/* Console Dialog */}
      <Dialog
        open={consoleOpen}
        onClose={() => setConsoleOpen(false)}
        maxWidth="lg"
        fullWidth
        PaperProps={{
          sx: {
            background: COLORS.background.paper,
            borderRadius: RADIUS.md,
            border: `1px solid ${COLORS.border.main}`,
            height: '80vh',
            maxHeight: '800px',
          },
        }}
      >
        <DialogTitle
          sx={{
            color: COLORS.text.primary,
            display: 'flex',
            alignItems: 'center',
            gap: 1,
            borderBottom: `1px solid ${COLORS.border.main}`,
          }}
        >
          <Terminal sx={{ color: COLORS.info.main }} />
          Console Logs - Table {tableId}
        </DialogTitle>
        <DialogContent sx={{ p: 0, height: 'calc(100% - 120px)' }}>
          <ConsolePanel logs={consoleLogs} />
        </DialogContent>
        <DialogActions sx={{ borderTop: `1px solid ${COLORS.border.main}`, p: 2 }}>
          <Button
            variant="ghost"
            onClick={() => {
              setConsoleLogs([]);
            }}
          >
            Clear Logs
          </Button>
          <Button variant="primary" onClick={() => setConsoleOpen(false)}>
            Close
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};
