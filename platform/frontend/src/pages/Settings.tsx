import React from 'react';
import { Box, Container, Typography, Stack, Divider } from '@mui/material';
import { AccountBalance, Person, EmojiEvents } from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import { AppLayout } from '../components/common/AppLayout';
import { Card } from '../components/common/Card';
import { Chip } from '../components/common/Chip';
import { COLORS } from '../constants';
import { formatTimestamp } from '../utils';

export const Settings: React.FC = () => {
  const { user } = useAuth();

  if (!user) {
    return null;
  }

  return (
    <AppLayout>
      <Container maxWidth="md" sx={{ mt: 4, mb: 6 }}>
        <Typography variant="h4" sx={{ mb: 4, fontWeight: 700 }}>
          Settings
        </Typography>

        {/* Account Balance Card */}
        <Card variant="elevated" sx={{ mb: 3 }}>
          <Stack spacing={3}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              <Box
                sx={{
                  width: 48,
                  height: 48,
                  borderRadius: '12px',
                  background: `linear-gradient(135deg, ${COLORS.accent.main} 0%, ${COLORS.accent.dark} 100%)`,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  boxShadow: `0 4px 12px ${COLORS.accent.glow}`,
                }}
              >
                <AccountBalance sx={{ fontSize: '1.5rem', color: COLORS.text.primary }} />
              </Box>
              <Box>
                <Typography variant="h6" fontWeight={700}>
                  Account Balance
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Your current chip balance
                </Typography>
              </Box>
            </Box>
            <Divider sx={{ borderColor: COLORS.border.main }} />
            <Box
              sx={{
                p: 3,
                background: COLORS.background.tertiary,
                borderRadius: '12px',
                textAlign: 'center',
              }}
            >
              <Typography variant="h3" fontWeight={700} color={COLORS.accent.main}>
                <Chip amount={user.chips || 0} variant="default" size="large" />
              </Typography>
            </Box>
          </Stack>
        </Card>

        {/* Account Information Card */}
        <Card variant="elevated" sx={{ mb: 3 }}>
          <Stack spacing={3}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              <Box
                sx={{
                  width: 48,
                  height: 48,
                  borderRadius: '12px',
                  background: `linear-gradient(135deg, ${COLORS.primary.main} 0%, ${COLORS.primary.dark} 100%)`,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  boxShadow: `0 4px 12px ${COLORS.primary.glow}`,
                }}
              >
                <Person sx={{ fontSize: '1.5rem', color: COLORS.text.primary }} />
              </Box>
              <Box>
                <Typography variant="h6" fontWeight={700}>
                  Account Information
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Your profile details
                </Typography>
              </Box>
            </Box>
            <Divider sx={{ borderColor: COLORS.border.main }} />
            <Stack spacing={2}>
              <Box
                sx={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  p: 2,
                  background: COLORS.background.secondary,
                  borderRadius: '8px',
                }}
              >
                <Typography variant="body2" color="text.secondary">
                  Username
                </Typography>
                <Typography variant="body2" fontWeight={600}>
                  {user.username}
                </Typography>
              </Box>
              <Box
                sx={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  p: 2,
                  background: COLORS.background.secondary,
                  borderRadius: '8px',
                }}
              >
                <Typography variant="body2" color="text.secondary">
                  Email
                </Typography>
                <Typography variant="body2" fontWeight={600}>
                  {user.email}
                </Typography>
              </Box>
              <Box
                sx={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  p: 2,
                  background: COLORS.background.secondary,
                  borderRadius: '8px',
                }}
              >
                <Typography variant="body2" color="text.secondary">
                  Member Since
                </Typography>
                <Typography variant="body2" fontWeight={600}>
                  {user.created_at ? formatTimestamp(user.created_at) : 'N/A'}
                </Typography>
              </Box>
            </Stack>
          </Stack>
        </Card>

        {/* Game Statistics Card */}
        <Card variant="elevated">
          <Stack spacing={3}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              <Box
                sx={{
                  width: 48,
                  height: 48,
                  borderRadius: '12px',
                  background: `linear-gradient(135deg, ${COLORS.secondary.main} 0%, ${COLORS.secondary.dark} 100%)`,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  boxShadow: `0 4px 12px ${COLORS.secondary.glow}`,
                }}
              >
                <EmojiEvents sx={{ fontSize: '1.5rem', color: COLORS.text.primary }} />
              </Box>
              <Box>
                <Typography variant="h6" fontWeight={700}>
                  Game Statistics
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Coming soon
                </Typography>
              </Box>
            </Box>
            <Divider sx={{ borderColor: COLORS.border.main }} />
            <Box
              sx={{
                p: 3,
                background: COLORS.background.tertiary,
                borderRadius: '8px',
                textAlign: 'center',
              }}
            >
              <Typography variant="body2" color="text.secondary">
                Game statistics will be available in a future update
              </Typography>
            </Box>
          </Stack>
        </Card>
      </Container>
    </AppLayout>
  );
};
