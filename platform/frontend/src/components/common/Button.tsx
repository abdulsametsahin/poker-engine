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
      background: `linear-gradient(135deg, ${COLORS.primary.main} 0%, ${COLORS.primary.dark} 100%)`,
      color: COLORS.text.primary,
      '&:hover': {
        background: `linear-gradient(135deg, ${COLORS.primary.light} 0%, ${COLORS.primary.main} 100%)`,
        boxShadow: `0 0 20px ${COLORS.primary.glow}`,
      },
      '&:disabled': {
        background: COLORS.background.tertiary,
        color: COLORS.text.disabled,
      },
    },
    secondary: {
      background: `linear-gradient(135deg, ${COLORS.secondary.main} 0%, ${COLORS.secondary.dark} 100%)`,
      color: COLORS.text.primary,
      '&:hover': {
        background: `linear-gradient(135deg, ${COLORS.secondary.light} 0%, ${COLORS.secondary.main} 100%)`,
        boxShadow: `0 0 20px ${COLORS.secondary.glow}`,
      },
      '&:disabled': {
        background: COLORS.background.tertiary,
        color: COLORS.text.disabled,
      },
    },
    danger: {
      background: `linear-gradient(135deg, ${COLORS.danger.main} 0%, ${COLORS.danger.dark} 100%)`,
      color: COLORS.text.primary,
      '&:hover': {
        background: `linear-gradient(135deg, ${COLORS.danger.light} 0%, ${COLORS.danger.main} 100%)`,
        boxShadow: `0 0 20px ${COLORS.danger.glow}`,
      },
      '&:disabled': {
        background: COLORS.background.tertiary,
        color: COLORS.text.disabled,
      },
    },
    success: {
      background: `linear-gradient(135deg, ${COLORS.success.main} 0%, ${COLORS.success.dark} 100%)`,
      color: COLORS.text.primary,
      '&:hover': {
        background: `linear-gradient(135deg, ${COLORS.success.light} 0%, ${COLORS.success.main} 100%)`,
        boxShadow: `0 0 20px ${COLORS.success.glow}`,
      },
      '&:disabled': {
        background: COLORS.background.tertiary,
        color: COLORS.text.disabled,
      },
    },
    warning: {
      background: `linear-gradient(135deg, ${COLORS.warning.main} 0%, ${COLORS.warning.dark} 100%)`,
      color: COLORS.text.inverse,
      '&:hover': {
        background: `linear-gradient(135deg, ${COLORS.warning.light} 0%, ${COLORS.warning.main} 100%)`,
        boxShadow: `0 0 20px ${COLORS.warning.glow}`,
      },
      '&:disabled': {
        background: COLORS.background.tertiary,
        color: COLORS.text.disabled,
      },
    },
    ghost: {
      background: 'transparent',
      color: COLORS.text.primary,
      border: `1px solid ${COLORS.border.main}`,
      '&:hover': {
        background: COLORS.background.tertiary,
        borderColor: COLORS.primary.main,
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
