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
export type PlayerAction = 'fold' | 'check' | 'call' | 'raise' | 'allin';

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
  | 'tournament_paused'
  | 'tournament_resumed'
  | 'tournament_complete'
  | 'player_eliminated'
  | 'blind_level_increased'
  | 'chat_message'
  | 'table_created'
  | 'table_updated'
  | 'table_completed'
  | 'table_player_joined'
  | 'table_player_left'
  | 'tournament_created'
  | 'tournament_started'
  | 'tournament_player_registered'
  | 'tournament_player_unregistered'
  | 'error';

// WebSocket payload type definitions
export interface MatchFoundPayload {
  table_id: string;
  game_mode?: GameMode;
  players?: string[];
}

export interface PlayerActionPayload {
  action: PlayerAction;
  amount?: number;
}

export interface TableStatePayload extends TableState {
  // TableState already has all needed properties
}

export interface GameUpdatePayload extends TableState {
  last_action?: {
    player_id: string;
    action: PlayerAction;
    amount?: number;
  };
}

export interface GameCompletePayload {
  table_id?: string;
  winner_id: string;
  winner_name?: string;
  winner_username?: string;
  final_chip_count: number;
  players_defeated?: number;
  total_hands?: number;
  biggest_pot?: number;
  best_hand?: string;
}

export interface TournamentPausedPayload {
  tournament_id: string;
  reason?: string;
  paused_at?: string;
}

export interface TournamentResumedPayload {
  tournament_id: string;
  resumed_at?: string;
}

export interface TournamentCompletePayload {
  tournament_id: string;
  winner_id: string;
  winner_name: string;
  final_standings?: Array<{
    player_id: string;
    player_name: string;
    position: number;
    prize?: number;
  }>;
}

export interface PlayerEliminatedPayload {
  tournament_id: string;
  player_id: string;
  player_name: string;
  position: number;
  eliminated_by?: string;
}

export interface BlindLevelIncreasedPayload {
  tournament_id: string;
  level: number;
  small_blind: number;
  big_blind: number;
  ante?: number;
}

export interface ErrorPayload {
  code?: string;
  message: string;
  details?: any;
}

// Chat payload
export interface ChatMessagePayload {
  table_id: string;
  user_id: string;
  username: string;
  message: string;
  timestamp: string;
}

// Table event payloads
export interface TableCreatedPayload {
  table_id: string;
  game_mode: GameMode;
  small_blind: number;
  big_blind: number;
  current_players: number;
  max_players: number;
  status: string;
  created_at: string;
}

export interface TableUpdatedPayload {
  table_id: string;
  status?: string;
  current_players?: number;
}

export interface TableCompletedPayload {
  table_id: string;
  winner_id: string;
  winner_name: string;
  completed_at: string;
}

export interface TablePlayerJoinedPayload {
  table_id: string;
  user_id: string;
  username: string;
  seat: number;
}

export interface TablePlayerLeftPayload {
  table_id: string;
  user_id: string;
  username: string;
  reason?: string;
}

// Tournament event payloads
export interface TournamentCreatedPayload {
  tournament_id: string;
  name: string;
  buy_in: number;
  starting_chips: number;
  max_players: number;
  min_players: number;
  status: string;
  created_at: string;
}

export interface TournamentStartedPayload {
  tournament_id: string;
  started_at: string;
}

export interface TournamentPlayerRegisteredPayload {
  tournament_id: string;
  user_id: string;
  username: string;
}

export interface TournamentPlayerUnregisteredPayload {
  tournament_id: string;
  user_id: string;
  username: string;
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
