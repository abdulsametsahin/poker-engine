import axios from 'axios';
import { STORAGE_KEYS } from '../constants';

const API_URL = process.env.REACT_APP_API_URL || 'http://64.226.82.96:8080/api';

const api = axios.create({
  baseURL: API_URL,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem(STORAGE_KEYS.AUTH_TOKEN);
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
  getCurrentUser: () => api.get('/user'),
};

export const tableAPI = {
  getTables: () => api.get('/tables'),
  getActiveTables: () => api.get('/tables/active'),
  getPastTables: () => api.get('/tables/past'),
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
