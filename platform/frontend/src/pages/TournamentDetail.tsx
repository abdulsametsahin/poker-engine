import React, { useEffect, useState, useCallback } from 'react';
import {
  Box,
  Container,
  Typography,
  Stack,
  Chip,
  LinearProgress,
  IconButton,
  Divider,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Alert,
  AlertTitle,
} from '@mui/material';
import {
  ArrowBack,
  ContentCopy,
  PlayArrow,
  Pause,
  ExitToApp,
  EmojiEvents,
  AccessTime,
  People,
  TrendingUp,
  PersonAdd,
} from '@mui/icons-material';
import { useNavigate, useParams } from 'react-router-dom';
import { tournamentAPI } from '../services/api';
import { useAuth } from '../contexts/AuthContext';
import { useToast } from '../contexts/ToastContext';
import { useWebSocket } from '../contexts/WebSocketContext';
import { AppLayout } from '../components/common/AppLayout';
import { Button } from '../components/common/Button';
import { Card } from '../components/common/Card';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { COLORS } from '../constants';

interface Tournament {
  id: string;
  tournament_code: string;
  name: string;
  creator_id?: string;
  status: 'registering' | 'starting' | 'in_progress' | 'paused' | 'completed' | 'cancelled';
  buy_in: number;
  starting_chips: number;
  max_players: number;
  min_players: number;
  current_players: number;
  prize_pool: number;
  structure: string;
  prize_structure: string;
  start_time?: string;
  registration_closes_at?: string;
  registration_completed_at?: string;
  auto_start_delay: number;
  current_level: number;
  level_started_at?: string;
  started_at?: string;
  completed_at?: string;
  prizes_distributed: boolean;
  created_at: string;
}

interface TournamentPlayer {
  user_id: string;
  username: string;
  chips: number;
  status: string;
  registered_at: string;
}

interface BlindLevel {
  level: number;
  small_blind: number;
  big_blind: number;
  ante: number;
  duration: number;
}

interface TournamentStructure {
  name: string;
  description: string;
  blind_levels: BlindLevel[];
}

interface Standing {
  user_id: string;
  username?: string;
  position: number | null;
  prize_amount: number;
  eliminated_at?: string;
}

