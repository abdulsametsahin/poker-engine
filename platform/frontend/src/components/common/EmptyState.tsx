import React from 'react';
import { Box, Typography } from '@mui/material';
import { COLORS } from '../../constants';

interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: React.ReactNode;
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  icon,
  title,
  description,
  action,
}) => {
  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 6,
        textAlign: 'center',
      }}
    >
      {icon && (
        <Box
          sx={{
            fontSize: '3rem',
            color: COLORS.text.disabled,
            mb: 2,
            opacity: 0.5,
          }}
        >
          {icon}
        </Box>
      )}
      <Typography
        variant="h6"
        sx={{
          color: COLORS.text.secondary,
          mb: 1,
        }}
      >
        {title}
      </Typography>
      {description && (
        <Typography
          variant="body2"
          sx={{
            color: COLORS.text.disabled,
            mb: 3,
            maxWidth: 400,
          }}
        >
          {description}
        </Typography>
      )}
      {action && <Box>{action}</Box>}
    </Box>
  );
};
