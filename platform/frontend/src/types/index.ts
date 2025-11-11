// Core game types
export interface Player {
  user_id: string;
  username?: string;
  seat_number: number;
  chips: number;
  current_bet: number;
  cards?: Card[] | string[]; // Can be Card objects or strings like "Ah", "Kd"
  folded: boolean;
  all_in: boolean;
  is_dealer: boolean;
  is_active: boolean;
  last_action?: PlayerAction;
  last_action_amount?: number;
}

export interface Card {
  rank: string;
  suit: string;
}

export interface CardObject {
  rank: string;
  suit: string;
  display?: string;
}

export interface TableState {
  table_id: string;
  table_name?: string;
  game_mode: GameMode;
  status: TableStatus;
  players: Player[];
  community_cards: Card[];
  pot: number;
  current_bet: number;
  current_player_seat?: number;
  dealer_seat: number;
  small_blind: number;
  big_blind: number;
  betting_round: BettingRound;
  min_raise?: number;
  hand_number?: number;
  created_at?: string;
  completed_at?: string;
  total_hands?: number;
}

export interface WinnerInfo {
  playerId: string;
  username?: string;
  amount: number;
  handRank: string;
  handCards: Card[];
}

export interface GameComplete {
  winner_id: string;
  winner_username?: string;
  final_chip_count: number;
  players_defeated: number;
  total_hands: number;
  biggest_pot?: number;
  best_hand?: string;
}

// Enums
export type GameMode = 'heads_up' | '3_player' | '6_player';
export type TableStatus = 'waiting' | 'playing' | 'completed';
export type BettingRound = 'preflop' | 'flop' | 'turn' | 'river';
export type PlayerAction = 'fold' | 'check' | 'call' | 'raise' | 'all_in';

// API types
export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user_id: string;
  username: string;
}

export interface CreateTableRequest {
  game_mode: GameMode;
  buy_in: number;
  small_blind: number;
  big_blind: number;
}

export interface JoinTableRequest {
  buy_in: number;
}

export interface MatchmakingRequest {
  game_mode: GameMode;
}

export interface MatchmakingStatus {
  in_queue: boolean;
  game_mode?: GameMode;
  players_in_queue: number;
  players_needed: number;
}

// WebSocket message types
export interface WSMessage<T = any> {
  type: WSMessageType;
  payload: T;
}

export type WSMessageType =
  | 'subscribe_table'
  | 'game_action'
  | 'match_found'
  | 'table_state'
  | 'game_update'
  | 'game_complete'
  | 'player_action'
  | 'error';

export interface MatchFoundPayload {
  table_id: string;
}

export interface PlayerActionPayload {
  action: PlayerAction;
  amount?: number;
}

// UI types
export interface User {
  id: string;
  username: string;
  email?: string;
  chips?: number;
  games_played?: number;
  games_won?: number;
  created_at?: string;
}

export interface ToastMessage {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  message: string;
  duration?: number;
}

// Component props types
export interface ButtonVariant {
  variant: 'primary' | 'secondary' | 'danger' | 'ghost' | 'success';
}

export interface AvatarSize {
  size: 'small' | 'medium' | 'large';
}

export interface CardSize {
  size: 'small' | 'medium' | 'large';
}
