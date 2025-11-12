import React, { useEffect, useState, useRef } from 'react';
import { Box, LinearProgress, Typography } from '@mui/material';

interface ActionTimerProps {
  deadline?: string | Date | number | null; // ISO string, Date object, or timestamp
  totalTime?: number; // Total time in seconds (default: 30)
}

const ActionTimer: React.FC<ActionTimerProps> = ({ deadline, totalTime = 30 }) => {
  const [timeLeft, setTimeLeft] = useState(0);
  const [percentage, setPercentage] = useState(100);
  const deadlineTimeRef = useRef<number>(0);
  const startTimeRef = useRef<number>(0);
  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  // Parse and update deadline when it changes
  useEffect(() => {
    // If no deadline provided, reset to 0
    if (!deadline) {
      deadlineTimeRef.current = 0;
      startTimeRef.current = 0;
      setTimeLeft(0);
      setPercentage(0);
      return;
    }

    let deadlineTime: number;

    if (typeof deadline === 'string') {
      deadlineTime = new Date(deadline).getTime();
    } else if (deadline instanceof Date) {
      deadlineTime = deadline.getTime();
    } else {
      deadlineTime = deadline;
    }

    // Only update if deadline actually changed (more than 100ms difference)
    if (Math.abs(deadlineTime - deadlineTimeRef.current) > 100) {
      deadlineTimeRef.current = deadlineTime;
      // Calculate the original start time
      startTimeRef.current = deadlineTime - (totalTime * 1000);
    }
  }, [deadline, totalTime]);

  // Timer effect
  useEffect(() => {
    const updateTimer = () => {
      if (deadlineTimeRef.current === 0) return;

      const now = Date.now();
      const remaining = Math.max(0, deadlineTimeRef.current - now);
      const remainingSeconds = Math.ceil(remaining / 1000);
      
      // Calculate percentage based on original total time
      const totalDuration = deadlineTimeRef.current - startTimeRef.current;
      const elapsed = now - startTimeRef.current;
      const pct = Math.max(0, Math.min(100, ((totalDuration - elapsed) / totalDuration) * 100));

      setTimeLeft(remainingSeconds);
      setPercentage(pct);

      // Stop timer when time runs out
      if (remaining <= 0 && intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };

    // Initial update
    updateTimer();
    
    // Clear existing interval
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
    }

    // Start new interval
    intervalRef.current = setInterval(updateTimer, 100);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [deadline]); // Re-run when deadline changes

  const getColor = () => {
    if (percentage > 50) return '#10b981';
    if (percentage > 25) return '#fbbf24';
    return '#ef4444';
  };

  // Check if timer is paused (no deadline)
  const isPaused = !deadline || deadlineTimeRef.current === 0;

  return (
    <Box sx={{ width: '100%' }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 0.5 }}>
        <Typography variant="caption" sx={{ color: 'rgba(156, 163, 175, 0.7)', fontSize: '10px', fontWeight: 600 }}>
          Time
        </Typography>
        <Typography
          variant="caption"
          fontWeight="bold"
          sx={{ color: isPaused ? '#9ca3af' : getColor(), fontSize: '11px' }}
        >
          {isPaused ? 'PAUSED' : `${timeLeft}s`}
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
            width: isPaused ? '0%' : `${percentage}%`,
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
