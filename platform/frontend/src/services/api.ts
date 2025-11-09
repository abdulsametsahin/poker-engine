import axios from 'axios';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_URL,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const authAPI = {
  register: (data: { username: string; email: string; password: string }) =>
    api.post('/auth/register', data),
  login: (data: { username: string; password: string }) =>
    api.post('/auth/login', data),
};

export const tableAPI = {
  getTables: () => api.get('/tables'),
  createTable: (data: any) => api.post('/tables', data),
  joinTable: (tableId: string, buyIn: number) =>
    api.post(`/tables/${tableId}/join`, { buy_in: buyIn }),
};

export const matchmakingAPI = {
  join: (gameMode: string) => api.post('/matchmaking/join', { game_mode: gameMode }),
  status: () => api.get('/matchmaking/status'),
  leave: () => api.post('/matchmaking/leave'),
};

export default api;
