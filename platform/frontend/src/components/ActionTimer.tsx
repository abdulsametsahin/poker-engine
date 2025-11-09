import React, { useEffect, useState, useRef } from 'react';
import { Box, LinearProgress, Typography } from '@mui/material';

interface ActionTimerProps {
  deadline: string | Date | number; // ISO string, Date object, or timestamp
  totalTime?: number; // Total time in seconds (default: 30)
}

const ActionTimer: React.FC<ActionTimerProps> = ({ deadline, totalTime = 30 }) => {
  const [timeLeft, setTimeLeft] = useState(0);
  const [percentage, setPercentage] = useState(100);
  const deadlineRef = useRef<number>(0);
  const totalTimeRef = useRef<number>(totalTime);

  // Only update deadline reference when it actually changes
  useEffect(() => {
    let deadlineTime: number;

    if (typeof deadline === 'string') {
      deadlineTime = new Date(deadline).getTime();
    } else if (deadline instanceof Date) {
      deadlineTime = deadline.getTime();
    } else {
      deadlineTime = deadline;
    }

    // Only update if deadline actually changed (more than 100ms difference to account for small variations)
    if (Math.abs(deadlineTime - deadlineRef.current) > 100) {
      deadlineRef.current = deadlineTime;
      totalTimeRef.current = totalTime;
    }
  }, [deadline, totalTime]);

  useEffect(() => {
    const updateTimer = () => {
      if (deadlineRef.current === 0) return;

      const now = Date.now();
      const remaining = Math.max(0, deadlineRef.current - now);
      const remainingSeconds = Math.ceil(remaining / 1000);
      const pct = Math.min(100, (remaining / (totalTimeRef.current * 1000)) * 100);

      setTimeLeft(remainingSeconds);
      setPercentage(pct);
    };

    updateTimer();
    const interval = setInterval(updateTimer, 100);

    return () => clearInterval(interval);
  }, []); // Empty dependency array - timer runs independently

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
