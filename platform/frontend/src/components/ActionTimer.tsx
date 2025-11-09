import React, { useEffect, useState } from 'react';
import { Box, LinearProgress, Typography } from '@mui/material';

interface ActionTimerProps {
  deadline: string | Date | number; // ISO string, Date object, or timestamp
  totalTime?: number; // Total time in seconds (default: 30)
}

const ActionTimer: React.FC<ActionTimerProps> = ({ deadline, totalTime = 30 }) => {
  const [timeLeft, setTimeLeft] = useState(0);
  const [percentage, setPercentage] = useState(100);

  useEffect(() => {
    const updateTimer = () => {
      const now = Date.now();
      let deadlineTime: number;

      if (typeof deadline === 'string') {
        deadlineTime = new Date(deadline).getTime();
      } else if (deadline instanceof Date) {
        deadlineTime = deadline.getTime();
      } else {
        deadlineTime = deadline;
      }

      const remaining = Math.max(0, deadlineTime - now);
      const remainingSeconds = Math.ceil(remaining / 1000);
      const pct = Math.min(100, (remaining / (totalTime * 1000)) * 100);

      setTimeLeft(remainingSeconds);
      setPercentage(pct);
    };

    updateTimer();
    const interval = setInterval(updateTimer, 100);

    return () => clearInterval(interval);
  }, [deadline, totalTime]);

  const getColor = () => {
    if (percentage > 50) return 'success';
    if (percentage > 25) return 'warning';
    return 'error';
  };

  return (
    <Box sx={{ width: '100%' }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
        <Typography variant="caption" color="text.secondary">
          Time
        </Typography>
        <Typography
          variant="caption"
          fontWeight="bold"
          color={getColor() === 'error' ? 'error.main' : 'text.primary'}
        >
          {timeLeft}s
        </Typography>
      </Box>
      <LinearProgress
        variant="determinate"
        value={percentage}
        color={getColor()}
        sx={{
          height: 6,
          borderRadius: 3,
          bgcolor: 'rgba(0,0,0,0.1)',
        }}
      />
    </Box>
  );
};

export default ActionTimer;
