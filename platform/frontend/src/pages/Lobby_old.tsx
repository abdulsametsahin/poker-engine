import React, { useEffect, useState } from 'react';
import { Box, Container, Typography, Stack, Dialog, DialogContent, LinearProgress, Grid, Tabs, Tab } from '@mui/material';
import { PlayArrow, Group, EmojiEvents, History, Close } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { tableAPI, matchmakingAPI } from '../services/api';
import { useWebSocket } from '../contexts/WebSocketContext';
import { useAuth } from '../contexts/AuthContext';
import { useToast } from '../contexts/ToastContext';
import { AppLayout } from '../components/common/AppLayout';
import { Button } from '../components/common/Button';
import { Card } from '../components/common/Card';
import { Badge } from '../components/common/Badge';
import { Chip } from '../components/common/Chip';
import { EmptyState } from '../components/common/EmptyState';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { COLORS, GAME, ROUTES } from '../constants';
import { formatTimestamp, getGameModeName } from '../utils';
import { TableState } from '../types';

interface Table {
  id: string;
  name: string;
  game_type: string;
  game_mode?: string;
  status: string;
  small_blind: number;
  big_blind: number;
  max_players: number;
  current_players?: number;
  total_players?: number;
  min_buy_in: number;
  max_buy_in: number;
  is_playing?: boolean;
  participated?: boolean;
  completed_at?: string;
  created_at?: string;
  total_hands?: number;
  pot?: number;
  current_hand?: number;
}

interface GameModeCardProps {
  title: string;
  description: string;
  blinds: string;
  buyIn: string;
  maxPlayers: number;
  icon: React.ReactNode;
  color: string;
  onJoin: () => void;
  disabled: boolean;
}

export const Lobby: React.FC = () => {
  const navigate = useNavigate();
  const { isConnected, lastMessage } = useWebSocket();
  const [activeTab, setActiveTab] = useState(0);
  const [activeTables, setActiveTables] = useState<Table[]>([]);
  const [pastTables, setPastTables] = useState<Table[]>([]);
  const [loading, setLoading] = useState(false);
  const [matchmaking, setMatchmaking] = useState<{
    active: boolean;
    gameMode: string;
    queueSize: number;
    required: number;
  } | null>(null);

  const loadActiveTables = async () => {
    try {
      const response = await tableAPI.getActiveTables();
      setActiveTables(response.data);
    } catch (error) {
      console.error('Failed to load active tables:', error);
    }
  };

  const loadPastTables = async () => {
    try {
      const response = await tableAPI.getPastTables();
      setPastTables(response.data);
    } catch (error) {
      console.error('Failed to load past tables:', error);
    }
  };

  useEffect(() => {
    loadActiveTables();
    loadPastTables();
    const interval = setInterval(() => {
      loadActiveTables();
      if (activeTab === 1) {
        loadPastTables();
      }
    }, 5000);
    return () => clearInterval(interval);
  }, [activeTab]);

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

          <Box>
            <Tabs value={activeTab} onChange={(e, newValue) => setActiveTab(newValue)}>
              <Tab label="Active Games" />
              <Tab label="Past Games" />
            </Tabs>

            {activeTab === 0 && (
              <Box sx={{ mt: 3 }}>
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
                  {activeTables.map((table) => (
                    <Card
                      key={table.id}
                      sx={{
                        height: '100%',
                        border: table.is_playing ? '2px solid' : 'none',
                        borderColor: 'primary.main',
                      }}
                    >
                      <CardContent>
                        <Stack spacing={2}>
                          <Box>
                            <Typography variant="h6" gutterBottom>
                              {table.name}
                            </Typography>
                            <Stack direction="row" spacing={1} flexWrap="wrap">
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
                              {table.is_playing && (
                                <Chip
                                  label="YOU'RE IN"
                                  size="small"
                                  color="primary"
                                />
                              )}
                            </Stack>
                          </Box>

                          <Box>
                            <Typography variant="body2" color="text.secondary">
                              Blinds: ${table.small_blind}/${table.big_blind}
                            </Typography>
                            <Typography variant="body2" color="text.secondary">
                              Players: {table.current_players || 0}/{table.max_players}
                            </Typography>
                            <Typography variant="body2" color="text.secondary">
                              Buy-in: ${table.min_buy_in} - ${table.max_buy_in}
                            </Typography>
                          </Box>

                          <Button
                            variant="contained"
                            fullWidth
                            onClick={() => table.is_playing ? navigate(`/game/${table.id}`) : handleJoinTable(table.id, table.min_buy_in)}
                            disabled={loading || (!table.is_playing && (table.current_players || 0) >= table.max_players)}
                          >
                            {table.is_playing ? 'Resume Game' : (table.current_players || 0) >= table.max_players ? 'Full' : 'Join Table'}
                          </Button>
                        </Stack>
                      </CardContent>
                    </Card>
                  ))}
                </Box>

                {activeTables.length === 0 && (
                  <Box sx={{ textAlign: 'center', py: 4 }}>
                    <Typography color="text.secondary">
                      No active tables. Try Quick Match to create a new game!
                    </Typography>
                  </Box>
                )}
              </Box>
            )}

            {activeTab === 1 && (
              <Box sx={{ mt: 3 }}>
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
                  {pastTables.map((table) => (
                    <Card
                      key={table.id}
                      sx={{
                        height: '100%',
                        opacity: 0.9,
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
                                label="COMPLETED"
                                size="small"
                                color="default"
                              />
                              {table.participated && (
                                <Chip
                                  label="YOU PLAYED"
                                  size="small"
                                  color="info"
                                />
                              )}
                            </Stack>
                          </Box>

                          <Box>
                            <Typography variant="body2" color="text.secondary">
                              Blinds: ${table.small_blind}/${table.big_blind}
                            </Typography>
                            <Typography variant="body2" color="text.secondary">
                              Players: {table.total_players}
                            </Typography>
                            <Typography variant="body2" color="text.secondary">
                              Hands Played: {table.total_hands || 0}
                            </Typography>
                            {table.completed_at && (
                              <Typography variant="body2" color="text.secondary">
                                Completed: {new Date(table.completed_at).toLocaleString()}
                              </Typography>
                            )}
                          </Box>
                        </Stack>
                      </CardContent>
                    </Card>
                  ))}
                </Box>

                {pastTables.length === 0 && (
                  <Box sx={{ textAlign: 'center', py: 4 }}>
                    <Typography color="text.secondary">
                      No past games found.
                    </Typography>
                  </Box>
                )}
              </Box>
            )}
          </Box>
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
