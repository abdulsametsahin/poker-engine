import React, { useEffect, useState, useCallback } from 'react';
import { Box, Stack, IconButton, Dialog, DialogTitle, DialogContent, DialogActions, Typography } from '@mui/material';
import { ArrowBack, ExitToApp, Pause, EmojiEvents, Home } from '@mui/icons-material';
import { useParams, useNavigate } from 'react-router-dom';
import { useWebSocket } from '../contexts/WebSocketContext';
import { useAuth } from '../contexts/AuthContext';
import { useToast } from '../contexts/ToastContext';
import { PokerTable } from '../components/game/PokerTable';
import { GameSidebar } from '../components/game/GameSidebar';
import { TableSwitcher } from '../components/game/TableSwitcher';
import { WinnerDisplay, HandCompleteDisplay, TournamentPausedModal } from '../components/modals';
import { Button } from '../components/common/Button';
import { Badge } from '../components/common/Badge';
import { COLORS, RADIUS, GAME } from '../constants';
import {
  Player,
  WSMessage,
  ErrorPayload,
  TournamentPausedPayload,
  TournamentResumedPayload,
  TournamentCompletePayload,
  ChatMessagePayload
} from '../types';
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
  is_tournament?: boolean;
  action_sequence?: number;
}

interface PendingAction {
  type: string;
  amount?: number;
  requestId: string;
  timestamp: number;
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
  const [tournamentId, setTournamentId] = useState<string | null>(null);
  const [pendingAction, setPendingAction] = useState<PendingAction | null>(null);
  const [lastActionSequence, setLastActionSequence] = useState<number>(0);
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

