import React from 'react';
import { Box, Typography, Stack } from '@mui/material';
import { Card } from './Card';
import { Button } from './Button';
import { COLORS } from '../../constants';

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
  queueCount?: number;
}

export const GameModeCard: React.FC<GameModeCardProps> = ({
  title,
  description,
  blinds,
  buyIn,
  maxPlayers,
  icon,
  color,
  onJoin,
  disabled,
  queueCount,
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
      {/* Background accent */}
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
        {/* Icon and Title */}
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

        {/* Details */}
        <Box
          sx={{
            padding: 2,
            background: COLORS.background.tertiary,
            borderRadius: '8px',
          }}
        >
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
            {queueCount !== undefined && queueCount > 0 && (
              <Box
                sx={{
                  mt: 1,
                  pt: 1,
                  borderTop: `1px solid ${COLORS.border.light}`,
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                }}
              >
                <Typography variant="caption" color="text.secondary">
                  In queue:
                </Typography>
                <Box
                  sx={{
                    px: 1,
                    py: 0.5,
                    borderRadius: '4px',
                    background: `${color}20`,
                    color,
                    fontWeight: 600,
                    fontSize: '0.75rem',
                  }}
                >
                  {queueCount} player{queueCount !== 1 ? 's' : ''}
                </Box>
              </Box>
            )}
          </Stack>
        </Box>

        {/* Action Button */}
        <Button variant="primary" size="large" fullWidth onClick={onJoin} disabled={disabled}>
          Find Match
        </Button>
      </Stack>
    </Card>
  );
};
