import React, { useEffect, useState } from 'react';
import { Box, Container, Typography, Stack, Dialog, DialogContent, LinearProgress, Tabs, Tab, IconButton } from '@mui/material';
import { PlayArrow, Group, EmojiEvents, History, Close, AccessTime } from '@mui/icons-material';
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
import { COLORS, ROUTES } from '../constants';
import { formatTimestamp } from '../utils';

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

const GameModeCard: React.FC<GameModeCardProps> = ({
  title,
  description,
  blinds,
  buyIn,
  maxPlayers,
  icon,
  color,
  onJoin,
  disabled,
}) => {
  return (
    <Card
      variant="glass"
      sx={{
        position: 'relative',
        overflow: 'hidden',
        transition: 'all 300ms ease-in-out',
        '&:hover': {
          transform: 'translateY(-4px)',
          boxShadow: `0 12px 32px ${color}40`,
          borderColor: color,
        },
      }}
    >
      <Box
        sx={{
          position: 'absolute',
          top: -50,
          right: -50,
          width: 150,
          height: 150,
          borderRadius: '50%',
          background: `radial-gradient(circle, ${color}20 0%, transparent 70%)`,
          pointerEvents: 'none',
        }}
      />
      <Stack spacing={2.5} sx={{ position: 'relative', zIndex: 1 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <Box
            sx={{
              width: 48,
              height: 48,
              borderRadius: '12px',
              background: `linear-gradient(135deg, ${color} 0%, ${color}cc 100%)`,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              boxShadow: `0 4px 12px ${color}40`,
              color: COLORS.text.primary,
              fontSize: '1.5rem',
            }}
          >
            {icon}
          </Box>
          <Box>
            <Typography variant="h6" sx={{ fontWeight: 700 }}>
              {title}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              {description}
            </Typography>
          </Box>
        </Box>
        <Box sx={{ padding: 2, background: COLORS.background.tertiary, borderRadius: '8px' }}>
          <Stack spacing={1}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
              <Typography variant="body2" color="text.secondary">
                Players:
              </Typography>
              <Typography variant="body2" fontWeight={600}>
                {maxPlayers}
              </Typography>
            </Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
              <Typography variant="body2" color="text.secondary">
                Blinds:
              </Typography>
              <Typography variant="body2" fontWeight={600}>
                {blinds}
              </Typography>
            </Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
              <Typography variant="body2" color="text.secondary">
                Buy-in:
              </Typography>
              <Typography variant="body2" fontWeight={600}>
                {buyIn}
              </Typography>
            </Box>
          </Stack>
        </Box>
        <Button variant="primary" size="large" fullWidth onClick={onJoin} disabled={disabled}>
          Find Match
        </Button>
      </Stack>
    </Card>
  );
};

export const Lobby: React.FC = () => {
  const navigate = useNavigate();
  const { isConnected, lastMessage, addMessageHandler, removeMessageHandler } = useWebSocket();
  const { user } = useAuth();
  const { showError, showSuccess } = useToast();
  const [activeTab, setActiveTab] = useState(0);
  const [activeTables, setActiveTables] = useState<Table[]>([]);
  const [pastTables, setPastTables] = useState<Table[]>([]);
  const [loading, setLoading] = useState(false);
  const [initialLoad, setInitialLoad] = useState(true);
  const [matchmaking, setMatchmaking] = useState<{
    active: boolean;
    gameMode: string;
    queueSize: number;
    required: number;
  } | null>(null);

  const loadActiveTables = async () => {
    try {
      const response = await tableAPI.getActiveTables();
      setActiveTables(response.data || []);
    } catch (error) {
      console.error('Failed to load active tables:', error);
    } finally {
      setInitialLoad(false);
    }
  };

  const loadPastTables = async () => {
    try {
      const response = await tableAPI.getPastTables();
      setPastTables(response.data || []);
    } catch (error) {
      console.error('Failed to load past tables:', error);
    }
  };

  useEffect(() => {
    loadActiveTables();
    loadPastTables();
  }, []);

  // Listen for match_found WebSocket event
  useEffect(() => {
    const handler = (message: any) => {
      const { table_id } = message.payload;
      setMatchmaking(null);
      showSuccess('Match found! Joining table...');
      navigate(`/game/${table_id}`);
    };

    addMessageHandler('match_found', handler);
    return () => removeMessageHandler('match_found');
  }, [addMessageHandler, removeMessageHandler, navigate, showSuccess]);

  const handleJoinTable = async (tableId: string, minBuyIn: number) => {
    try {
      setLoading(true);
      await tableAPI.joinTable(tableId, minBuyIn);
      showSuccess('Joined table successfully!');
      navigate(`/game/${tableId}`);
    } catch (error: any) {
      showError(error.response?.data?.error || 'Failed to join table');
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
      showSuccess('Joined matchmaking queue!');
    } catch (error: any) {
      showError(error.response?.data?.error || 'Failed to join matchmaking');
    } finally {
      setLoading(false);
    }
  };

  const handleCancelMatchmaking = async () => {
    try {
      await matchmakingAPI.leave();
      setMatchmaking(null);
      showSuccess('Left matchmaking queue');
    } catch (error) {
      console.error('Failed to leave matchmaking:', error);
    }
  };

  if (initialLoad) {
    return (
      <AppLayout>
        <LoadingSpinner fullScreen message="Loading lobby..." />
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <Box sx={{ flexGrow: 1, pb: 6 }}>
        {/* Hero Section */}
        <Box
          sx={{
            background: `linear-gradient(135deg, ${COLORS.primary.dark} 0%, ${COLORS.background.primary} 100%)`,
            borderBottom: `1px solid ${COLORS.border.main}`,
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          <Box
            sx={{
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              opacity: 0.1,
              background: `radial-gradient(circle at 30% 50%, ${COLORS.secondary.main} 0%, transparent 50%)`,
            }}
          />
          <Container maxWidth="lg" sx={{ py: 6, position: 'relative', zIndex: 1 }}>
            <Stack spacing={2} alignItems="center" textAlign="center">
              <Typography
                variant="h3"
                sx={{
                  fontWeight: 700,
                  background: `linear-gradient(135deg, ${COLORS.primary.light} 0%, ${COLORS.secondary.light} 100%)`,
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                }}
              >
                Welcome back, {user?.username}
              </Typography>
              <Typography variant="h6" color="text.secondary" sx={{ maxWidth: 600 }}>
                Choose your game mode and start playing Texas Hold'em poker
              </Typography>
            </Stack>
          </Container>
        </Box>

        {/* Game Modes Section */}
        <Container maxWidth="lg" sx={{ mt: -4 }}>
          <Card variant="elevated" sx={{ mb: 6, p: 4 }}>
            <Typography variant="h5" sx={{ mb: 3, fontWeight: 700 }}>
              <PlayArrow sx={{ mr: 1, verticalAlign: 'middle' }} />
              Quick Match
            </Typography>
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 3 }}>
              <Box>
                <GameModeCard
                  title="Heads-Up"
                  description="1 vs 1 intense poker action"
                  blinds="$5 / $10"
                  buyIn="$100 - $1,000"
                  maxPlayers={2}
                  icon={<Group />}
                  color={COLORS.primary.main}
                  onJoin={() => handleQuickMatch('headsup')}
                  disabled={loading || matchmaking !== null}
                />
              </Box>
              <Box>
                <GameModeCard
                  title="3-Player"
                  description="Three-way poker showdown"
                  blinds="$10 / $20"
                  buyIn="$200 - $2,000"
                  maxPlayers={3}
                  icon={<Group />}
                  color={COLORS.secondary.main}
                  onJoin={() => handleQuickMatch('3player')}
                  disabled={loading || matchmaking !== null}
                />
              </Box>
            </Box>
          </Card>

          {/* Tables Section */}
          <Card variant="glass" sx={{ p: 0, overflow: 'hidden' }}>
            <Box sx={{ borderBottom: 1, borderColor: 'divider', px: 3, pt: 2 }}>
              <Tabs value={activeTab} onChange={(_, v) => setActiveTab(v)}>
                <Tab label="Active Games" icon={<EmojiEvents />} iconPosition="start" />
                <Tab label="Past Games" icon={<History />} iconPosition="start" />
              </Tabs>
            </Box>

            <Box sx={{ p: 3 }}>
              {activeTab === 0 && (
                <>
                  {activeTables.length > 0 ? (
                    <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)', lg: 'repeat(3, 1fr)' }, gap: 3 }}>
                      {activeTables.map((table) => (
                        <Box key={table.id}>
                          <Card
                            variant="default"
                            sx={{
                              height: '100%',
                              border: table.is_playing ? `2px solid ${COLORS.primary.main}` : undefined,
                              position: 'relative',
                            }}
                          >
                            {table.is_playing && (
                              <Box
                                sx={{
                                  position: 'absolute',
                                  top: -1,
                                  right: -1,
                                  px: 2,
                                  py: 0.5,
                                  background: `linear-gradient(135deg, ${COLORS.primary.main} 0%, ${COLORS.primary.dark} 100%)`,
                                  borderBottomLeftRadius: '8px',
                                }}
                              >
                                <Typography variant="caption" fontWeight={700}>
                                  YOU'RE IN
                                </Typography>
                              </Box>
                            )}
                            <Stack spacing={2}>
                              <Box>
                                <Typography variant="h6" fontWeight={700} sx={{ mb: 1 }}>
                                  {table.name}
                                </Typography>
                                <Stack direction="row" spacing={1} flexWrap="wrap">
                                  <Badge
                                    variant={table.status === 'playing' ? 'success' : 'warning'}
                                    pulse={table.status === 'playing'}
                                  >
                                    {table.status === 'playing' ? 'LIVE' : 'WAITING'}
                                  </Badge>
                                  <Badge variant="info" size="small">
                                    {table.current_players || 0}/{table.max_players}
                                  </Badge>
                                </Stack>
                              </Box>
                              <Box
                                sx={{
                                  p: 2,
                                  background: COLORS.background.secondary,
                                  borderRadius: '8px',
                                }}
                              >
                                <Stack spacing={1}>
                                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                                    <Typography variant="body2" color="text.secondary">
                                      Blinds:
                                    </Typography>
                                    <Typography variant="body2" fontWeight={600}>
                                      ${table.small_blind}/${table.big_blind}
                                    </Typography>
                                  </Box>
                                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                                    <Typography variant="body2" color="text.secondary">
                                      Buy-in:
                                    </Typography>
                                    <Typography variant="body2" fontWeight={600}>
                                      ${table.min_buy_in} - ${table.max_buy_in}
                                    </Typography>
                                  </Box>
                                  {table.pot && table.pot > 0 && (
                                    <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                                      <Typography variant="body2" color="text.secondary">
                                        Current Pot:
                                      </Typography>
                                      <Chip amount={table.pot} variant="pot" size="small" />
                                    </Box>
                                  )}
                                </Stack>
                              </Box>
                              <Button
                                variant={table.is_playing ? 'success' : 'primary'}
                                fullWidth
                                onClick={() =>
                                  table.is_playing
                                    ? navigate(`/game/${table.id}`)
                                    : handleJoinTable(table.id, table.min_buy_in)
                                }
                                disabled={
                                  loading || (!table.is_playing && (table.current_players || 0) >= table.max_players)
                                }
                              >
                                {table.is_playing
                                  ? 'Resume Game'
                                  : (table.current_players || 0) >= table.max_players
                                  ? 'Table Full'
                                  : 'Join Table'}
                              </Button>
                            </Stack>
                          </Card>
                        </Box>
                      ))}
                    </Box>
                  ) : (
                    <EmptyState
                      icon={<EmojiEvents />}
                      title="No active games"
                      description="Use Quick Match to start a new game or create a private table"
                    />
                  )}
                </>
              )}

              {activeTab === 1 && (
                <>
                  {pastTables.length > 0 ? (
                    <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)', lg: 'repeat(3, 1fr)' }, gap: 3 }}>
                      {pastTables.map((table) => (
                        <Box key={table.id}>
                          <Card variant="default" sx={{ height: '100%', opacity: 0.9 }}>
                            <Stack spacing={2}>
                              <Box>
                                <Typography variant="h6" fontWeight={700} sx={{ mb: 1 }}>
                                  {table.name}
                                </Typography>
                                <Stack direction="row" spacing={1}>
                                  <Badge variant="secondary" size="small">
                                    COMPLETED
                                  </Badge>
                                  {table.participated && (
                                    <Badge variant="primary" size="small">
                                      YOU PLAYED
                                    </Badge>
                                  )}
                                </Stack>
                              </Box>
                              <Box
                                sx={{
                                  p: 2,
                                  background: COLORS.background.secondary,
                                  borderRadius: '8px',
                                }}
                              >
                                <Stack spacing={1}>
                                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                                    <Typography variant="body2" color="text.secondary">
                                      Players:
                                    </Typography>
                                    <Typography variant="body2" fontWeight={600}>
                                      {table.total_players}
                                    </Typography>
                                  </Box>
                                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                                    <Typography variant="body2" color="text.secondary">
                                      Hands Played:
                                    </Typography>
                                    <Typography variant="body2" fontWeight={600}>
                                      {table.total_hands || 0}
                                    </Typography>
                                  </Box>
                                  {table.completed_at && (
                                    <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                                      <Typography variant="body2" color="text.secondary">
                                        Completed:
                                      </Typography>
                                      <Typography variant="body2" fontWeight={600}>
                                        {formatTimestamp(table.completed_at)}
                                      </Typography>
                                    </Box>
                                  )}
                                </Stack>
                              </Box>
                            </Stack>
                          </Card>
                        </Box>
                      ))}
                    </Box>
                  ) : (
                    <EmptyState
                      icon={<History />}
                      title="No past games"
                      description="Your completed games will appear here"
                    />
                  )}
                </>
              )}
            </Box>
          </Card>
        </Container>
      </Box>

      {/* Matchmaking Dialog */}
      <Dialog
        open={matchmaking !== null}
        maxWidth="sm"
        fullWidth
        PaperProps={{
          sx: {
            borderRadius: '16px',
            background: COLORS.background.paper,
          },
        }}
      >
        <DialogContent sx={{ p: 4 }}>
          <Stack spacing={4}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Typography variant="h5" fontWeight={700}>
                Finding Match
              </Typography>
              <IconButton onClick={handleCancelMatchmaking} size="small">
                <Close />
              </IconButton>
            </Box>

            <Box
              sx={{
                textAlign: 'center',
                py: 3,
                '@keyframes spin': {
                  '0%': { transform: 'rotate(0deg)' },
                  '100%': { transform: 'rotate(360deg)' },
                },
              }}
            >
              <Box
                sx={{
                  width: 80,
                  height: 80,
                  margin: '0 auto',
                  borderRadius: '50%',
                  border: `4px solid ${COLORS.border.main}`,
                  borderTopColor: COLORS.primary.main,
                  animation: 'spin 1s linear infinite',
                }}
              />
            </Box>

            <Box>
              <Typography variant="body1" color="text.secondary" align="center" sx={{ mb: 2 }}>
                Waiting for players to join...
              </Typography>
              <Typography variant="h4" align="center" fontWeight={700} color="primary" sx={{ mb: 3 }}>
                {matchmaking?.queueSize} / {matchmaking?.required}
              </Typography>
              <LinearProgress
                variant="determinate"
                value={((matchmaking?.queueSize || 0) / (matchmaking?.required || 1)) * 100}
                sx={{
                  height: 8,
                  borderRadius: 4,
                  backgroundColor: COLORS.background.tertiary,
                  '& .MuiLinearProgress-bar': {
                    background: `linear-gradient(90deg, ${COLORS.primary.main} 0%, ${COLORS.secondary.main} 100%)`,
                  },
                }}
              />
            </Box>

            {!isConnected && (
              <Box
                sx={{
                  p: 2,
                  background: `${COLORS.warning.main}20`,
                  border: `1px solid ${COLORS.warning.main}`,
                  borderRadius: '8px',
                }}
              >
                <Typography variant="body2" color="warning.main" align="center">
                  Connection lost. Reconnecting...
                </Typography>
              </Box>
            )}

            <Button variant="danger" fullWidth onClick={handleCancelMatchmaking}>
              Cancel Matchmaking
            </Button>
          </Stack>
        </DialogContent>
      </Dialog>
    </AppLayout>
  );
};