  // Generate unique request ID for idempotency
  const generateRequestId = useCallback(() => {
    return `${currentUserId}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }, [currentUserId]);

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
      // Filter by table_id - only process messages for our table
      if (message.payload.table_id && message.payload.table_id !== tableId) {
        console.log(`[GameView] Ignoring table_state for table ${message.payload.table_id}, current table: ${tableId}`);
        return;
      }

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
        is_tournament: message.payload.is_tournament,
        action_sequence: message.payload.action_sequence || 0,
      };

      // Check if action sequence advanced (action was confirmed)
      if (newState.action_sequence > lastActionSequence) {
        setLastActionSequence(newState.action_sequence);

        // Clear pending action immediately when sequence advances
        // This indicates the server processed an action
        if (pendingAction) {
          setPendingAction(null);
        }
      }

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

      // Detect transitions between game states
      if (tableState?.status !== newState.status) {
        // Track state change in history (optional - can be removed if not needed)
        if (newState.status === 'playing' && tableState?.status !== 'playing') {
          // Add hand_started event to history
          setHistory(prev => [...prev, {
            id: `hand_started-${Date.now()}`,
            eventType: 'hand_started',
            timestamp: new Date(),
            metadata: {
              hand_number: prev.length + 1, // Simple hand counter
            },
          }]);
        }
      }

      // Track betting round changes and add to history
      if (tableState?.betting_round !== newState.betting_round && newState.betting_round) {
        const communityCards = newState.community_cards || [];

        // Add round_advanced event to history (for flop, turn, river)
        if (newState.betting_round !== 'preflop' && newState.betting_round !== 'waiting') {
          setHistory(prev => [...prev, {
            id: `round_advanced-${Date.now()}`,
            eventType: 'round_advanced',
            timestamp: new Date(),
            metadata: {
              new_round: newState.betting_round,
              round: newState.betting_round,
              community_cards: communityCards,
            },
          }]);
        }
      }

      setTableState(newState);

      // Show hand complete display when hand is complete (not game complete)
      if (newState.status === 'handComplete' && newState.winners && newState.winners.length > 0) {
        const winnersStr = newState.winners.map((w: any) =>
          `${w.playerName} (${w.handRank}) - ${w.amount} chips`
        ).join(', ');
        const communityCards = newState.community_cards || [];
        const cardsStr = communityCards.length > 0 ? ` - Board: ${communityCards.join(' ')}` : '';
        setShowHandComplete(true);

        // Add hand_complete event to history
        setHistory(prev => [...prev, {
          id: `hand_complete-${Date.now()}`,
          eventType: 'hand_complete',
          timestamp: new Date(),
          metadata: {
            winners: newState.winners,
            final_pot: newState.pot,
            pot: newState.pot,
          },
        }]);
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

    const handleError = (message: WSMessage<ErrorPayload>) => {
      showError(message.payload.message || 'An error occurred');
    };

    const handleTournamentPaused = (message: WSMessage<TournamentPausedPayload>) => {
      // Filter by tournament_id if we're in a tournament
      if (tournamentId && message.payload.tournament_id !== tournamentId) {
        console.log(`[GameView] Ignoring tournament_paused for ${message.payload.tournament_id}, current: ${tournamentId}`);
        return;
      }

      showWarning('Tournament has been paused. Game is on hold.');
      // Update table state to paused
      setTableState(prev => prev ? { ...prev, status: 'paused' } : null);
    };

    const handleTournamentResumed = (message: WSMessage<TournamentResumedPayload>) => {
      // Filter by tournament_id if we're in a tournament
      if (tournamentId && message.payload.tournament_id !== tournamentId) {
        console.log(`[GameView] Ignoring tournament_resumed for ${message.payload.tournament_id}, current: ${tournamentId}`);
        return;
      }

      showSuccess('Tournament has been resumed. Game continues!');
      // Update table state back to playing
      setTableState(prev => prev ? { ...prev, status: 'playing' } : null);
    };

    const handleTournamentComplete = (message: WSMessage<TournamentCompletePayload>) => {
      // Filter by tournament_id if we're in a tournament
      if (tournamentId && message.payload.tournament_id !== tournamentId) {
        console.log(`[GameView] Ignoring tournament_complete for ${message.payload.tournament_id}, current: ${tournamentId}`);
        return;
      }

      showSuccess(`Tournament complete! Winner: ${message.payload.winner_name}`);
      // Store tournament ID for navigation
      setTournamentId(message.payload.tournament_id);
    };

    const handleChatMessage = (message: WSMessage<ChatMessagePayload>) => {
      // Filter by table_id - only process messages for our table
      if (message.payload.table_id !== tableId) {
        console.log(`[GameView] Ignoring chat_message for table ${message.payload.table_id}, current: ${tableId}`);
        return;
      }

      const newMessage = {
        id: `${Date.now()}-${Math.random()}`,
        userId: message.payload.user_id,
        username: message.payload.username,
        message: message.payload.message,
        timestamp: new Date(message.payload.timestamp),
      };

      setChatMessages(prev => [...prev, newMessage]);
    };

    const handleActionConfirmed = (message: WSMessage) => {
      // Immediate confirmation from server that action was processed
      const { user_id, action } = message.payload;

      if (user_id === currentUserId && pendingAction?.type === action) {
        setPendingAction(null); // Clear immediately
      }
    };

    const handlePlayerActionBroadcast = (message: WSMessage) => {
      // Broadcast of player action to all players for history updates
      const { player_name, action, amount, timestamp } = message.payload;

      // Add to history
      setHistory(prev => {
        const now = Date.now();
        const recentDuplicate = prev.some(entry =>
          entry.playerName === player_name &&
          entry.action === action &&
          entry.amount === amount &&
          now - new Date(entry.timestamp).getTime() < 500
        );

        if (recentDuplicate) {
          return prev; // Skip duplicate
        }

        const amountStr = amount ? ` $${amount}` : '';

        return [...prev, {
          id: `${player_name}-${Date.now()}`,
          playerName: player_name,
          action,
          amount,
          timestamp: new Date(timestamp * 1000), // Convert Unix timestamp to Date
        }];
      });
    };

    // Register handlers and store cleanup functions
    const cleanup1 = addMessageHandler('table_state', handleTableState);
    const cleanup2 = addMessageHandler('game_update', handleGameUpdate);
    const cleanup3 = addMessageHandler('game_complete', handleGameComplete);
    const cleanup4 = addMessageHandler('error', handleError);
    const cleanup5 = addMessageHandler('tournament_paused', handleTournamentPaused);
    const cleanup6 = addMessageHandler('tournament_resumed', handleTournamentResumed);
    const cleanup7 = addMessageHandler('tournament_complete', handleTournamentComplete);
    const cleanup8 = addMessageHandler('chat_message', handleChatMessage);
    const cleanup9 = addMessageHandler('action_confirmed', handleActionConfirmed);
    const cleanup10 = addMessageHandler('player_action_broadcast', handlePlayerActionBroadcast);

    return () => {
      cleanup1();
      cleanup2();
      cleanup3();
      cleanup4();
      cleanup5();
      cleanup6();
      cleanup7();
      cleanup8();
      cleanup9();
      cleanup10();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps

  const handleAction = useCallback((action: string, amount?: number) => {
    // Prevent multiple actions
    if (pendingAction) {
      return;
    }

    // Client-side validation
    if (!isMyTurn) {
      return;
    }

    if (tableState?.status !== 'playing') {
      return;
    }

    const requestId = generateRequestId();
    const amountStr = amount ? ` $${amount}` : '';

    // Set pending state IMMEDIATELY (disables buttons)
    setPendingAction({
      type: action,
      amount,
      requestId,
      timestamp: Date.now(),
    });


    // Send action to server with request_id
    sendMessage({
      type: 'game_action',
      payload: {
        action,
        amount: amount || 0,
        request_id: requestId,
        timestamp: Date.now(),
      },
    });

    // Timeout fallback: Clear pending state after 5 seconds if no confirmation
    setTimeout(() => {
      setPendingAction(prev => {
        if (prev && prev.requestId === requestId) {
          return null;
        }
        return prev;
      });
    }, 5000);

    // Note: History will be updated automatically when state changes are received

  const handleSendChatMessage = useCallback((message: string) => {
    if (!message.trim() || !tableId || !user) return;

    const username = user.username || 'Anonymous';

    // Optimistic update - add to local state immediately
    const tempMessage = {
      id: `temp-${Date.now()}`,
      userId: currentUserId || user.id,
      username,
      message: message.trim(),
      timestamp: new Date(),
    };
    setChatMessages(prev => [...prev, tempMessage]);

    // Send to server via WebSocket
    sendMessage({
      type: 'chat_message',
      payload: {
        table_id: tableId,
        user_id: user.id,
        username,
        message: message.trim(),
        timestamp: new Date().toISOString(),
      }
    });


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
              variant={
                tableState.status === 'playing' ? 'primary'
                : tableState.status === 'completed' ? 'success'
                : 'secondary'
              }
              pulse={tableState.status === 'playing'}
            >
              {tableState.status.toUpperCase()}
            </Badge>
          )}
        </Stack>

        <Stack direction="row" spacing={1}>
          <TableSwitcher />
          
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

          {/* Tournament Complete Overlay */}
          {tableState?.status === 'completed' && tableState?.is_tournament && (
            <Box
              sx={{
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                backgroundColor: 'rgba(0, 0, 0, 0.85)',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                zIndex: 1000,
              }}
            >
              <EmojiEvents sx={{ fontSize: 100, color: COLORS.success.main, mb: 3 }} />
              <Typography variant="h3" sx={{ color: 'white', mb: 2, fontWeight: 'bold' }}>
                Tournament Complete!
              </Typography>
              <Typography variant="body1" sx={{ color: COLORS.text.secondary, mb: 4 }}>
                The tournament has ended. Check the tournament page for final standings.
              </Typography>
              <Button
                variant="primary"
                onClick={() => tournamentId ? navigate(`/tournaments/${tournamentId}`) : navigate('/tournaments')}
                sx={{
                  py: 1.5,
                  px: 4,
                  fontSize: '16px',
                  fontWeight: 700,
                }}
              >
                <Home sx={{ mr: 1 }} />
                RETURN TO TOURNAMENT
              </Button>
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

      {/* Pending action indicator */}
      {pendingAction && tableState?.status === 'playing' && (
        <Box
          sx={{
            px: 4,
            py: 2,
            background: 'linear-gradient(90deg, rgba(124, 58, 237, 0.2) 0%, rgba(6, 182, 212, 0.2) 100%)',
            backdropFilter: 'blur(10px)',
            borderBottom: `1px solid ${COLORS.primary.main}40`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 2,
            zIndex: 11,
          }}
        >
          <Box
            sx={{
              width: 20,
              height: 20,
              borderRadius: '50%',
              border: `3px solid ${COLORS.primary.main}`,
              borderTopColor: 'transparent',
              animation: 'spin 1s linear infinite',
              '@keyframes spin': {
                '0%': { transform: 'rotate(0deg)' },
                '100%': { transform: 'rotate(360deg)' },
              },
            }}
          />
          <Typography sx={{ color: COLORS.text.primary, fontSize: '14px', fontWeight: 600 }}>
            Processing {pendingAction.type}...
          </Typography>
        </Box>
      )}

      {/* Action bar - Only show when playing and it's my turn */}
      {tableState?.status === 'playing' && isMyTurn && (
        <Box
          sx={{
            px: 4,
            py: 3,
            background: 'linear-gradient(180deg, rgba(0, 0, 0, 0.9) 0%, rgba(0, 0, 0, 0.95) 100%)',
            backdropFilter: 'blur(20px)',
            borderTop: `2px solid ${COLORS.primary.main}40`,
            position: 'relative',
            zIndex: 10,
            boxShadow: '0 -8px 32px rgba(0, 0, 0, 0.5)',

            // Animated glow effect on top border
            '&::before': {
              content: '""',
              position: 'absolute',
              top: -2,
              left: 0,
              right: 0,
              height: 2,
              background: `linear-gradient(90deg,
                transparent 0%,
                ${COLORS.primary.main}80 20%,
                ${COLORS.primary.main} 50%,
                ${COLORS.primary.main}80 80%,
                transparent 100%)`,
              animation: 'borderGlow 3s ease-in-out infinite',
            },

            '@keyframes borderGlow': {
              '0%, 100%': { opacity: 0.6 },
              '50%': { opacity: 1 },
            },
          }}
        >
          <Stack spacing={2.5}>
            {/* Top row: Basic actions */}
            <Stack direction="row" spacing={2} alignItems="center" justifyContent="center">
              {/* Fold button - Danger style */}
              <Button
                variant="danger"
                onClick={() => handleAction('fold')}
                disabled={!!pendingAction}
                sx={{
                  minWidth: 120,
                  height: 48,
                  fontSize: '14px',
                  fontWeight: 600,
                  px: 3,
                  borderRadius: RADIUS.md,
                  boxShadow: `0 4px 12px ${COLORS.danger.main}40`,
                  transition: 'all 0.3s ease',
                  '&:hover': {
                    transform: 'translateY(-2px)',
                    boxShadow: `0 6px 16px ${COLORS.danger.main}60`,
                  },
                  '&:active': {
                    transform: 'translateY(0)',
                  },
                }}
              >
                FOLD
              </Button>

              {/* Check/Call buttons */}
              {currentBet <= playerBet ? (
                <Button
                  variant="secondary"
                  onClick={() => handleAction('check')}
                  disabled={!!pendingAction}
                  sx={{
                    minWidth: 120,
                    height: 48,
                    fontSize: '14px',
                    fontWeight: 600,
                    px: 3,
                    borderRadius: RADIUS.md,
                    transition: 'all 0.3s ease',
                    '&:hover': {
                      transform: 'translateY(-2px)',
                    },
                    '&:active': {
                      transform: 'translateY(0)',
                    },
                  }}
                >
                  CHECK
                </Button>
              ) : (
                <Button
                  variant="success"
                  onClick={() => handleAction('call')}
                  disabled={!!pendingAction}
                  sx={{
                    minWidth: 140,
                    height: 48,
                    fontSize: '14px',
                    fontWeight: 600,
                    px: 3,
                    borderRadius: RADIUS.md,
                    boxShadow: `0 4px 12px ${COLORS.success.main}40`,
                    transition: 'all 0.3s ease',
                    '&:hover': {
                      transform: 'translateY(-2px)',
                      boxShadow: `0 6px 16px ${COLORS.success.main}60`,
                    },
                    '&:active': {
                      transform: 'translateY(0)',
                    },
                  }}
                >
                  <Stack direction="row" alignItems="center" spacing={1}>
                    <span>CALL</span>
                    <Box
                      sx={{
                        px: 1.5,
                        py: 0.5,
                        borderRadius: RADIUS.sm,
                        background: 'rgba(255, 255, 255, 0.2)',
                        fontSize: '13px',
                        fontWeight: 700,
                      }}
                    >
                      ${callAmount}
                    </Box>
                  </Stack>
                </Button>
              )}

              {/* All-in button - Prominent */}
              <Button
                variant="warning"
                onClick={() => handleAction('allin')}
                disabled={!!pendingAction}
                sx={{
                  minWidth: 140,
                  height: 48,
                  fontSize: '14px',
                  fontWeight: 700,
                  px: 3,
                  borderRadius: RADIUS.md,
                  background: `linear-gradient(135deg, ${COLORS.warning.main} 0%, ${COLORS.warning.dark} 100%)`,
                  boxShadow: `0 4px 12px ${COLORS.warning.main}50`,
                  position: 'relative',
                  overflow: 'hidden',
                  transition: 'all 0.3s ease',
                  
                  // Animated shine effect
                  '&::before': {
                    content: '""',
                    position: 'absolute',
                    top: 0,
                    left: '-100%',
                    width: '100%',
                    height: '100%',
                    background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.3), transparent)',
                    animation: 'shine 3s infinite',
                  },
                  
                  '&:hover': {
                    transform: 'translateY(-2px)',
                    boxShadow: `0 6px 16px ${COLORS.warning.main}70`,
                    background: `linear-gradient(135deg, ${COLORS.warning.light} 0%, ${COLORS.warning.main} 100%)`,
                  },
                  '&:active': {
                    transform: 'translateY(0)',
                  },
                  
                  '@keyframes shine': {
                    '0%': { left: '-100%' },
                    '50%, 100%': { left: '200%' },
                  },
                }}
              >
                <Stack direction="row" alignItems="center" spacing={1}>
                  <span>ALL IN</span>
                  <Box
                    sx={{
                      px: 1.5,
                      py: 0.5,
                      borderRadius: RADIUS.sm,
                      background: 'rgba(0, 0, 0, 0.3)',
                      fontSize: '12px',
                      fontWeight: 700,
                    }}
                  >
                    ${maxRaiseAmount}
                  </Box>
                </Stack>
              </Button>
            </Stack>

            {/* Bottom row: Raise controls */}
            <Box
              sx={{
                p: 2.5,
                borderRadius: RADIUS.md,
                background: 'rgba(124, 58, 237, 0.08)',
                border: `1px solid ${COLORS.primary.main}30`,
              }}
            >
              <Stack spacing={2}>
                {/* Raise amount display and slider */}
                <Stack direction="row" spacing={3} alignItems="center">
                  <Typography
                    sx={{
                      color: COLORS.text.secondary,
                      fontSize: '13px',
                      fontWeight: 600,
                      minWidth: 80,
                    }}
                  >
                    RAISE TO
                  </Typography>

                  {/* Custom amount input */}
                  <Box
                    sx={{
                      px: 2,
                      py: 1,
                      borderRadius: RADIUS.md,
                      background: 'rgba(0, 0, 0, 0.5)',
                      border: `1px solid ${COLORS.primary.main}40`,
                      display: 'flex',
                      alignItems: 'center',
                      minWidth: 140,
                    }}
                  >
                    <Typography
                      sx={{
                        color: COLORS.text.secondary,
                        fontSize: '16px',
                        mr: 1,
                      }}
                    >
                      $
                    </Typography>
                    <input
                      type="number"
                      value={raiseAmount}
                      onChange={(e) => setRaiseAmount(Number(e.target.value))}
                      min={minRaiseAmount}
                      max={maxRaiseAmount}
                      step={10}
                      style={{
                        background: 'transparent',
                        border: 'none',
                        outline: 'none',
                        color: COLORS.text.primary,
                        fontSize: '20px',
                        fontWeight: 700,
                        width: '100%',
                        fontFamily: 'inherit',
                      }}
                    />
                  </Box>

                  {/* Quick bet buttons */}
                  <Stack direction="row" spacing={1} flex={1}>
                    {[
                      { label: 'MIN', value: minRaiseAmount },
                      { label: '½ POT', value: Math.min(Math.floor((tableState?.pot || 0) * 0.5), maxRaiseAmount) },
                      { label: 'POT', value: Math.min((tableState?.pot || 0), maxRaiseAmount) },
                      { label: 'MAX', value: maxRaiseAmount },
                    ].map((bet) => (
                      <Button
                        key={bet.label}
                        variant="ghost"
                        onClick={() => setRaiseAmount(bet.value)}
                        sx={{
                          minWidth: 0,
                          height: 36,
                          px: 2,
                          fontSize: '11px',
                          fontWeight: 600,
                          borderRadius: RADIUS.sm,
                          border: `1px solid ${COLORS.primary.main}40`,
                          background: raiseAmount === bet.value ? `${COLORS.primary.main}30` : 'transparent',
                          color: raiseAmount === bet.value ? COLORS.primary.light : COLORS.text.secondary,
                          transition: 'all 0.2s ease',
                          '&:hover': {
                            background: `${COLORS.primary.main}20`,
                            borderColor: COLORS.primary.main,
                            color: COLORS.primary.light,
                          },
                        }}
                      >
                        {bet.label}
                      </Button>
                    ))}
                  </Stack>

                  {/* Raise action button */}
                  <Button
                    variant="primary"
                    onClick={() => handleAction('raise', raiseAmount)}
                    disabled={!!pendingAction || raiseAmount < minRaiseAmount || raiseAmount > maxRaiseAmount}
                    sx={{
                      minWidth: 120,
                      height: 48,
                      fontSize: '14px',
                      fontWeight: 700,
                      px: 3,
                      borderRadius: RADIUS.md,
                      boxShadow: `0 4px 12px ${COLORS.primary.main}40`,
                      transition: 'all 0.3s ease',
                      '&:hover:not(:disabled)': {
                        transform: 'translateY(-2px)',
                        boxShadow: `0 6px 16px ${COLORS.primary.main}60`,
                      },
                      '&:active:not(:disabled)': {
                        transform: 'translateY(0)',
                      },
                      '&:disabled': {
                        opacity: 0.4,
                        cursor: 'not-allowed',
                      },
                    }}
                  >
                    RAISE
                  </Button>
                </Stack>

                {/* Range info */}
                <Stack direction="row" justifyContent="space-between" alignItems="center">
                  <Typography
                    variant="caption"
                    sx={{
                      color: COLORS.text.secondary,
                      fontSize: '11px',
                    }}
                  >
                    Min: <strong style={{ color: COLORS.text.primary }}>${minRaiseAmount}</strong>
                  </Typography>
                  <Typography
                    variant="caption"
                    sx={{
                      color: COLORS.text.secondary,
                      fontSize: '11px',
                    }}
                  >
                    Your chips: <strong style={{ color: COLORS.primary.light }}>${currentPlayer?.chips || 0}</strong>
                  </Typography>
                  <Typography
                    variant="caption"
                    sx={{
                      color: COLORS.text.secondary,
                      fontSize: '11px',
                    }}
                  >
                    Max: <strong style={{ color: COLORS.text.primary }}>${maxRaiseAmount}</strong>
                  </Typography>
                </Stack>
              </Stack>
            </Box>
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

      {/* Tournament Paused Modal */}
      <TournamentPausedModal open={tableState?.status === 'paused'} />

      {/* Leave game confirmation dialog */}
    </Box>
  );
};
