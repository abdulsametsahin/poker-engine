import React from 'react';
import { Box, Paper, Typography, Stack, Chip, Avatar, Badge } from '@mui/material';
import { AccountCircle, AttachMoney } from '@mui/icons-material';
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
          minWidth: 160,
          height: 180,
          borderRadius: 2,
          border: '2px dashed',
          borderColor: 'grey.300',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          bgcolor: 'grey.50',
          opacity: 0.3,
        }}
      >
        <Typography variant="caption" color="text.secondary" align="center">
          Seat {position}
          <br />
          Empty
        </Typography>
      </Box>
    );
  }

  return (
    <Paper
      elevation={isActive ? 8 : 2}
      sx={{
        minWidth: 160,
        height: 180,
        p: 2,
        position: 'relative',
        border: isActive ? '4px solid' : '2px solid transparent',
        borderColor: isActive ? 'warning.main' : 'transparent',
        bgcolor: isActive ? 'warning.50' : 'background.paper',
        transition: 'all 0.3s',
        animation: isActive ? 'pulse 2s infinite' : 'none',
        '@keyframes pulse': {
          '0%, 100%': { boxShadow: '0 0 20px rgba(237, 108, 2, 0.5)' },
          '50%': { boxShadow: '0 0 40px rgba(237, 108, 2, 0.8)' },
        },
      }}
    >
      {/* Dealer Button */}
      {player.is_dealer && (
        <Chip
          label="D"
          size="small"
          color="error"
          sx={{
            position: 'absolute',
            top: -10,
            right: -10,
            width: 24,
            height: 24,
            fontSize: 12,
            fontWeight: 'bold',
          }}
        />
      )}

      <Stack spacing={1} alignItems="center" sx={{ height: '100%' }}>
        {/* Player Name & Avatar */}
        <Stack direction="row" spacing={1} alignItems="center" sx={{ width: '100%' }}>
          <Badge
            overlap="circular"
            anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
            badgeContent={
              player.status === 'folded' ? (
                <Chip
                  label="FOLD"
                  size="small"
                  color="error"
                  sx={{ height: 16, fontSize: 9 }}
                />
              ) : null
            }
          >
            <Avatar
              sx={{
                width: 40,
                height: 40,
                bgcolor: isActive ? 'warning.main' : 'primary.main',
                opacity: player.status === 'folded' ? 0.5 : 1,
              }}
            >
              <AccountCircle sx={{ fontSize: 32 }} />
            </Avatar>
          </Badge>
          <Box sx={{ flex: 1 }}>
            <Typography
              variant="subtitle2"
              fontWeight="bold"
              noWrap
              sx={{ opacity: player.status === 'folded' ? 0.5 : 1 }}
            >
              {player.user_id.slice(0, 8)}
            </Typography>
            <Typography variant="caption" color="text.secondary">
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

        {/* Chips */}
        <Chip
          icon={<AttachMoney />}
          label={`$${player.chips}`}
          size="small"
          color={player.chips > 500 ? 'success' : 'warning'}
          sx={{ fontWeight: 'bold', width: '100%' }}
        />

        {/* Current Bet */}
        {player.bet && player.bet > 0 && (
          <Chip
            label={`Bet: $${player.bet}`}
            size="small"
            variant="filled"
            color="error"
            sx={{ fontWeight: 'bold', width: '100%' }}
          />
        )}

        {/* Last Action */}
        {player.last_action && (
          <Chip
            label={player.last_action.toUpperCase()}
            size="small"
            color="info"
            sx={{ fontSize: 10, height: 20, width: '100%' }}
          />
        )}

        {/* Action Timer */}
        {isActive && actionDeadline && (
          <Box sx={{ width: '100%', mt: 'auto' }}>
            <ActionTimer deadline={actionDeadline} />
          </Box>
        )}
      </Stack>
    </Paper>
  );
};

