import React, { useEffect, useState, useCallback } from 'react';
import {
  Box,
  Container,
  Typography,
  Stack,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  IconButton,
} from '@mui/material';
import { EmojiEvents, Add, ContentCopy, Close } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { tournamentAPI } from '../services/api';
import { useAuth } from '../contexts/AuthContext';
import { useToast } from '../contexts/ToastContext';
import { useWebSocket } from '../contexts/WebSocketContext';
import { AppLayout } from '../components/common/AppLayout';
import { Button } from '../components/common/Button';
import { Card } from '../components/common/Card';
import { Badge } from '../components/common/Badge';
import { EmptyState } from '../components/common/EmptyState';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { COLORS } from '../constants';

interface Tournament {
  id: string;
  tournament_code: string;
  name: string;
  status: string;
  buy_in: number;
  starting_chips: number;
  max_players: number;
  min_players: number;
  current_players: number;
  prize_pool: number;
  structure: string;
  prize_structure: string;
  start_time?: string;
  auto_start_delay: number;
  created_at: string;
}

export const Tournaments: React.FC = () => {
  const { user } = useAuth();
  const { showSuccess, showError } = useToast();
  const { addMessageHandler } = useWebSocket();
  const navigate = useNavigate();
  const [tournaments, setTournaments] = useState<Tournament[]>([]);
  const [registeredTournaments, setRegisteredTournaments] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(false);
  const [initialLoad, setInitialLoad] = useState(true);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);

  // Form state for tournament creation
  const [formData, setFormData] = useState({
    name: '',
    buy_in: 100,
    starting_chips: 1000,
    max_players: 10,
    min_players: 2,
    structure: 'standard',
    prize_structure: 'top3',
    auto_start_delay: 300,
  });

  const fetchTournaments = useCallback(async () => {
    try {
      const response = await tournamentAPI.getTournaments();
      const tournamentsData = Array.isArray(response.data) ? response.data : [];
      setTournaments(tournamentsData);

      // Check registration status for each tournament
      if (user && tournamentsData.length > 0) {
        const registeredIds = new Set<string>();
        await Promise.all(
          tournamentsData.map(async (tournament) => {
            try {
              const playersRes = await tournamentAPI.getTournamentPlayers(tournament.id);
              const isRegistered = playersRes.data?.some((p: any) => p.user_id === user.id);
              if (isRegistered) {
                registeredIds.add(tournament.id);
              }
            } catch (error) {
              console.error(`Failed to check registration for tournament ${tournament.id}:`, error);
            }
          })
        );
        setRegisteredTournaments(registeredIds);
      }
    } catch (error: any) {
      console.error('Failed to fetch tournaments:', error);
    } finally {
      setInitialLoad(false);
    }
  }, [user]);

  useEffect(() => {
    fetchTournaments();
  }, [fetchTournaments]);

  // Setup WebSocket handlers for real-time tournament updates
  useEffect(() => {
    const handleTournamentPaused = (message: { payload: { tournament_id: string } }) => {
      const tournamentId = message.payload?.tournament_id;
      if (tournamentId) {
        // Update the tournament status in the list
        setTournaments(prev => prev.map(t =>
          t.id === tournamentId
            ? { ...t, status: 'paused' }
            : t
        ));
      }
    };

    const handleTournamentResumed = (message: { payload: { tournament_id: string } }) => {
      const tournamentId = message.payload?.tournament_id;
      if (tournamentId) {
        // Update the tournament status in the list
        setTournaments(prev => prev.map(t =>
          t.id === tournamentId
            ? { ...t, status: 'in_progress' }
            : t
        ));
      }
    };

    const handleTournamentUpdate = (message: { payload: { tournament?: Tournament } }) => {
      if (message.payload?.tournament) {
        const updatedTournament = message.payload.tournament;
        setTournaments(prev => prev.map(t =>
          t.id === updatedTournament.id
            ? { ...t, ...updatedTournament }
            : t
        ));
      }
    };

    const handleTournamentCreated = (message: { payload: {
      tournament_id: string;
      name: string;
      buy_in: number;
      starting_chips: number;
      max_players: number;
      min_players: number;
      status: string;
      created_at: string;
    } }) => {
      const newTournament: Tournament = {
        id: message.payload.tournament_id,
        tournament_code: '', // Will be filled when fetched
        name: message.payload.name,
        status: message.payload.status as any,
        buy_in: message.payload.buy_in,
        starting_chips: message.payload.starting_chips,
        min_players: message.payload.min_players,
        max_players: message.payload.max_players,
        current_players: 1, // Creator is first player
        prize_pool: message.payload.buy_in,
        created_at: message.payload.created_at,
        structure: '{}',
        prize_structure: 'winner_takes_all',
        auto_start_delay: 300,
      };

      setTournaments(prev => [newTournament, ...prev]);
      showSuccess(`New tournament "${message.payload.name}" created!`);
      console.log('[Tournaments] New tournament created:', newTournament.id);
    };

    const handleTournamentStarted = (message: { payload: { tournament_id: string } }) => {
      const { tournament_id } = message.payload;

      setTournaments(prev => prev.map(t =>
        t.id === tournament_id
          ? { ...t, status: 'in_progress' }
          : t
      ));

      console.log('[Tournaments] Tournament started:', tournament_id);
    };

    const handleTournamentPlayerRegistered = (message: { payload: {
      tournament_id: string;
      user_id: string;
      username: string;
    } }) => {
      const { tournament_id } = message.payload;

      setTournaments(prev => prev.map(t =>
        t.id === tournament_id
          ? {
              ...t,
              current_players: (t.current_players || 0) + 1,
              prize_pool: (t.current_players || 0 + 1) * t.buy_in,
            }
          : t
      ));

      console.log('[Tournaments] Player registered for tournament:', tournament_id);
    };

    const cleanup1 = addMessageHandler('tournament_paused', handleTournamentPaused);
    const cleanup2 = addMessageHandler('tournament_resumed', handleTournamentResumed);
    const cleanup3 = addMessageHandler('tournament_update', handleTournamentUpdate);
    const cleanup4 = addMessageHandler('tournament_created', handleTournamentCreated);
    const cleanup5 = addMessageHandler('tournament_started', handleTournamentStarted);
    const cleanup6 = addMessageHandler('tournament_player_registered', handleTournamentPlayerRegistered);

    return () => {
      cleanup1();
      cleanup2();
      cleanup3();
      cleanup4();
      cleanup5();
      cleanup6();
    };
  }, [addMessageHandler, showSuccess]);

  const handleCreateTournament = async () => {
    try {
      setLoading(true);
      const response = await tournamentAPI.createTournament(formData);
      showSuccess('Tournament created successfully!');
      setCreateDialogOpen(false);
      fetchTournaments();

      // Show the tournament code (backend returns tournament directly)
      const tournament = response.data;
      showSuccess(`Tournament Code: ${tournament.tournament_code} - Share this code with players!`);
    } catch (error: any) {
      showError(error.response?.data?.error || 'Failed to create tournament');
    } finally {
      setLoading(false);
    }
  };


  const copyTournamentCode = (code: string) => {
    navigator.clipboard.writeText(code);
    showSuccess('Tournament code copied to clipboard!');
  };

  const getStructureName = (structureJson: string): string => {
    try {
      const parsed = JSON.parse(structureJson);
      return parsed.name || 'Unknown';
    } catch {
      return 'Unknown';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'registering':
        return 'info';
      case 'starting':
        return 'warning';
      case 'in_progress':
        return 'success';
      case 'completed':
        return 'default';
      case 'cancelled':
        return 'error';
      default:
        return 'default';
    }
  };

  if (initialLoad) {
    return (
      <AppLayout>
        <LoadingSpinner fullScreen message="Loading tournaments..." />
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <Container maxWidth="lg" sx={{ mt: 4, pb: 6 }}>
        {/* Header */}
        <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Box>
            <Typography variant="h4" fontWeight={700} sx={{ mb: 1 }}>
              <EmojiEvents sx={{ mr: 1, verticalAlign: 'middle', fontSize: 40 }} />
              Tournaments
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Compete for glory and prizes
            </Typography>
          </Box>
          <Button
            startIcon={<Add />}
            onClick={() => setCreateDialogOpen(true)}
            size="large"
          >
            Create Tournament
          </Button>
        </Box>

        {/* Tournaments Grid */}
        {tournaments.length > 0 ? (
          <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)' }, gap: 3 }}>
            {tournaments.map((tournament) => (
              <Box
                key={tournament.id}
                onClick={() => navigate(`/tournaments/${tournament.id}`)}
                sx={{ cursor: 'pointer', transition: 'transform 0.2s', '&:hover': { transform: 'translateY(-4px)' } }}
              >
                <Card variant="glass">
                  <Stack spacing={2}>
                    <Box>
                      <Typography variant="h6" fontWeight={700} sx={{ mb: 1 }}>
                        {tournament.name}
                      </Typography>
                      <Stack direction="row" spacing={1} flexWrap="wrap" sx={{ mb: 2 }}>
                        <Badge variant={getStatusColor(tournament.status) as any}>
                          {tournament.status.toUpperCase()}
                        </Badge>
                        <Badge variant="info" size="small">
                          {tournament.current_players}/{tournament.max_players} Players
                        </Badge>
                        {registeredTournaments.has(tournament.id) && (
                          <Badge variant="success" size="small">
                            REGISTERED
                          </Badge>
                        )}
                      </Stack>
                    </Box>

                    <Box sx={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 2 }}>
                      <Box>
                        <Typography variant="caption" color="text.secondary">
                          Buy-in
                        </Typography>
                        <Typography variant="body1" fontWeight={600}>
                          ${tournament.buy_in}
                        </Typography>
                      </Box>
                      <Box>
                        <Typography variant="caption" color="text.secondary">
                          Prize Pool
                        </Typography>
                        <Typography variant="body1" fontWeight={600} color={COLORS.success.main}>
                          ${tournament.prize_pool}
                        </Typography>
                      </Box>
                      <Box>
                        <Typography variant="caption" color="text.secondary">
                          Starting Chips
                        </Typography>
                        <Typography variant="body1" fontWeight={600}>
                          {tournament.starting_chips}
                        </Typography>
                      </Box>
                      <Box>
                        <Typography variant="caption" color="text.secondary">
                          Structure
                        </Typography>
                        <Typography variant="body1" fontWeight={600}>
                          {getStructureName(tournament.structure)}
                        </Typography>
                      </Box>
                    </Box>

                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, pt: 1 }}>
                      <Typography variant="caption" color="text.secondary">
                        Code:
                      </Typography>
                      <Typography variant="body2" fontWeight={600} sx={{ fontFamily: 'monospace' }}>
                        {tournament.tournament_code}
                      </Typography>
                      <IconButton
                        size="small"
                        onClick={(e) => {
                          e.stopPropagation();
                          copyTournamentCode(tournament.tournament_code);
                        }}
                        sx={{ ml: 'auto' }}
                      >
                        <ContentCopy fontSize="small" />
                      </IconButton>
                    </Box>
                  </Stack>
                </Card>
              </Box>
            ))}
          </Box>
        ) : (
          <EmptyState
            icon={<EmojiEvents sx={{ fontSize: 80 }} />}
            title="No tournaments available"
            description="Be the first to create a tournament!"
            action={
              <Button startIcon={<Add />} onClick={() => setCreateDialogOpen(true)}>
                Create Tournament
              </Button>
            }
          />
        )}
      </Container>

      {/* Create Tournament Dialog */}
      <Dialog open={createDialogOpen} onClose={() => setCreateDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            Create Tournament
            <IconButton onClick={() => setCreateDialogOpen(false)}>
              <Close />
            </IconButton>
          </Box>
        </DialogTitle>
        <DialogContent>
          <Stack spacing={3} sx={{ mt: 1 }}>
            <TextField
              label="Tournament Name"
              fullWidth
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            />
            <TextField
              label="Buy-in (chips)"
              type="number"
              fullWidth
              value={formData.buy_in}
              onChange={(e) => setFormData({ ...formData, buy_in: parseInt(e.target.value) })}
            />
            <TextField
              label="Starting Chips"
              type="number"
              fullWidth
              value={formData.starting_chips}
              onChange={(e) => setFormData({ ...formData, starting_chips: parseInt(e.target.value) })}
            />
            <Box sx={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 2 }}>
              <TextField
                label="Min Players"
                type="number"
                value={formData.min_players}
                onChange={(e) => setFormData({ ...formData, min_players: parseInt(e.target.value) })}
              />
              <TextField
                label="Max Players"
                type="number"
                value={formData.max_players}
                onChange={(e) => setFormData({ ...formData, max_players: parseInt(e.target.value) })}
              />
            </Box>
            <FormControl fullWidth>
              <InputLabel>Blind Structure</InputLabel>
              <Select
                value={formData.structure}
                label="Blind Structure"
                onChange={(e) => setFormData({ ...formData, structure: e.target.value })}
              >
                <MenuItem value="turbo">Turbo (5 min levels)</MenuItem>
                <MenuItem value="standard">Standard (10 min levels)</MenuItem>
                <MenuItem value="deep_stack">Deep Stack (15 min levels)</MenuItem>
                <MenuItem value="hyper_turbo">Hyper Turbo (3 min levels)</MenuItem>
              </Select>
            </FormControl>
            <FormControl fullWidth>
              <InputLabel>Prize Structure</InputLabel>
              <Select
                value={formData.prize_structure}
                label="Prize Structure"
                onChange={(e) => setFormData({ ...formData, prize_structure: e.target.value })}
              >
                <MenuItem value="winner_takes_all">Winner Takes All</MenuItem>
                <MenuItem value="top3">Top 3 (50/30/20)</MenuItem>
                <MenuItem value="top5">Top 5</MenuItem>
                <MenuItem value="top10">Top 10</MenuItem>
              </Select>
            </FormControl>
            <TextField
              label="Auto-start Delay (seconds)"
              type="number"
              fullWidth
              value={formData.auto_start_delay}
              onChange={(e) => setFormData({ ...formData, auto_start_delay: parseInt(e.target.value) })}
              helperText="Time to wait after reaching minimum players before auto-starting"
            />
          </Stack>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 3 }}>
          <Button variant="ghost" onClick={() => setCreateDialogOpen(false)}>
            Cancel
          </Button>
          <Button onClick={handleCreateTournament} disabled={loading || !formData.name}>
            Create
          </Button>
        </DialogActions>
      </Dialog>
    </AppLayout>
  );
};
