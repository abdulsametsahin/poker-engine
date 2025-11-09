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
    if (percentage > 50) return '#10b981';
    if (percentage > 25) return '#fbbf24';
    return '#ef4444';
  };

  return (
    <Box sx={{ width: '100%' }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 0.5 }}>
        <Typography variant="caption" sx={{ color: 'rgba(156, 163, 175, 0.7)', fontSize: '10px', fontWeight: 600 }}>
          Time
        </Typography>
        <Typography
          variant="caption"
          fontWeight="bold"
          sx={{ color: getColor(), fontSize: '11px' }}
        >
          {timeLeft}s
        </Typography>
      </Box>
      <Box
        sx={{
          width: '100%',
          height: 4,
          borderRadius: 2,
          bgcolor: 'rgba(31, 41, 55, 0.5)',
          overflow: 'hidden',
        }}
      >
        <Box
          sx={{
            width: `${percentage}%`,
            height: '100%',
            bgcolor: getColor(),
            transition: 'width 0.1s linear',
            boxShadow: `0 0 8px ${getColor()}`,
          }}
        />
      </Box>
    </Box>
  );
};

export default ActionTimer;
