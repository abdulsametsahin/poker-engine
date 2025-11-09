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
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { tableAPI, matchmakingAPI } from '../services/api';

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
  const [tables, setTables] = useState<Table[]>([]);
  const [loading, setLoading] = useState(false);

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

  const handleQuickMatch = async () => {
    try {
      setLoading(true);
      const response = await matchmakingAPI.join();
      const tableId = response.data.table_id;
      navigate(`/game/${tableId}`);
    } catch (error) {
      console.error('Matchmaking failed:', error);
      alert('Matchmaking failed');
    } finally {
      setLoading(false);
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
          <Box sx={{ textAlign: 'center' }}>
            <Button
              variant="contained"
              color="primary"
              size="large"
              onClick={handleQuickMatch}
              disabled={loading}
              sx={{ minWidth: 200 }}
            >
              Quick Match
            </Button>
            <Typography variant="caption" display="block" sx={{ mt: 1, color: 'text.secondary' }}>
              Join a random table instantly
            </Typography>
          </Box>

          <Typography variant="h5" gutterBottom>
            Available Tables
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
            {tables.map((table) => (
              <Card key={table.id} sx={{ height: '100%' }}>
                <CardContent>
                  <Stack spacing={2}>
                    <Box>
                      <Typography variant="h6" gutterBottom>
                        {table.name}
                      </Typography>
                      <Chip
                        label={table.status.toUpperCase()}
                        size="small"
                        color={table.status === 'waiting' ? 'success' : 'warning'}
                      />
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
    </Box>
  );
};
