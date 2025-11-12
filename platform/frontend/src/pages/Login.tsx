import React, { useState } from 'react';
import { Box, TextField, Typography, Stack, LinearProgress, InputAdornment, IconButton } from '@mui/material';
import { Visibility, VisibilityOff, Person, Email, Lock } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { authAPI } from '../services/api';
import { useAuth } from '../contexts/AuthContext';
import { useToast } from '../contexts/ToastContext';
import { Logo } from '../components/common/Logo';
import { Button } from '../components/common/Button';
import { Card } from '../components/common/Card';
import { COLORS, ROUTES, GAME } from '../constants';
import { validateUsername, validatePassword, validateEmail, getPasswordStrength } from '../utils';

export const Login: React.FC = () => {
  const navigate = useNavigate();
  const { login: authLogin } = useAuth();
  const { showError, showSuccess } = useToast();
  const [isLogin, setIsLogin] = useState(true);
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const [loginData, setLoginData] = useState({
    username: '',
    password: '',
  });

  const [registerData, setRegisterData] = useState({
    username: 'test1',
    email: 'test1@example.com',
    password: '1q2w3e++',
    confirmPassword: '1q2w3e++',
  });

  const passwordStrength = getPasswordStrength(registerData.password);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const response = await authAPI.login(loginData);
      authLogin(response.data.token, response.data.user);
      showSuccess('Welcome back to PokerStreet!');
      navigate(ROUTES.LOBBY);
    } catch (err: any) {
      showError(err.response?.data?.error || 'Login failed. Please check your credentials.');
    } finally {
      setLoading(false);
    }
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validation
    const usernameValidation = validateUsername(registerData.username);
    if (!usernameValidation.valid) {
      showError(usernameValidation.error || 'Invalid username');
      return;
    }

    if (!validateEmail(registerData.email)) {
      showError('Please enter a valid email address');
      return;
    }

    const passwordValidation = validatePassword(registerData.password);
    if (!passwordValidation.valid) {
      showError(passwordValidation.error || 'Invalid password');
      return;
    }

    if (registerData.password !== registerData.confirmPassword) {
      showError('Passwords do not match');
      return;
    }

    setLoading(true);

    try {
      const response = await authAPI.register({
        username: registerData.username,
        email: registerData.email,
        password: registerData.password,
      });
      authLogin(response.data.token, response.data.user);
      showSuccess('Welcome to PokerStreet!');
      navigate(ROUTES.LOBBY);
    } catch (err: any) {
      showError(err.response?.data?.error || 'Registration failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const getPasswordStrengthColor = () => {
    if (!registerData.password) return COLORS.text.disabled;
    switch (passwordStrength.strength) {
      case 'weak':
        return COLORS.danger.main;
      case 'medium':
        return COLORS.warning.main;
      case 'strong':
        return COLORS.success.main;
      default:
        return COLORS.text.disabled;
    }
  };

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        overflow: 'hidden',
        position: 'relative',
        background: COLORS.background.primary,
      }}
    >
      {/* Animated Background */}
      <Box
        sx={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          opacity: 0.05,
          background: `
            radial-gradient(circle at 20% 50%, ${COLORS.primary.main} 0%, transparent 50%),
            radial-gradient(circle at 80% 50%, ${COLORS.secondary.main} 0%, transparent 50%)
          `,
          animation: 'pulse 8s ease-in-out infinite',
          '@keyframes pulse': {
            '0%, 100%': { opacity: 0.05 },
            '50%': { opacity: 0.1 },
          },
        }}
      />

      {/* Left Side - Branding */}
      <Box
        sx={{
          flex: 1,
          display: { xs: 'none', md: 'flex' },
          flexDirection: 'column',
          justifyContent: 'center',
          alignItems: 'center',
          padding: 8,
          position: 'relative',
          '&::before': {
            content: '""',
            position: 'absolute',
            top: 0,
            right: 0,
            bottom: 0,
            width: 1,
            background: `linear-gradient(to bottom, transparent, ${COLORS.border.main}, transparent)`,
          },
        }}
      >
        <Logo size="large" />
        <Typography
          variant="h3"
          sx={{
            mt: 4,
            mb: 2,
            textAlign: 'center',
            fontWeight: 700,
            background: `linear-gradient(135deg, ${COLORS.primary.light} 0%, ${COLORS.secondary.light} 100%)`,
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
          }}
        >
          Welcome to PokerStreet
        </Typography>
        <Typography
          variant="h6"
          sx={{
            textAlign: 'center',
            color: COLORS.text.secondary,
            maxWidth: 400,
          }}
        >
          The premium online poker experience where street meets sophistication
        </Typography>

        {/* Decorative elements */}
        <Box
          sx={{
            mt: 8,
            display: 'flex',
            gap: 4,
            color: COLORS.text.disabled,
          }}
        >
          <Box sx={{ textAlign: 'center' }}>
            <Typography variant="h4" sx={{ fontWeight: 700, color: COLORS.primary.light }}>
              24/7
            </Typography>
            <Typography variant="caption">Live Games</Typography>
          </Box>
          <Box sx={{ textAlign: 'center' }}>
            <Typography variant="h4" sx={{ fontWeight: 700, color: COLORS.secondary.light }}>
              Fast
            </Typography>
            <Typography variant="caption">Matchmaking</Typography>
          </Box>
          <Box sx={{ textAlign: 'center' }}>
            <Typography variant="h4" sx={{ fontWeight: 700, color: COLORS.accent.main }}>
              Fair
            </Typography>
            <Typography variant="caption">Play</Typography>
          </Box>
        </Box>
      </Box>

      {/* Right Side - Form */}
      <Box
        sx={{
          flex: 1,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: { xs: 3, sm: 4, md: 8 },
        }}
      >
        <Card
          variant="glass"
          sx={{
            maxWidth: 480,
            width: '100%',
            padding: { xs: 3, sm: 4 },
          }}
        >
          {/* Mobile Logo */}
          <Box sx={{ display: { xs: 'flex', md: 'none' }, justifyContent: 'center', mb: 4 }}>
            <Logo size="medium" />
          </Box>

          {/* Toggle */}
          <Box
            sx={{
              display: 'flex',
              gap: 1,
              mb: 4,
              p: 0.5,
              background: COLORS.background.tertiary,
              borderRadius: '8px',
            }}
          >
            <Button
              fullWidth
              variant={isLogin ? 'primary' : 'ghost'}
              onClick={() => setIsLogin(true)}
              sx={{ flex: 1 }}
            >
              Login
            </Button>
            <Button
              fullWidth
              variant={!isLogin ? 'primary' : 'ghost'}
              onClick={() => setIsLogin(false)}
              sx={{ flex: 1 }}
            >
              Register
            </Button>
          </Box>

          {/* Login Form */}
          {isLogin ? (
            <form onSubmit={handleLogin}>
              <Stack spacing={3}>
                <TextField
                  label="Username"
                  fullWidth
                  required
                  autoFocus
                  value={loginData.username}
                  onChange={(e) => setLoginData({ ...loginData, username: e.target.value })}
                  InputProps={{
                    startAdornment: (
                      <InputAdornment position="start">
                        <Person sx={{ color: COLORS.text.secondary }} />
                      </InputAdornment>
                    ),
                  }}
                />
                <TextField
                  label="Password"
                  type={showPassword ? 'text' : 'password'}
                  fullWidth
                  required
                  value={loginData.password}
                  onChange={(e) => setLoginData({ ...loginData, password: e.target.value })}
                  InputProps={{
                    startAdornment: (
                      <InputAdornment position="start">
                        <Lock sx={{ color: COLORS.text.secondary }} />
                      </InputAdornment>
                    ),
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton
                          onClick={() => setShowPassword(!showPassword)}
                          edge="end"
                        >
                          {showPassword ? <VisibilityOff /> : <Visibility />}
                        </IconButton>
                      </InputAdornment>
                    ),
                  }}
                />
                <Button type="submit" variant="primary" size="large" fullWidth loading={loading}>
                  Login
                </Button>
              </Stack>
            </form>
          ) : (
            /* Register Form */
            <form onSubmit={handleRegister}>
              <Stack spacing={3}>
                <TextField
                  label="Username"
                  fullWidth
                  required
                  autoFocus
                  value={registerData.username}
                  onChange={(e) => setRegisterData({ ...registerData, username: e.target.value })}
                  helperText={`${GAME.MIN_USERNAME_LENGTH}-${GAME.MAX_USERNAME_LENGTH} characters, letters, numbers, and underscores`}
                  InputProps={{
                    startAdornment: (
                      <InputAdornment position="start">
                        <Person sx={{ color: COLORS.text.secondary }} />
                      </InputAdornment>
                    ),
                  }}
                />
                <TextField
                  label="Email"
                  type="email"
                  fullWidth
                  required
                  value={registerData.email}
                  onChange={(e) => setRegisterData({ ...registerData, email: e.target.value })}
                  InputProps={{
                    startAdornment: (
                      <InputAdornment position="start">
                        <Email sx={{ color: COLORS.text.secondary }} />
                      </InputAdornment>
                    ),
                  }}
                />
                <Box>
                  <TextField
                    label="Password"
                    type={showPassword ? 'text' : 'password'}
                    fullWidth
                    required
                    value={registerData.password}
                    onChange={(e) => setRegisterData({ ...registerData, password: e.target.value })}
                    helperText={`At least ${GAME.MIN_PASSWORD_LENGTH} characters`}
                    InputProps={{
                      startAdornment: (
                        <InputAdornment position="start">
                          <Lock sx={{ color: COLORS.text.secondary }} />
                        </InputAdornment>
                      ),
                      endAdornment: (
                        <InputAdornment position="end">
                          <IconButton
                            onClick={() => setShowPassword(!showPassword)}
                            edge="end"
                          >
                            {showPassword ? <VisibilityOff /> : <Visibility />}
                          </IconButton>
                        </InputAdornment>
                      ),
                    }}
                  />
                  {registerData.password && (
                    <Box sx={{ mt: 1 }}>
                      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
                        <Typography variant="caption" color="text.secondary">
                          Password strength:
                        </Typography>
                        <Typography
                          variant="caption"
                          sx={{ color: getPasswordStrengthColor(), fontWeight: 600, textTransform: 'capitalize' }}
                        >
                          {passwordStrength.strength}
                        </Typography>
                      </Box>
                      <LinearProgress
                        variant="determinate"
                        value={(passwordStrength.score / 5) * 100}
                        sx={{
                          height: 4,
                          borderRadius: 2,
                          backgroundColor: COLORS.background.tertiary,
                          '& .MuiLinearProgress-bar': {
                            backgroundColor: getPasswordStrengthColor(),
                          },
                        }}
                      />
                    </Box>
                  )}
                </Box>
                <TextField
                  label="Confirm Password"
                  type={showConfirmPassword ? 'text' : 'password'}
                  fullWidth
                  required
                  value={registerData.confirmPassword}
                  onChange={(e) => setRegisterData({ ...registerData, confirmPassword: e.target.value })}
                  error={registerData.confirmPassword !== '' && registerData.password !== registerData.confirmPassword}
                  helperText={
                    registerData.confirmPassword !== '' && registerData.password !== registerData.confirmPassword
                      ? 'Passwords do not match'
                      : ''
                  }
                  InputProps={{
                    startAdornment: (
                      <InputAdornment position="start">
                        <Lock sx={{ color: COLORS.text.secondary }} />
                      </InputAdornment>
                    ),
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton
                          onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                          edge="end"
                        >
                          {showConfirmPassword ? <VisibilityOff /> : <Visibility />}
                        </IconButton>
                      </InputAdornment>
                    ),
                  }}
                />
                <Button type="submit" variant="primary" size="large" fullWidth loading={loading}>
                  Create Account
                </Button>
              </Stack>
            </form>
          )}
        </Card>
      </Box>
    </Box>
  );
};