const PokerTable: React.FC<PokerTableProps> = ({ tableState }) => {
  if (!tableState) {
    return (
      <Paper sx={{ p: 4, textAlign: 'center' }}>
        <Typography variant="h6" color="text.secondary">
          Connecting to table...
        </Typography>
      </Paper>
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
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* Table Header */}
      <Paper sx={{ p: 2, mb: 2 }}>
        <Stack spacing={1}>
          <Stack direction="row" justifyContent="space-between" alignItems="center">
            <Typography variant="h5" fontWeight="bold">
              {tableState.table_id || 'Poker Table'}
            </Typography>
            <Stack direction="row" spacing={1}>
              <Chip
                label={tableState.status || 'waiting'}
                color={tableState.status === 'playing' ? 'success' : 'default'}
                size="small"
              />
            </Stack>
          </Stack>

          {/* Current Turn Indicator */}
          {currentTurn && (
            <Box
              sx={{
                p: 1.5,
                bgcolor: currentPlayer ? 'warning.light' : 'error.light',
                borderRadius: 1,
              }}
            >
              <Stack
                direction="row"
                spacing={2}
                alignItems="center"
                justifyContent="space-between"
              >
                <Typography variant="body1" fontWeight="bold">
                  {currentPlayer ? (
                    <>
                      🎯 <strong>{currentPlayer.user_id.slice(0, 8)}</strong>'s Turn
                    </>
                  ) : (
                    <>⚠️ Waiting for action...</>
                  )}
                </Typography>
                {actionDeadline && (
                  <Chip
                    label="⏱️ Timer Active"
                    size="small"
                    color="warning"
                    sx={{ fontWeight: 'bold' }}
                  />
                )}
              </Stack>
            </Box>
          )}
        </Stack>
      </Paper>

      {/* Poker Table - Horizontal Layout */}
      <Paper
        sx={{
          flex: 1,
          background: 'linear-gradient(135deg, #0a5f0a 0%, #0d7a0d 100%)',
          border: '8px solid #8b4513',
          borderRadius: 4,
          p: 3,
          overflow: 'auto',
          display: 'flex',
          flexDirection: 'column',
          gap: 2,
        }}
      >
        {/* Top Section: Pot and Community Cards */}
        <Stack direction="row" spacing={3} justifyContent="center" alignItems="center">
          {/* Pot */}
          <Box
            sx={{
              bgcolor: 'rgba(0,0,0,0.4)',
              borderRadius: 2,
              p: 2,
              minWidth: 150,
              textAlign: 'center',
            }}
          >
            <Typography variant="h4" color="warning.main" fontWeight="bold">
              💰 ${pot}
            </Typography>
            <Typography variant="caption" color="white">
              POT
            </Typography>
          </Box>

          {/* Community Cards */}
          {communityCards.length > 0 && (
            <Stack
              direction="row"
              spacing={1}
              sx={{
                bgcolor: 'rgba(0,0,0,0.4)',
                p: 2,
                borderRadius: 2,
              }}
            >
              {communityCards.map((card, idx) => (
                <PlayingCard key={idx} card={card} size="large" />
              ))}
            </Stack>
          )}

          {/* Betting Round Info */}
          {tableState.status === 'playing' && (
            <Box
              sx={{
                bgcolor: 'rgba(0,0,0,0.4)',
                color: 'white',
                px: 3,
                py: 2,
                borderRadius: 2,
                textAlign: 'center',
              }}
            >
              <Typography variant="h6" fontWeight="bold">
                {bettingRound.toUpperCase()}
              </Typography>
              <Typography variant="caption">Current Bet: ${currentBet}</Typography>
            </Box>
          )}
        </Stack>

        {/* Players - Horizontal Layout */}
        <Box
          sx={{
            display: 'flex',
            gap: 2,
            justifyContent: 'center',
            flexWrap: 'wrap',
            alignItems: 'stretch',
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
      </Paper>
    </Box>
  );
};

export default PokerTable;
