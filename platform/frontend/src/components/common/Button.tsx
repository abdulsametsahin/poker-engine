import React from 'react';
import { Button as MuiButton, ButtonProps as MuiButtonProps, CircularProgress } from '@mui/material';
import { COLORS } from '../../constants';

interface ButtonProps extends Omit<MuiButtonProps, 'variant'> {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'success' | 'warning';
  loading?: boolean;
  fullWidth?: boolean;
}

export const Button: React.FC<ButtonProps> = ({
  variant = 'primary',
  loading = false,
  children,
  disabled,
  fullWidth = false,
  sx,
  ...props
}) => {
  const variantStyles = {
    primary: {
      background: `linear-gradient(145deg, ${COLORS.primary.light} 0%, ${COLORS.primary.main} 50%, ${COLORS.primary.dark} 100%)`,
      color: COLORS.text.primary,
      fontWeight: 700,
      letterSpacing: '0.05em',
      border: `2px solid ${COLORS.primary.main}`,
      borderRadius: '12px',
      boxShadow: `
        0 4px 12px rgba(0, 0, 0, 0.4),
        0 2px 6px rgba(0, 0, 0, 0.3),
        inset 0 2px 4px rgba(255, 255, 255, 0.2),
        inset 0 -2px 4px rgba(0, 0, 0, 0.3)
      `,
      transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
      '&:hover': {
        background: `linear-gradient(145deg, ${COLORS.primary.main} 0%, ${COLORS.primary.light} 50%, ${COLORS.primary.main} 100%)`,
        boxShadow: `
          0 6px 16px rgba(0, 0, 0, 0.5),
          0 3px 8px rgba(0, 0, 0, 0.4),
          0 0 24px ${COLORS.primary.glow},
          inset 0 2px 4px rgba(255, 255, 255, 0.3),
          inset 0 -2px 4px rgba(0, 0, 0, 0.2)
        `,
        transform: 'translateY(-2px)',
      },
      '&:active': {
        transform: 'translateY(0px)',
        boxShadow: `
          0 2px 6px rgba(0, 0, 0, 0.4),
          inset 0 2px 6px rgba(0, 0, 0, 0.3)
        `,
      },
      '&:disabled': {
        background: 'rgba(60, 60, 60, 0.5)',
        color: COLORS.text.disabled,
        border: `2px solid rgba(60, 60, 60, 0.8)`,
        boxShadow: 'none',
      },
    },
    secondary: {
      background: `linear-gradient(145deg, ${COLORS.secondary.light} 0%, ${COLORS.secondary.main} 50%, ${COLORS.secondary.dark} 100%)`,
      color: COLORS.text.primary,
      fontWeight: 700,
      letterSpacing: '0.05em',
      border: `2px solid ${COLORS.secondary.main}`,
      borderRadius: '12px',
      boxShadow: `
        0 4px 12px rgba(0, 0, 0, 0.4),
        0 2px 6px rgba(0, 0, 0, 0.3),
        inset 0 2px 4px rgba(255, 255, 255, 0.2),
        inset 0 -2px 4px rgba(0, 0, 0, 0.3)
      `,
      transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
      '&:hover': {
        background: `linear-gradient(145deg, ${COLORS.secondary.main} 0%, ${COLORS.secondary.light} 50%, ${COLORS.secondary.main} 100%)`,
        boxShadow: `
          0 6px 16px rgba(0, 0, 0, 0.5),
          0 3px 8px rgba(0, 0, 0, 0.4),
          0 0 24px ${COLORS.secondary.glow},
          inset 0 2px 4px rgba(255, 255, 255, 0.3),
          inset 0 -2px 4px rgba(0, 0, 0, 0.2)
        `,
        transform: 'translateY(-2px)',
      },
      '&:active': {
        transform: 'translateY(0px)',
        boxShadow: `
          0 2px 6px rgba(0, 0, 0, 0.4),
          inset 0 2px 6px rgba(0, 0, 0, 0.3)
        `,
      },
      '&:disabled': {
        background: 'rgba(60, 60, 60, 0.5)',
        color: COLORS.text.disabled,
        border: `2px solid rgba(60, 60, 60, 0.8)`,
        boxShadow: 'none',
      },
    },
    danger: {
      background: `linear-gradient(145deg, ${COLORS.danger.light} 0%, ${COLORS.danger.main} 50%, ${COLORS.danger.dark} 100%)`,
      color: COLORS.text.primary,
      fontWeight: 700,
      letterSpacing: '0.05em',
      border: `2px solid ${COLORS.danger.main}`,
      borderRadius: '12px',
      boxShadow: `
        0 4px 12px rgba(0, 0, 0, 0.4),
        0 2px 6px rgba(0, 0, 0, 0.3),
        inset 0 2px 4px rgba(255, 255, 255, 0.2),
        inset 0 -2px 4px rgba(0, 0, 0, 0.3)
      `,
      transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
      '&:hover': {
        background: `linear-gradient(145deg, ${COLORS.danger.main} 0%, ${COLORS.danger.light} 50%, ${COLORS.danger.main} 100%)`,
        boxShadow: `
          0 6px 16px rgba(0, 0, 0, 0.5),
          0 3px 8px rgba(0, 0, 0, 0.4),
          0 0 24px ${COLORS.danger.glow},
          inset 0 2px 4px rgba(255, 255, 255, 0.3),
          inset 0 -2px 4px rgba(0, 0, 0, 0.2)
        `,
        transform: 'translateY(-2px)',
      },
      '&:active': {
        transform: 'translateY(0px)',
        boxShadow: `
          0 2px 6px rgba(0, 0, 0, 0.4),
          inset 0 2px 6px rgba(0, 0, 0, 0.3)
        `,
      },
      '&:disabled': {
        background: 'rgba(60, 60, 60, 0.5)',
        color: COLORS.text.disabled,
        border: `2px solid rgba(60, 60, 60, 0.8)`,
        boxShadow: 'none',
      },
    },
    success: {
      background: `linear-gradient(145deg, ${COLORS.success.light} 0%, ${COLORS.success.main} 50%, ${COLORS.success.dark} 100%)`,
      color: COLORS.text.primary,
      fontWeight: 700,
      letterSpacing: '0.05em',
      border: `2px solid ${COLORS.success.main}`,
      borderRadius: '12px',
      boxShadow: `
        0 4px 12px rgba(0, 0, 0, 0.4),
        0 2px 6px rgba(0, 0, 0, 0.3),
        inset 0 2px 4px rgba(255, 255, 255, 0.2),
        inset 0 -2px 4px rgba(0, 0, 0, 0.3)
      `,
      transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
      '&:hover': {
        background: `linear-gradient(145deg, ${COLORS.success.main} 0%, ${COLORS.success.light} 50%, ${COLORS.success.main} 100%)`,
        boxShadow: `
          0 6px 16px rgba(0, 0, 0, 0.5),
          0 3px 8px rgba(0, 0, 0, 0.4),
          0 0 24px ${COLORS.success.glow},
          inset 0 2px 4px rgba(255, 255, 255, 0.3),
          inset 0 -2px 4px rgba(0, 0, 0, 0.2)
        `,
        transform: 'translateY(-2px)',
      },
      '&:active': {
        transform: 'translateY(0px)',
        boxShadow: `
          0 2px 6px rgba(0, 0, 0, 0.4),
          inset 0 2px 6px rgba(0, 0, 0, 0.3)
        `,
      },
      '&:disabled': {
        background: 'rgba(60, 60, 60, 0.5)',
        color: COLORS.text.disabled,
        border: `2px solid rgba(60, 60, 60, 0.8)`,
        boxShadow: 'none',
      },
    },
    warning: {
      background: `linear-gradient(145deg, ${COLORS.warning.light} 0%, ${COLORS.warning.main} 50%, ${COLORS.warning.dark} 100%)`,
      color: COLORS.text.inverse,
      fontWeight: 700,
      letterSpacing: '0.05em',
      border: `2px solid ${COLORS.warning.main}`,
      borderRadius: '12px',
      boxShadow: `
        0 4px 12px rgba(0, 0, 0, 0.4),
        0 2px 6px rgba(0, 0, 0, 0.3),
        inset 0 2px 4px rgba(255, 255, 255, 0.3),
        inset 0 -2px 4px rgba(0, 0, 0, 0.2)
      `,
      transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
      '&:hover': {
        background: `linear-gradient(145deg, ${COLORS.warning.main} 0%, ${COLORS.warning.light} 50%, ${COLORS.warning.main} 100%)`,
        boxShadow: `
          0 6px 16px rgba(0, 0, 0, 0.5),
          0 3px 8px rgba(0, 0, 0, 0.4),
          0 0 24px ${COLORS.warning.glow},
          inset 0 2px 4px rgba(255, 255, 255, 0.4),
          inset 0 -2px 4px rgba(0, 0, 0, 0.2)
        `,
        transform: 'translateY(-2px)',
      },
      '&:active': {
        transform: 'translateY(0px)',
        boxShadow: `
          0 2px 6px rgba(0, 0, 0, 0.4),
          inset 0 2px 6px rgba(0, 0, 0, 0.3)
        `,
      },
      '&:disabled': {
        background: 'rgba(60, 60, 60, 0.5)',
        color: COLORS.text.disabled,
        border: `2px solid rgba(60, 60, 60, 0.8)`,
        boxShadow: 'none',
      },
    },
    ghost: {
      background: 'transparent',
      color: COLORS.text.primary,
      fontWeight: 600,
      letterSpacing: '0.05em',
      border: `2px solid ${COLORS.border.main}`,
      borderRadius: '12px',
      transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
      '&:hover': {
        background: COLORS.background.tertiary,
        borderColor: COLORS.primary.main,
        boxShadow: `0 4px 12px rgba(0, 0, 0, 0.3)`,
        transform: 'translateY(-1px)',
      },
      '&:active': {
        transform: 'translateY(0px)',
      },
      '&:disabled': {
        borderColor: COLORS.border.light,
        color: COLORS.text.disabled,
      },
    },
  };

  return (
    <MuiButton
      {...props}
      disabled={disabled || loading}
      fullWidth={fullWidth}
      sx={{
        ...variantStyles[variant],
        position: 'relative',
        ...(fullWidth && { width: '100%' }),
        ...sx,
      }}
    >
      {loading && (
        <CircularProgress
          size={20}
          sx={{
            position: 'absolute',
            color: COLORS.text.primary,
          }}
        />
      )}
      <span style={{ opacity: loading ? 0 : 1 }}>{children}</span>
    </MuiButton>
  );
};