export const TournamentDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const { showSuccess, showError } = useToast();
  const navigate = useNavigate();
  const { addMessageHandler, removeMessageHandler } = useWebSocket();

  const [tournament, setTournament] = useState<Tournament | null>(null);
  const [players, setPlayers] = useState<TournamentPlayer[]>([]);
  const [standings, setStandings] = useState<Standing[]>([]);
  const [loading, setLoading] = useState(true);
  const [isRegistered, setIsRegistered] = useState(false);
  const [countdown, setCountdown] = useState<number | null>(null);
  const [blindCountdown, setBlindCountdown] = useState<number | null>(null);
  const [isPausing, setIsPausing] = useState(false);
  const [isResuming, setIsResuming] = useState(false);

  const fetchTournamentData = useCallback(async () => {
    if (!id) return;

    try {
      const [tournamentRes, playersRes] = await Promise.all([
        tournamentAPI.getTournament(id),
        tournamentAPI.getTournamentPlayers(id),
      ]);

      setTournament(tournamentRes.data);
      setPlayers(playersRes.data || []);

      // Check if current user is registered
      if (user && playersRes.data) {
        const registered = playersRes.data.some((p: TournamentPlayer) => p.user_id === user.id);
        setIsRegistered(registered);
      }

      // Fetch standings if tournament is in progress or completed
      if (tournamentRes.data.status === 'in_progress' || tournamentRes.data.status === 'completed') {
        try {
          const standingsRes = await tournamentAPI.getTournamentStandings(id);
          setStandings(standingsRes.data.standings || []);
        } catch (error) {
          console.error('Failed to fetch standings:', error);
        }
      }
    } catch (error: any) {
      showError(error.response?.data?.error || 'Failed to load tournament');
      navigate('/tournaments');
    } finally {
      setLoading(false);
    }
  }, [id, user, showError, navigate]);

  useEffect(() => {
    fetchTournamentData();
  }, [fetchTournamentData]);

  // Setup WebSocket handlers for real-time updates
  useEffect(() => {
    if (!id) return;

    const handleTournamentUpdate = (message: any) => {
      if (message.payload?.tournament?.id === id) {
        setTournament(message.payload.tournament);
        if (message.payload.players) {
          setPlayers(message.payload.players);
        }
      }
    };

    const handleTournamentStarted = (message: any) => {
      if (message.payload?.tournament_id === id) {
        fetchTournamentData();
        showSuccess('Tournament has started!');
      }
    };

    const handleTournamentPaused = (message: { payload: { tournament_id: string; tournament?: Tournament } }) => {
      if (message.payload?.tournament_id === id) {
        // Update tournament status to paused
        if (message.payload.tournament) {
          setTournament(message.payload.tournament);
        } else {
          setTournament(prev => prev ? { ...prev, status: 'paused' } : null);
        }
        showSuccess('Tournament has been paused');
      }
    };

    const handleTournamentResumed = (message: { payload: { tournament_id: string; tournament?: Tournament } }) => {
      if (message.payload?.tournament_id === id) {
        // Update tournament status to in_progress
        if (message.payload.tournament) {
          setTournament(message.payload.tournament);
        } else {
          setTournament(prev => prev ? { ...prev, status: 'in_progress' } : null);
        }
        showSuccess('Tournament has been resumed');
      }
    };

    const handleBlindIncrease = (message: { payload: { tournament_id: string; current_level: number } }) => {
      if (message.payload?.tournament_id === id) {
        setTournament(prev => prev ? {
          ...prev,
          current_level: message.payload.current_level,
          level_started_at: new Date().toISOString(),
        } : null);
      }
    };

    const handlePlayerEliminated = (message: { payload: {
      tournament_id: string;
      player_id: string;
      player_name: string;
      position: number;
      eliminated_by?: string;
    } }) => {
      if (message.payload?.tournament_id !== id) return;

      const { player_id, player_name, position } = message.payload;

      // Update players list to mark as eliminated
      setPlayers(prev => prev.map(player =>
        player.user_id === player_id
          ? { ...player, status: 'eliminated' as any }
          : player
      ));

      // Update standings if we have them
      setStandings(prev => {
        // Check if this position exists
        const existingIndex = prev.findIndex(s => s.position === position);

        const newStanding = {
          position,
          player_name,
          chips: 0,
          status: 'eliminated',
        };

        if (existingIndex >= 0) {
          // Update existing
          const updated = [...prev];
          updated[existingIndex] = newStanding;
          return updated;
        } else {
          // Add new and sort
          return [...prev, newStanding].sort((a, b) => a.position - b.position);
        }
      });

      console.log(`[TournamentDetail] Player ${player_name} eliminated in position ${position}`);
    };

    const handleTournamentComplete = (message: { payload: {
      tournament_id: string;
      winner_id: string;
      winner_name: string;
      final_standings?: Array<{
        player_id: string;
        player_name: string;
        position: number;
        prize?: number;
      }>;
    } }) => {
      if (message.payload?.tournament_id !== id) return;

      // Update tournament status
      setTournament(prev => prev ? { ...prev, status: 'completed' } : null);

      // Update standings if provided
      if (message.payload.final_standings) {
        setStandings(message.payload.final_standings.map(s => ({
          position: s.position,
          player_name: s.player_name,
          chips: 0,
          status: s.position === 1 ? 'winner' : 'eliminated',
        })));
      }

      showSuccess(`Tournament complete! Winner: ${message.payload.winner_name}`);
      console.log('[TournamentDetail] Tournament completed');
    };

    // Register handlers and store cleanup functions
    const cleanup1 = addMessageHandler('tournament_update', handleTournamentUpdate);
    const cleanup2 = addMessageHandler('tournament_started', handleTournamentStarted);
    const cleanup3 = addMessageHandler('tournament_paused', handleTournamentPaused);
    const cleanup4 = addMessageHandler('tournament_resumed', handleTournamentResumed);
    const cleanup5 = addMessageHandler('blind_level_increased', handleBlindIncrease);
    const cleanup6 = addMessageHandler('player_eliminated', handlePlayerEliminated);
    const cleanup7 = addMessageHandler('tournament_complete', handleTournamentComplete);

    return () => {
      cleanup1();
      cleanup2();
      cleanup3();
      cleanup4();
      cleanup5();
      cleanup6();
      cleanup7();
    };
  }, [id, addMessageHandler, removeMessageHandler, fetchTournamentData, showSuccess]);

  // Countdown timer for tournament start
  useEffect(() => {
    if (!tournament || tournament.status !== 'registering') {
      setCountdown(null);
      return;
    }

    const calculateCountdown = () => {
      if (tournament.start_time) {
        // Scheduled start time
        const startTime = new Date(tournament.start_time).getTime();
        const now = Date.now();
        const diff = Math.max(0, Math.floor((startTime - now) / 1000));
        setCountdown(diff);
      } else if (tournament.registration_completed_at) {
        // Auto-start countdown based on when min_players was reached
        const registrationCompletedTime = new Date(tournament.registration_completed_at).getTime();
        const now = Date.now();
        const elapsed = Math.floor((now - registrationCompletedTime) / 1000);
        const remaining = Math.max(0, tournament.auto_start_delay - elapsed);
        setCountdown(remaining);
      } else {
        // Min players not reached yet
        setCountdown(null);
      }
    };

    calculateCountdown();
    const interval = setInterval(calculateCountdown, 1000);

    return () => clearInterval(interval);
  }, [tournament]);

  // Countdown timer for blind level
  useEffect(() => {
    if (!tournament || tournament.status !== 'in_progress' || !tournament.level_started_at) {
      setBlindCountdown(null);
      return;
    }

    const calculateBlindCountdown = () => {
      try {
        const structure: TournamentStructure = JSON.parse(tournament.structure);
        const currentLevel = structure.blind_levels[tournament.current_level - 1];

        if (currentLevel) {
          const levelStartTime = new Date(tournament.level_started_at!).getTime();
          const now = Date.now();
          const elapsed = Math.floor((now - levelStartTime) / 1000);
          const remaining = Math.max(0, currentLevel.duration - elapsed);
          setBlindCountdown(remaining);
        }
      } catch (error) {
        console.error('Failed to parse tournament structure:', error);
      }
    };

    calculateBlindCountdown();
    const interval = setInterval(calculateBlindCountdown, 1000);

    return () => clearInterval(interval);
  }, [tournament]);

  const handleRegister = async () => {
    if (!id) return;

    try {
      await tournamentAPI.registerForTournament(id);
      showSuccess('Successfully registered for tournament!');
      fetchTournamentData();
    } catch (error: any) {
      showError(error.response?.data?.error || 'Failed to register');
    }
  };

  const handleUnregister = async () => {
    if (!id) return;

    try {
      await tournamentAPI.unregisterFromTournament(id);
      showSuccess('Successfully unregistered from tournament');
      fetchTournamentData();
    } catch (error: any) {
      showError(error.response?.data?.error || 'Failed to unregister');
    }
  };

  const handleStartTournament = async () => {
    if (!id) return;

    try {
      await tournamentAPI.startTournament(id);
      showSuccess('Tournament started!');
      fetchTournamentData();
    } catch (error: any) {
      showError(error.response?.data?.error || 'Failed to start tournament');
    }
  };

  const handlePauseTournament = async () => {
    if (!id) return;

    if (!window.confirm('Are you sure you want to pause this tournament? All games will be paused.')) {
      return;
    }

    setIsPausing(true);
    try {
      await tournamentAPI.pauseTournament(id);
      showSuccess('Tournament paused');
      fetchTournamentData();
    } catch (error: any) {
      showError(error.response?.data?.error || 'Failed to pause tournament');
    } finally {
      setIsPausing(false);
    }
  };

  const handleResumeTournament = async () => {
    if (!id) return;

    setIsResuming(true);
    try {
      await tournamentAPI.resumeTournament(id);
      showSuccess('Tournament resumed');
      fetchTournamentData();
    } catch (error: any) {
      showError(error.response?.data?.error || 'Failed to resume tournament');
    } finally {
      setIsResuming(false);
    }
  };

  const handleCopyCode = () => {
    if (tournament) {
      navigator.clipboard.writeText(tournament.tournament_code);
      showSuccess('Tournament code copied to clipboard!');
    }
  };

  const formatTime = (seconds: number | null): string => {
    if (seconds === null) return '--:--';
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'registering':
        return 'success';
      case 'starting':
        return 'warning';
      case 'in_progress':
        return 'info';
      case 'paused':
        return 'warning';
      case 'completed':
        return 'default';
      case 'cancelled':
        return 'error';
      default:
        return 'default';
    }
  };

  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'registering':
        return 'Open for Registration';
      case 'starting':
        return 'Starting Soon...';
      case 'in_progress':
        return 'In Progress';
      case 'paused':
        return 'Paused';
      case 'completed':
        return 'Completed';
      case 'cancelled':
        return 'Cancelled';
      default:
        return status;
    }
  };

  if (loading) {
    return (
      <AppLayout>
        <Container maxWidth="xl" sx={{ py: 4 }}>
          <LoadingSpinner />
        </Container>
      </AppLayout>
    );
  }

  if (!tournament) {
    return (
      <AppLayout>
        <Container maxWidth="xl" sx={{ py: 4 }}>
          <Typography>Tournament not found</Typography>
        </Container>
      </AppLayout>
    );
  }

  const structure: TournamentStructure = JSON.parse(tournament.structure);
  const currentLevel = structure.blind_levels[tournament.current_level - 1];
  const nextLevel = structure.blind_levels[tournament.current_level];
  const isCreator = user?.id === tournament.creator_id;
  const canStart = isCreator && tournament.status === 'registering' && tournament.current_players >= tournament.min_players;
  const canPause = isCreator && tournament.status === 'in_progress';
  const canResume = isCreator && tournament.status === 'paused';

  return (
    <AppLayout>
      <Container maxWidth="xl" sx={{ py: 4 }}>
        <Stack spacing={3}>
          {/* Header */}
          <Box display="flex" alignItems="center" justifyContent="space-between">
            <Box display="flex" alignItems="center" gap={2}>
              <IconButton onClick={() => navigate('/tournaments')} sx={{ color: COLORS.text.secondary }}>
                <ArrowBack />
              </IconButton>
              <Box>
                <Typography variant="h4" fontWeight="bold" color={COLORS.text.primary}>
                  {tournament.name}
                </Typography>
                <Box display="flex" alignItems="center" gap={1} mt={0.5}>
                  <Chip
                    label={getStatusLabel(tournament.status)}
                    color={getStatusColor(tournament.status) as any}
                    size="small"
                  />
                  <Typography variant="body2" color={COLORS.text.secondary}>
                    Code: {tournament.tournament_code}
                  </Typography>
                  <IconButton size="small" onClick={handleCopyCode}>
                    <ContentCopy fontSize="small" />
                  </IconButton>
                </Box>
              </Box>
            </Box>

            {/* Action Buttons */}
            <Stack direction="row" spacing={2}>
              {tournament.status === 'registering' && !isRegistered && (
                <Button
                  variant="primary"
                  startIcon={<PersonAdd />}
                  onClick={handleRegister}
                  disabled={tournament.current_players >= tournament.max_players}
                >
                  Register ({tournament.buy_in} chips)
                </Button>
              )}
              {tournament.status === 'registering' && isRegistered && (
                <Button
                  variant="ghost"
                  startIcon={<ExitToApp />}
                  onClick={handleUnregister}
                >
                  Unregister
                </Button>
              )}
              {canStart && (
                <Button
                  variant="success"
                  startIcon={<PlayArrow />}
                  onClick={handleStartTournament}
                >
                  Start Tournament Now
                </Button>
              )}
              {canPause && (
                <Button
                  variant="warning"
                  startIcon={<Pause />}
                  onClick={handlePauseTournament}
                  loading={isPausing}
                >
                  Pause Tournament
                </Button>
              )}
              {canResume && (
                <Button
                  variant="success"
                  startIcon={<PlayArrow />}
                  onClick={handleResumeTournament}
                  loading={isResuming}
                >
                  Resume Tournament
                </Button>
              )}
            </Stack>
          </Box>

          {/* Paused Banner */}
          {tournament.status === 'paused' && (
            <Alert severity="warning">
              <AlertTitle>Tournament Paused</AlertTitle>
              This tournament is currently paused. All games are on hold.
              {isCreator && ' Click "Resume Tournament" to continue.'}
            </Alert>
          )}

          {/* Tournament Info Cards */}
          <Box display="grid" gridTemplateColumns={{ xs: "1fr", md: "repeat(4, 1fr)" }} gap={3}>
            <Box>
              <Card>
                <Stack spacing={1} p={2}>
                  <Box display="flex" alignItems="center" gap={1}>
                    <People sx={{ color: COLORS.secondary.main }} />
                    <Typography variant="body2" color={COLORS.text.secondary}>
                      Players
                    </Typography>
                  </Box>
                  <Typography variant="h5" fontWeight="bold">
                    {tournament.current_players} / {tournament.max_players}
                  </Typography>
                  <LinearProgress
                    variant="determinate"
                    value={(tournament.current_players / tournament.max_players) * 100}
                    sx={{ mt: 1 }}
                  />
                  <Typography variant="caption" color={COLORS.text.secondary}>
                    Min: {tournament.min_players}
                  </Typography>
                </Stack>
              </Card>
            </Box>

            <Box>
              <Card>
                <Stack spacing={1} p={2}>
                  <Box display="flex" alignItems="center" gap={1}>
                    <EmojiEvents sx={{ color: COLORS.warning.main }} />
                    <Typography variant="body2" color={COLORS.text.secondary}>
                      Prize Pool
                    </Typography>
                  </Box>
                  <Typography variant="h5" fontWeight="bold">
                    {tournament.prize_pool.toLocaleString()} chips
                  </Typography>
                  <Typography variant="caption" color={COLORS.text.secondary}>
                    Buy-in: {tournament.buy_in} chips
                  </Typography>
                </Stack>
              </Card>
            </Box>

            {tournament.status === 'registering' && countdown !== null && (
              <Box>
                <Card>
                  <Stack spacing={1} p={2}>
                    <Box display="flex" alignItems="center" gap={1}>
                      <AccessTime sx={{ color: COLORS.info.main }} />
                      <Typography variant="body2" color={COLORS.text.secondary}>
                        Starts In
                      </Typography>
                    </Box>
                    <Typography variant="h5" fontWeight="bold" sx={{ fontFamily: 'monospace' }}>
                      {formatTime(countdown)}
                    </Typography>
                    <Typography variant="caption" color={COLORS.text.secondary}>
                      {tournament.start_time ? 'Scheduled start' : 'Auto-start delay'}
                    </Typography>
                  </Stack>
                </Card>
              </Box>
            )}

            {tournament.status === 'in_progress' && (
              <>
                <Box>
                  <Card>
                    <Stack spacing={1} p={2}>
                      <Box display="flex" alignItems="center" gap={1}>
                        <TrendingUp sx={{ color: COLORS.success.main }} />
                        <Typography variant="body2" color={COLORS.text.secondary}>
                          Current Blinds
                        </Typography>
                      </Box>
                      <Typography variant="h5" fontWeight="bold">
                        {currentLevel?.small_blind} / {currentLevel?.big_blind}
                      </Typography>
                      <Typography variant="caption" color={COLORS.text.secondary}>
                        Level {tournament.current_level}
                        {currentLevel?.ante > 0 && ` • Ante: ${currentLevel.ante}`}
                      </Typography>
                    </Stack>
                  </Card>
                </Box>

                <Box>
                  <Card>
                    <Stack spacing={1} p={2}>
                      <Box display="flex" alignItems="center" gap={1}>
                        <AccessTime sx={{ color: COLORS.warning.main }} />
                        <Typography variant="body2" color={COLORS.text.secondary}>
                          Next Level In
                        </Typography>
                      </Box>
                      <Typography variant="h5" fontWeight="bold" sx={{ fontFamily: 'monospace' }}>
                        {formatTime(blindCountdown)}
                      </Typography>
                      {nextLevel && (
                        <Typography variant="caption" color={COLORS.text.secondary}>
                          Next: {nextLevel.small_blind}/{nextLevel.big_blind}
                        </Typography>
                      )}
                    </Stack>
                  </Card>
                </Box>
              </>
            )}
          </Box>

          {/* Status-based Content */}
          {tournament.status === 'registering' && (
            <Card>
              <Box p={3}>
                <Typography variant="h6" fontWeight="bold" mb={2}>
                  Registered Players
                </Typography>
                {players.length === 0 ? (
                  <Typography color={COLORS.text.secondary}>
                    No players registered yet. Be the first!
                  </Typography>
                ) : (
                  <Box display="grid" gridTemplateColumns={{ xs: "1fr", sm: "repeat(2, 1fr)", md: "repeat(3, 1fr)" }} gap={2}>
                    {players.map((player) => (
                      <Box key={player.user_id}>
                        <Paper sx={{ p: 2, bgcolor: COLORS.background.paper }}>
                          <Typography fontWeight="bold">{player.username}</Typography>
                          <Typography variant="caption" color={COLORS.text.secondary}>
                            Registered {new Date(player.registered_at).toLocaleTimeString()}
                          </Typography>
                        </Paper>
                      </Box>
                    ))}
                  </Box>
                )}
              </Box>
            </Card>
          )}

          {tournament.status === 'starting' && (
            <Card>
              <Box p={3} textAlign="center">
                <LoadingSpinner />
                <Typography variant="h6" mt={2}>
                  Tournament is starting...
                </Typography>
                <Typography color={COLORS.text.secondary}>
                  Please wait while tables are being prepared
                </Typography>
              </Box>
            </Card>
          )}

          {(tournament.status === 'in_progress' || tournament.status === 'completed') && standings.length > 0 && (
            <Card>
              <Box p={3}>
                <Typography variant="h6" fontWeight="bold" mb={2}>
                  {tournament.status === 'completed' ? 'Final Standings' : 'Current Standings'}
                </Typography>
                <TableContainer>
                  <Table>
                    <TableHead>
                      <TableRow>
                        <TableCell>Position</TableCell>
                        <TableCell>Player</TableCell>
                        <TableCell align="right">Chips</TableCell>
                        {tournament.status === 'completed' && <TableCell align="right">Prize</TableCell>}
                        <TableCell>Status</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {standings
                        .sort((a, b) => (a.position || 999) - (b.position || 999))
                        .map((standing) => {
                          const player = players.find(p => p.user_id === standing.user_id);
                          return (
                            <TableRow key={standing.user_id}>
                              <TableCell>
                                {standing.position ? (
                                  <Chip
                                    label={`#${standing.position}`}
                                    color={standing.position === 1 ? 'warning' : 'default'}
                                    size="small"
                                  />
                                ) : (
                                  '-'
                                )}
                              </TableCell>
                              <TableCell>
                                {standing.username || player?.username || 'Unknown'}
                              </TableCell>
                              <TableCell align="right">
                                {player?.chips.toLocaleString() || '-'}
                              </TableCell>
                              {tournament.status === 'completed' && (
                                <TableCell align="right">
                                  {standing.prize_amount > 0 ? (
                                    <Typography color={COLORS.success.main} fontWeight="bold">
                                      +{standing.prize_amount.toLocaleString()}
                                    </Typography>
                                  ) : (
                                    '-'
                                  )}
                                </TableCell>
                              )}
                              <TableCell>
                                {standing.position ? (
                                  <Chip label="Eliminated" size="small" color="error" />
                                ) : (
                                  <Chip label="Active" size="small" color="success" />
                                )}
                              </TableCell>
                            </TableRow>
                          );
                        })}
                    </TableBody>
                  </Table>
                </TableContainer>
              </Box>
            </Card>
          )}

          {tournament.status === 'cancelled' && (
            <Card>
              <Box p={3} textAlign="center">
                <Typography variant="h6" color="error" mb={1}>
                  Tournament Cancelled
                </Typography>
                <Typography color={COLORS.text.secondary}>
                  This tournament has been cancelled. All buy-ins have been refunded.
                </Typography>
              </Box>
            </Card>
          )}

          {/* Tournament Details */}
          <Box display="grid" gridTemplateColumns={{ xs: "1fr", md: "repeat(2, 1fr)" }} gap={3}>
            <Box>
              <Card>
                <Box p={3}>
                  <Typography variant="h6" fontWeight="bold" mb={2}>
                    Blind Structure
                  </Typography>
                  <Typography variant="subtitle2" color={COLORS.text.secondary} mb={2}>
                    {structure.name}
                  </Typography>
                  <TableContainer>
                    <Table size="small">
                      <TableHead>
                        <TableRow>
                          <TableCell>Level</TableCell>
                          <TableCell>Small/Big</TableCell>
                          <TableCell>Ante</TableCell>
                          <TableCell>Duration</TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {structure.blind_levels.slice(0, 10).map((level) => (
                          <TableRow
                            key={level.level}
                            sx={{
                              bgcolor:
                                level.level === tournament.current_level
                                  ? 'rgba(25, 118, 210, 0.08)'
                                  : 'transparent',
                            }}
                          >
                            <TableCell>{level.level}</TableCell>
                            <TableCell>{level.small_blind}/{level.big_blind}</TableCell>
                            <TableCell>{level.ante || '-'}</TableCell>
                            <TableCell>{level.duration / 60} min</TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </TableContainer>
                </Box>
              </Card>
            </Box>

            <Box>
              <Card>
                <Box p={3}>
                  <Typography variant="h6" fontWeight="bold" mb={2}>
                    Tournament Info
                  </Typography>
                  <Stack spacing={2}>
                    <Box>
                      <Typography variant="body2" color={COLORS.text.secondary}>
                        Starting Chips
                      </Typography>
                      <Typography variant="body1" fontWeight="bold">
                        {tournament.starting_chips.toLocaleString()}
                      </Typography>
                    </Box>
                    <Divider />
                    <Box>
                      <Typography variant="body2" color={COLORS.text.secondary}>
                        Structure
                      </Typography>
                      <Typography variant="body1">{structure.description}</Typography>
                    </Box>
                    <Divider />
                    <Box>
                      <Typography variant="body2" color={COLORS.text.secondary}>
                        Created
                      </Typography>
                      <Typography variant="body1">
                        {new Date(tournament.created_at).toLocaleString()}
                      </Typography>
                    </Box>
                    {tournament.started_at && (
                      <>
                        <Divider />
                        <Box>
                          <Typography variant="body2" color={COLORS.text.secondary}>
                            Started
                          </Typography>
                          <Typography variant="body1">
                            {new Date(tournament.started_at).toLocaleString()}
                          </Typography>
                        </Box>
                      </>
                    )}
                    {tournament.completed_at && (
                      <>
                        <Divider />
                        <Box>
                          <Typography variant="body2" color={COLORS.text.secondary}>
                            Completed
                          </Typography>
                          <Typography variant="body1">
                            {new Date(tournament.completed_at).toLocaleString()}
                          </Typography>
                        </Box>
                      </>
                    )}
                  </Stack>
                </Box>
              </Card>
            </Box>
          </Box>
        </Stack>
      </Container>
    </AppLayout>
  );
};
