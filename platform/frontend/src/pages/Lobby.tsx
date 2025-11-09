import React, { useEffect, useState } from 'react';
import {
  Box,
  Container,
  Card,
  CardContent,
  Button,
  Typography,
  Chip,
  Stack,
  AppBar,
  Toolbar,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  CircularProgress,
  LinearProgress,
  Alert,
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { tableAPI, matchmakingAPI } from '../services/api';
import { useWebSocket } from '../hooks/useWebSocket';

interface Table {
  id: string;
  name: string;
  game_type: string;
  status: string;
  small_blind: number;
  big_blind: number;
  max_players: number;
  current_players: number;
  min_buy_in: number;
  max_buy_in: number;
}

export const Lobby: React.FC = () => {
  const navigate = useNavigate();
  const { isConnected, lastMessage } = useWebSocket();
  const [tables, setTables] = useState<Table[]>([]);
  const [loading, setLoading] = useState(false);
  const [matchmaking, setMatchmaking] = useState<{
    active: boolean;
    gameMode: string;
    queueSize: number;
    required: number;
  } | null>(null);

  const loadTables = async () => {
    try {
      const response = await tableAPI.getTables();
      setTables(response.data);
    } catch (error) {
      console.error('Failed to load tables:', error);
    }
  };

  useEffect(() => {
    loadTables();
    const interval = setInterval(loadTables, 5000);
    return () => clearInterval(interval);
  }, []);

  // Listen for match_found WebSocket event
  useEffect(() => {
    if (lastMessage && lastMessage.type === 'match_found') {
      const { table_id } = lastMessage.payload;
      setMatchmaking(null);
      navigate(`/game/${table_id}`);
    }
  }, [lastMessage, navigate]);

  const handleJoinTable = async (tableId: string, minBuyIn: number) => {
    try {
      setLoading(true);
      await tableAPI.joinTable(tableId, minBuyIn);
      navigate(`/game/${tableId}`);
    } catch (error) {
      console.error('Failed to join table:', error);
      alert('Failed to join table');
    } finally {
      setLoading(false);
    }
  };

  const handleQuickMatch = async (gameMode: string) => {
    try {
      setLoading(true);
      const response = await matchmakingAPI.join(gameMode);
      const { queue_size, required } = response.data;
      setMatchmaking({
        active: true,
        gameMode,
        queueSize: queue_size,
        required,
      });
    } catch (error: any) {
      console.error('Matchmaking failed:', error);
      alert(error.response?.data?.error || 'Matchmaking failed');
    } finally {
      setLoading(false);
    }
  };

  const handleCancelMatchmaking = async () => {
    try {
      await matchmakingAPI.leave();
      setMatchmaking(null);
    } catch (error) {
      console.error('Failed to leave matchmaking:', error);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login');
  };

  return (
    <Box sx={{ flexGrow: 1 }}>
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            Poker Lobby
          </Typography>
          <Button color="inherit" onClick={handleLogout}>
            Logout
          </Button>
        </Toolbar>
      </AppBar>

      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Stack spacing={3}>
          <Box>
            <Typography variant="h5" gutterBottom sx={{ mb: 2 }}>
              Quick Match
            </Typography>
            <Stack direction="row" spacing={2} justifyContent="center">
              <Card sx={{ minWidth: 250 }}>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Heads-Up
                  </Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    1 vs 1 poker<br />
                    Blinds: $5/$10<br />
                    Buy-in: $100-$1000
                  </Typography>
                  <Button
                    variant="contained"
                    color="primary"
                    fullWidth
                    onClick={() => handleQuickMatch('headsup')}
                    disabled={loading || matchmaking !== null}
                  >
                    Find Match
                  </Button>
                </CardContent>
              </Card>

              <Card sx={{ minWidth: 250 }}>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    3-Player
                  </Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    3-way poker<br />
                    Blinds: $10/$20<br />
                    Buy-in: $200-$2000
                  </Typography>
                  <Button
                    variant="contained"
                    color="secondary"
                    fullWidth
                    onClick={() => handleQuickMatch('3player')}
                    disabled={loading || matchmaking !== null}
                  >
                    Find Match
                  </Button>
                </CardContent>
              </Card>
            </Stack>
          </Box>

          <Typography variant="h5" gutterBottom>
            Active Tables
          </Typography>

          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: {
                xs: '1fr',
                sm: 'repeat(2, 1fr)',
                md: 'repeat(3, 1fr)',
              },
              gap: 3,
            }}
          >
            {tables.filter(t => t.status === 'playing' || t.current_players > 0).map((table) => (
              <Card
                key={table.id}
                sx={{
                  height: '100%',
                  border: table.status === 'playing' ? '2px solid' : 'none',
                  borderColor: 'success.main',
                }}
              >
                <CardContent>
                  <Stack spacing={2}>
                    <Box>
                      <Typography variant="h6" gutterBottom>
                        {table.name}
                      </Typography>
                      <Stack direction="row" spacing={1}>
                        <Chip
                          label={table.status.toUpperCase()}
                          size="small"
                          color={table.status === 'playing' ? 'success' : 'warning'}
                        />
                        {table.status === 'playing' && (
                          <Chip
                            label="LIVE"
                            size="small"
                            color="error"
                            sx={{ animation: 'pulse 2s infinite' }}
                          />
                        )}
                      </Stack>
                    </Box>

                    <Box>
                      <Typography variant="body2" color="text.secondary">
                        Blinds: ${table.small_blind}/${table.big_blind}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        Players: {table.current_players}/{table.max_players}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        Buy-in: ${table.min_buy_in} - ${table.max_buy_in}
                      </Typography>
                    </Box>

                    <Button
                      variant="contained"
                      fullWidth
                      onClick={() => handleJoinTable(table.id, table.min_buy_in)}
                      disabled={loading || table.current_players >= table.max_players}
                    >
                      {table.current_players >= table.max_players ? 'Full' : 'Join Table'}
                    </Button>
                  </Stack>
                </CardContent>
              </Card>
            ))}
          </Box>

          {tables.length === 0 && (
            <Box sx={{ textAlign: 'center', py: 4 }}>
              <Typography color="text.secondary">
                No tables available. Try Quick Match!
              </Typography>
            </Box>
          )}
        </Stack>
      </Container>

      {/* Matchmaking Dialog */}
      <Dialog open={matchmaking !== null} maxWidth="sm" fullWidth>
        <DialogTitle>
          Finding {matchmaking?.gameMode === 'headsup' ? 'Heads-Up' : '3-Player'} Match
        </DialogTitle>
        <DialogContent>
          <Stack spacing={3} sx={{ py: 2 }}>
            <Box sx={{ textAlign: 'center' }}>
              <CircularProgress size={60} />
            </Box>

            <Box>
              <Typography variant="body1" gutterBottom align="center">
                Waiting for players...
              </Typography>
              <Typography variant="h6" align="center" color="primary" sx={{ my: 2 }}>
                {matchmaking?.queueSize} / {matchmaking?.required} players
              </Typography>
              <LinearProgress
                variant="determinate"
                value={((matchmaking?.queueSize || 0) / (matchmaking?.required || 1)) * 100}
                sx={{ height: 8, borderRadius: 4 }}
              />
            </Box>

            {!isConnected && (
              <Alert severity="warning">
                WebSocket disconnected. Reconnecting...
              </Alert>
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCancelMatchmaking} color="error">
            Cancel
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};
