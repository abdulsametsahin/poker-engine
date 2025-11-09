import React, { useEffect, useState } from 'react';
import {
  Box,
  Card,
  CardContent,
  Button,
  Typography,
  Avatar,
  Chip,
  Stack,
  TextField,
} from '@mui/material';
import { useParams } from 'react-router-dom';
import { useWebSocket } from '../hooks/useWebSocket';

interface Player {
  user_id: string;
  seat_number: number;
  chips: number;
  status: string;
  bet?: number;
  cards?: string[];
}

export const GameView: React.FC = () => {
  const { tableId } = useParams<{ tableId: string }>();
  const { isConnected, lastMessage, send } = useWebSocket();
  const [players, setPlayers] = useState<Player[]>([]);
  const [communityCards, setCommunityCards] = useState<string[]>([]);
  const [pot, setPot] = useState(0);
  const [raiseAmount, setRaiseAmount] = useState(0);
  const [currentTurn, setCurrentTurn] = useState<string | null>(null);

  useEffect(() => {
    if (isConnected) {
      send({
        type: 'subscribe_table',
        payload: { table_id: tableId },
      });
    }
  }, [isConnected, tableId, send]);

  useEffect(() => {
    if (lastMessage) {
      switch (lastMessage.type) {
        case 'table_state':
          setPlayers(lastMessage.payload.players || []);
          setCommunityCards(lastMessage.payload.community_cards || []);
          setPot(lastMessage.payload.pot || 0);
          setCurrentTurn(lastMessage.payload.current_turn);
          break;
        case 'game_update':
          setPlayers(lastMessage.payload.players || []);
          setCommunityCards(lastMessage.payload.community_cards || []);
          setPot(lastMessage.payload.pot || 0);
          setCurrentTurn(lastMessage.payload.current_turn);
          break;
      }
    }
  }, [lastMessage]);

  const handleAction = (action: string, amount?: number) => {
    send({
      type: 'game_action',
      payload: { action, amount: amount || 0 },
    });
  };

  const getPlayerPosition = (seatNumber: number) => {
    const positions = [
      { top: '70%', left: '50%' },
      { top: '70%', left: '20%' },
      { top: '40%', left: '10%' },
      { top: '10%', left: '20%' },
      { top: '10%', left: '50%' },
      { top: '10%', left: '80%' },
      { top: '40%', left: '90%' },
      { top: '70%', left: '80%' },
    ];
    return positions[seatNumber] || positions[0];
  };

  return (
    <Box sx={{ height: '100vh', bgcolor: '#0D4715', position: 'relative', overflow: 'hidden' }}>
      <Box
        sx={{
          position: 'absolute',
          top: '50%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          width: '70%',
          height: '60%',
          borderRadius: '50%',
          bgcolor: '#116530',
          border: '8px solid #8B4513',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Box sx={{ textAlign: 'center' }}>
          <Typography variant="h4" sx={{ color: '#FFD700', mb: 2 }}>
            POT: ${pot}
          </Typography>
          
          {communityCards.length > 0 && (
            <Stack direction="row" spacing={1} justifyContent="center">
              {communityCards.map((card, idx) => (
                <Card
                  key={idx}
                  sx={{
                    width: 60,
                    height: 90,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    bgcolor: 'white',
                    fontWeight: 'bold',
                  }}
                >
                  {card}
                </Card>
              ))}
            </Stack>
          )}
        </Box>
      </Box>

      {players.map((player) => {
        const pos = getPlayerPosition(player.seat_number);
        const isCurrentTurn = currentTurn === player.user_id;
        
        return (
          <Box
            key={player.user_id}
            sx={{
              position: 'absolute',
              ...pos,
              transform: 'translate(-50%, -50%)',
            }}
          >
            <Card
              sx={{
                minWidth: 140,
                bgcolor: isCurrentTurn ? 'primary.dark' : 'background.paper',
                border: isCurrentTurn ? '3px solid #4CAF50' : 'none',
              }}
            >
              <CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 } }}>
                <Stack spacing={0.5} alignItems="center">
                  <Avatar sx={{ width: 40, height: 40, bgcolor: 'secondary.main' }}>
                    {player.user_id.slice(0, 2).toUpperCase()}
                  </Avatar>
                  <Typography variant="caption" noWrap sx={{ maxWidth: 120 }}>
                    {player.user_id.slice(0, 8)}
                  </Typography>
                  <Chip
                    label={`$${player.chips}`}
                    size="small"
                    color={player.chips > 0 ? 'success' : 'error'}
                  />
                  {player.bet && player.bet > 0 && (
                    <Chip label={`Bet: $${player.bet}`} size="small" color="warning" />
                  )}
                  <Chip
                    label={player.status}
                    size="small"
                    variant="outlined"
                  />
                </Stack>
              </CardContent>
            </Card>
          </Box>
        );
      })}

      <Box
        sx={{
          position: 'absolute',
          bottom: 20,
          left: '50%',
          transform: 'translateX(-50%)',
          bgcolor: 'background.paper',
          p: 2,
          borderRadius: 2,
          minWidth: 400,
        }}
      >
        <Stack spacing={2}>
          <Stack direction="row" spacing={1}>
            <Button
              variant="contained"
              color="error"
              onClick={() => handleAction('fold')}
              fullWidth
            >
              Fold
            </Button>
            <Button
              variant="contained"
              onClick={() => handleAction('check')}
              fullWidth
            >
              Check
            </Button>
            <Button
              variant="contained"
              color="primary"
              onClick={() => handleAction('call')}
              fullWidth
            >
              Call
            </Button>
          </Stack>
          
          <Stack direction="row" spacing={1}>
            <TextField
              type="number"
              value={raiseAmount}
              onChange={(e) => setRaiseAmount(Number(e.target.value))}
              size="small"
              label="Amount"
              sx={{ flex: 1 }}
            />
            <Button
              variant="contained"
              color="secondary"
              onClick={() => handleAction('raise', raiseAmount)}
            >
              Raise
            </Button>
            <Button
              variant="contained"
              color="warning"
              onClick={() => handleAction('allin')}
            >
              All-In
            </Button>
          </Stack>
        </Stack>
      </Box>

      <Box sx={{ position: 'absolute', top: 10, right: 10 }}>
        <Chip
          label={isConnected ? 'Connected' : 'Disconnected'}
          color={isConnected ? 'success' : 'error'}
        />
      </Box>
    </Box>
  );
};
