import React from 'react';
import { Box, Paper, Typography, Stack, Chip, Avatar, Badge, CircularProgress } from '@mui/material';
import { AccountCircle, AttachMoney, HourglassEmpty } from '@mui/icons-material';
import PlayingCard from './PlayingCard';
import ActionTimer from './ActionTimer';

interface Player {
  user_id: string;
  seat_number: number;
  chips: number;
  status: string;
  bet?: number;
  cards?: string[];
  last_action?: string;
  is_dealer?: boolean;
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

interface PokerTableProps {
  tableState: TableState | null;
}

const PlayerSeat: React.FC<{
  player: Player | null;
  position: number;
  isActive: boolean;
  actionDeadline?: string;
}> = ({ player, position, isActive, actionDeadline }) => {
  if (!player) {
    return (
      <Box
        sx={{
          minWidth: 140,
          height: 120,
          borderRadius: 2,
          border: '1px dashed rgba(75, 85, 99, 0.3)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          bgcolor: 'rgba(31, 41, 55, 0.2)',
          opacity: 0.4,
        }}
      >
        <Typography variant="caption" sx={{ color: 'rgba(156, 163, 175, 0.5)', fontSize: '11px', textAlign: 'center' }}>
          Seat {position}
          <br />
          Empty
        </Typography>
      </Box>
    );
  }

  return (
    <Box
      sx={{
        minWidth: 140,
        height: 120,
        p: 1.5,
        position: 'relative',
        borderRadius: 2,
        bgcolor: isActive ? 'rgba(251, 191, 36, 0.1)' : 'rgba(31, 41, 55, 0.6)',
        border: isActive ? '2px solid rgba(251, 191, 36, 0.5)' : '1px solid rgba(75, 85, 99, 0.3)',
        backdropFilter: 'blur(10px)',
        transition: 'all 0.3s',
        animation: isActive ? 'pulse 2s infinite' : 'none',
        opacity: player.status === 'folded' ? 0.5 : 1,
        '@keyframes pulse': {
          '0%, 100%': { boxShadow: '0 0 15px rgba(251, 191, 36, 0.3)' },
          '50%': { boxShadow: '0 0 25px rgba(251, 191, 36, 0.5)' },
        },
      }}
    >
      {/* Dealer Button */}
      {player.is_dealer && (
        <Box
          sx={{
            position: 'absolute',
            top: -8,
            right: -8,
            width: 20,
            height: 20,
            borderRadius: '50%',
            bgcolor: '#ef4444',
            color: 'white',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '10px',
            fontWeight: 900,
            border: '2px solid #0f1419',
            boxShadow: '0 2px 8px rgba(239, 68, 68, 0.4)',
          }}
        >
          D
        </Box>
      )}

      {/* Fold Badge */}
      {player.status === 'folded' && (
        <Box
          sx={{
            position: 'absolute',
            top: -8,
            left: -8,
            px: 1,
            py: 0.25,
            borderRadius: 1,
            bgcolor: 'rgba(239, 68, 68, 0.9)',
            color: 'white',
            fontSize: '9px',
            fontWeight: 700,
            border: '1px solid rgba(239, 68, 68, 0.5)',
          }}
        >
          FOLDED
        </Box>
      )}

      <Stack spacing={0.75} sx={{ height: '100%' }}>
        {/* Player Info */}
        <Stack direction="row" spacing={1} alignItems="center">
          <Avatar
            sx={{
              width: 32,
              height: 32,
              bgcolor: isActive ? 'rgba(251, 191, 36, 0.3)' : 'rgba(99, 102, 241, 0.3)',
              border: `2px solid ${isActive ? 'rgba(251, 191, 36, 0.5)' : 'rgba(99, 102, 241, 0.5)'}`,
            }}
          >
            <AccountCircle sx={{ fontSize: 24, color: isActive ? '#fbbf24' : '#6366f1' }} />
          </Avatar>
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Typography
              variant="caption"
              fontWeight="bold"
              noWrap
              sx={{ color: '#fff', fontSize: '12px', display: 'block' }}
            >
              {player.user_id.slice(0, 8)}
            </Typography>
            <Typography variant="caption" sx={{ color: 'rgba(156, 163, 175, 0.7)', fontSize: '10px' }}>
              Seat {position}
            </Typography>
          </Box>
        </Stack>

        {/* Cards */}
        {player.cards && player.cards.length > 0 && (
          <Stack direction="row" spacing={0.5} justifyContent="center">
            {player.cards.map((card, idx) => (
              <PlayingCard key={idx} card={card} size="small" />
            ))}
          </Stack>
        )}

        {/* Chips & Bet */}
        <Stack direction="row" spacing={0.5} sx={{ mt: 'auto' }}>
          <Box
            sx={{
              flex: 1,
              px: 1,
              py: 0.5,
              borderRadius: 1,
              bgcolor: player.chips > 500 ? 'rgba(16, 185, 129, 0.15)' : 'rgba(251, 191, 36, 0.15)',
              border: `1px solid ${player.chips > 500 ? 'rgba(16, 185, 129, 0.3)' : 'rgba(251, 191, 36, 0.3)'}`,
              textAlign: 'center',
            }}
          >
            <Typography
              variant="caption"
              fontWeight="bold"
              sx={{ color: player.chips > 500 ? '#10b981' : '#fbbf24', fontSize: '11px' }}
            >
              ${player.chips}
            </Typography>
          </Box>
          {player.bet !== undefined && player.bet > 0 && (
            <Box
              sx={{
                px: 1,
                py: 0.5,
                borderRadius: 1,
                bgcolor: 'rgba(239, 68, 68, 0.15)',
                border: '1px solid rgba(239, 68, 68, 0.3)',
                textAlign: 'center',
              }}
            >
              <Typography variant="caption" fontWeight="bold" sx={{ color: '#ef4444', fontSize: '11px' }}>
                ${player.bet}
              </Typography>
            </Box>
          )}
        </Stack>

        {/* Action Timer */}
        {isActive && actionDeadline && (
          <Box sx={{ width: '100%' }}>
            <ActionTimer deadline={actionDeadline} />
          </Box>
        )}
      </Stack>
    </Box>
  );
};

const PokerTable: React.FC<PokerTableProps> = ({ tableState }) => {
  if (!tableState) {
    return (
      <Box sx={{ p: 4, textAlign: 'center', bgcolor: 'rgba(31, 41, 55, 0.5)', borderRadius: 2 }}>
        <CircularProgress size={32} sx={{ color: '#6366f1' }} />
        <Typography variant="body2" sx={{ color: 'rgba(255, 255, 255, 0.7)', mt: 2 }}>
          Connecting to table...
        </Typography>
      </Box>
    );
  }

  const players = tableState.players || [];
  const communityCards = tableState.community_cards || [];
  const pot = tableState.pot || 0;
  const currentTurn = tableState.current_turn;
  const actionDeadline = tableState.action_deadline;
  const bettingRound = tableState.betting_round || 'waiting';
  const currentBet = tableState.current_bet || 0;

  // Find current player
  const currentPlayer = players.find((p) => p && p.user_id === currentTurn);

  return (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column', gap: 1 }}>
      {/* Current Turn / Status Indicator */}
      {currentTurn && tableState.status === 'playing' && (
        <Box
          sx={{
            px: 2,
            py: 1,
            bgcolor: 'rgba(251, 191, 36, 0.1)',
            borderRadius: 1.5,
            border: '1px solid rgba(251, 191, 36, 0.3)',
          }}
        >
          <Stack direction="row" spacing={2} alignItems="center" justifyContent="center">
            <Typography variant="body2" fontWeight="bold" sx={{ color: '#fbbf24', fontSize: '13px' }}>
              {currentPlayer ? (
                <>
                  🎯 <strong>{currentPlayer.user_id.slice(0, 8)}</strong>'s Turn
                </>
              ) : (
                <>⚠️ Waiting for action...</>
              )}
            </Typography>
            {actionDeadline && (
              <Box sx={{
                px: 1.5,
                py: 0.5,
                borderRadius: 1,
                bgcolor: 'rgba(239, 68, 68, 0.15)',
                border: '1px solid rgba(239, 68, 68, 0.3)',
              }}>
                <Typography variant="caption" sx={{ color: '#ef4444', fontSize: '11px', fontWeight: 700 }}>
                  ⏱️ Timer Active
                </Typography>
              </Box>
            )}
          </Stack>
        </Box>
      )}

      {/* Waiting for Game */}
      {tableState.status === 'waiting' && (
        <Box
          sx={{
            px: 2,
            py: 1,
            bgcolor: 'rgba(99, 102, 241, 0.1)',
            borderRadius: 1.5,
            border: '1px solid rgba(99, 102, 241, 0.3)',
            textAlign: 'center',
          }}
        >
          <Stack direction="row" spacing={1.5} alignItems="center" justifyContent="center">
            <CircularProgress size={16} sx={{ color: '#6366f1' }} />
            <Typography variant="body2" fontWeight="bold" sx={{ color: '#6366f1', fontSize: '13px' }}>
              Waiting for game to start... ({players.length} player{players.length !== 1 ? 's' : ''} at table)
            </Typography>
          </Stack>
        </Box>
      )}

      {/* Hand Complete */}
      {tableState.status === 'handComplete' && (
        <Box
          sx={{
            px: 2,
            py: 1,
            bgcolor: 'rgba(16, 185, 129, 0.1)',
            borderRadius: 1.5,
            border: '1px solid rgba(16, 185, 129, 0.3)',
            textAlign: 'center',
          }}
        >
          <Stack direction="row" spacing={1.5} alignItems="center" justifyContent="center">
            <HourglassEmpty sx={{ color: '#10b981', fontSize: 18, animation: 'spin 2s linear infinite', '@keyframes spin': { '0%': { transform: 'rotate(0deg)' }, '100%': { transform: 'rotate(360deg)' } } }} />
            <Typography variant="body2" fontWeight="bold" sx={{ color: '#10b981', fontSize: '13px' }}>
              Hand complete! Starting next round...
            </Typography>
          </Stack>
        </Box>
      )}

      {/* Poker Table - Premium Design */}
      <Box
        sx={{
          flex: 1,
          background: 'linear-gradient(135deg, rgba(5, 46, 22, 0.4) 0%, rgba(6, 78, 59, 0.3) 100%)',
          border: '3px solid rgba(16, 185, 129, 0.2)',
          borderRadius: 3,
          p: 2,
          overflow: 'auto',
          display: 'flex',
          flexDirection: 'column',
          gap: 1.5,
          position: 'relative',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: 'radial-gradient(ellipse at center, rgba(16, 185, 129, 0.05) 0%, transparent 70%)',
            pointerEvents: 'none',
          },
        }}
      >
        {/* Center Section: Pot and Community Cards */}
        <Stack direction="row" spacing={2} justifyContent="center" alignItems="center" sx={{ position: 'relative', zIndex: 1 }}>
          {/* Pot */}
          <Box
            sx={{
              bgcolor: 'rgba(17, 24, 39, 0.8)',
              borderRadius: 2,
              px: 2.5,
              py: 1.5,
              minWidth: 120,
              textAlign: 'center',
              border: '1px solid rgba(251, 191, 36, 0.3)',
              backdropFilter: 'blur(10px)',
            }}
          >
            <Typography variant="h5" sx={{ color: '#fbbf24', fontWeight: 900, fontSize: '24px' }}>
              💰 ${pot}
            </Typography>
            <Typography variant="caption" sx={{ color: 'rgba(156, 163, 175, 0.7)', fontSize: '10px', fontWeight: 600 }}>
              POT
            </Typography>
          </Box>

          {/* Community Cards */}
          {communityCards.length > 0 && (
            <Stack
              direction="row"
              spacing={0.75}
              sx={{
                bgcolor: 'rgba(17, 24, 39, 0.8)',
                px: 1.5,
                py: 1.5,
                borderRadius: 2,
                border: '1px solid rgba(99, 102, 241, 0.3)',
                backdropFilter: 'blur(10px)',
              }}
            >
              {communityCards.map((card, idx) => (
                <PlayingCard key={idx} card={card} size="medium" />
              ))}
            </Stack>
          )}

          {/* Betting Round Info */}
          {tableState.status === 'playing' && (
            <Box
              sx={{
                bgcolor: 'rgba(17, 24, 39, 0.8)',
                px: 2.5,
                py: 1.5,
                borderRadius: 2,
                textAlign: 'center',
                border: '1px solid rgba(99, 102, 241, 0.3)',
                backdropFilter: 'blur(10px)',
              }}
            >
              <Typography variant="body1" fontWeight="bold" sx={{ color: '#6366f1', fontSize: '14px' }}>
                {bettingRound.toUpperCase()}
              </Typography>
              <Typography variant="caption" sx={{ color: 'rgba(156, 163, 175, 0.7)', fontSize: '10px' }}>
                Current Bet: ${currentBet}
              </Typography>
            </Box>
          )}
        </Stack>

        {/* Players - Horizontal Compact Layout */}
        <Box
          sx={{
            display: 'flex',
            gap: 1.5,
            justifyContent: 'center',
            flexWrap: 'wrap',
            alignItems: 'stretch',
            position: 'relative',
            zIndex: 1,
          }}
        >
          {players.map((player, idx) => (
            <PlayerSeat
              key={player ? player.user_id : idx}
              player={player}
              position={player ? player.seat_number : idx}
              isActive={currentTurn === player?.user_id}
              actionDeadline={
                currentTurn === player?.user_id ? actionDeadline : undefined
              }
            />
          ))}
        </Box>
      </Box>
    </Box>
  );
};

export default PokerTable;
