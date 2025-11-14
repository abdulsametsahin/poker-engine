# Code Examples of Frontend Issues

## ISSUE #1: Multiple Toast Calls for Same Event

### File 1: GameView.tsx (Line 351-356)
```typescript
const handleTournamentComplete = (message: WSMessage) => {
  addConsoleLog('TOURNAMENT', `Tournament complete! Winner: ${message.payload.winner_name}`, 'success');
  showSuccess(`Tournament complete! Winner: ${message.payload.winner_name}`);  // <-- TOAST
  setTournamentId(message.payload.tournament_id);
};

// ... later ...
addMessageHandler('tournament_complete', handleTournamentComplete);
```

### File 2: TournamentDetail.tsx (Line 216-221)
```typescript
const handleTournamentComplete = (message: any) => {
  if (message.payload?.tournament_id === id) {
    fetchTournamentData();
    showSuccess(`Tournament complete! Winner: ${message.payload.winner_name}`);  // <-- SAME TOAST
  }
};

// ... later ...
addMessageHandler('tournament_complete', handleTournamentComplete);
```

**Impact:** If user navigates from Lobby to GameView to TournamentDetail, the last handler registered wins (overwrites previous). This is unpredictable and creates silent failures.

---

## ISSUE #2: Untyped Message Payloads

### File 1: Tournaments.tsx (Lines 110-120)
```typescript
const handleTournamentPaused = (message: any) => {  // ❌ Should be strongly typed
  const tournamentId = message.payload?.tournament_id;
  if (tournamentId) {
    setTournaments(prev => prev.map(t => 
      t.id === tournamentId 
        ? { ...t, status: 'paused' }
        : t
    ));
  }
};
```

**No type safety for:**
- `message.payload` structure
- Properties accessed on `message.payload`
- Return types

### File 2: Lobby.tsx (Line 200-207)
```typescript
const handler = (message: any) => {  // ❌ Any type
  const { table_id } = message.payload;  // ❌ No type checking
  const gameMode = matchmaking?.gameMode || 'heads_up';

  setMatchmaking(null);
  setMatchFound({ tableId: table_id, gameMode });  // ❌ Could receive wrong property name
  showSuccess('Match found!');
};
```

**Problems:**
1. Backend could rename `table_id` to `tableId` and no IDE warning
2. Typos in payload access not caught at compile time
3. No documentation of expected payload shape

---

## ISSUE #3: Duplicate Socket Handlers Overwrite Each Other

### Problem in WebSocketContext.tsx (Lines 152-158)
```typescript
const addMessageHandler = useCallback((type: string, handler: MessageHandler) => {
  messageHandlersRef.current.set(type, handler);  // ❌ Overwrites previous handler
}, []);

const removeMessageHandler = useCallback((type: string) => {
  messageHandlersRef.current.delete(type);
}, []);
```

Uses `Map.set()` which replaces the value instead of adding to an array.

### GameView.tsx Registers Multiple Handlers
```typescript
// Lines 337-356
const handleTournamentPaused = (message: WSMessage) => { /* ... */ };
const handleTournamentResumed = (message: WSMessage) => { /* ... */ };
const handleTournamentComplete = (message: WSMessage) => { /* ... */ };

// Lines 362-364
addMessageHandler('tournament_paused', handleTournamentPaused);
addMessageHandler('tournament_resumed', handleTournamentResumed);
addMessageHandler('tournament_complete', handleTournamentComplete);
```

### TournamentDetail.tsx Registers SAME Handlers
```typescript
// Lines 176-221
const handleTournamentPaused = (message: any) => { /* ... */ };
const handleTournamentResumed = (message: any) => { /* ... */ };
const handleTournamentComplete = (message: any) => { /* ... */ };

// Lines 225-229
addMessageHandler('tournament_paused', handleTournamentPaused);   // ❌ OVERWRITES GameView
addMessageHandler('tournament_resumed', handleTournamentResumed); // ❌ OVERWRITES GameView
addMessageHandler('tournament_complete', handleTournamentComplete); // ❌ OVERWRITES GameView
```

### Race Condition Example
```
Time 1: User loads GameView
  └─ Registers tournament_paused handler from GameView
     messageHandlersRef = { tournament_paused: GameViewHandler }

Time 2: User navigates to TournamentDetail
  └─ Registers tournament_paused handler from TournamentDetail
     messageHandlersRef = { tournament_paused: TournamentDetailHandler }  // ❌ GameViewHandler lost!

Time 3: tournament_paused message arrives
  └─ Only TournamentDetailHandler executes
  └─ GameView never gets notified
  └─ Game state never updates
```

---

## ISSUE #4: Deprecated Hook Implementation (Dead Code)

### File: /src/hooks/useWebSocket.ts (47 lines)
```typescript
import { useEffect, useRef, useState } from 'react';

const WS_URL = process.env.REACT_APP_WS_URL || 'ws://localhost:8080/ws';

export interface WSMessage {
  type: string;
  payload: any;  // ❌ Uses 'any'
}

export const useWebSocket = () => {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<WSMessage | null>(null);
  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('token');  // ❌ Wrong! Should use AuthContext
    if (!token) return;

    ws.current = new WebSocket(`${WS_URL}?token=${token}`);

    ws.current.onopen = () => {
      setIsConnected(true);
    };

    ws.current.onmessage = (event) => {
      const message = JSON.parse(event.data);
      setLastMessage(message);
    };

    ws.current.onclose = () => {
      setIsConnected(false);
    };

    return () => {
      ws.current?.close();
    };
  }, []);  // ❌ Missing dependencies

  const send = (message: WSMessage) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(message));
    }
  };

  return { isConnected, lastMessage, send };
};
```

**Why This is Bad:**
1. Not used anywhere (found via grep search - zero imports)
2. Reads token from localStorage instead of AuthContext
3. No reconnection logic (unlike WebSocketContext)
4. No heartbeat mechanism
5. Missing dependency in useEffect (missing ws.current?)
6. Conflicts with actual WebSocketContext.tsx useWebSocket hook name (name collision!)

---

## ISSUE #5: GameView.tsx - Component Too Large

### File: /src/pages/GameView.tsx (1,155 lines)

**13 useState declarations:**
```typescript
const [tableState, setTableState] = useState<TableState | null>(null);           // Game state
const [raiseAmount, setRaiseAmount] = useState(0);                               // UI state
const [showHandComplete, setShowHandComplete] = useState(false);                 // Modal state
const [showGameComplete, setShowGameComplete] = useState(false);                 // Modal state
const [gameMode, setGameMode] = useState<string>('heads_up');                   // Game config
const [consoleOpen, setConsoleOpen] = useState(false);                           // Debug UI
const [consoleLogs, setConsoleLogs] = useState<any[]>([]);                      // Debug data
const [tournamentId, setTournamentId] = useState<string | null>(null);          // Context data
const [pendingAction, setPendingAction] = useState<PendingAction | null>(null); // Action state
const [lastActionSequence, setLastActionSequence] = useState<number>(0);        // Sync state
const [history, setHistory] = useState<any[]>([...localStorage]);               // Game history
const [chatMessages, setChatMessages] = useState<any[]>([]);                    // Chat state
```

**Plus 7 message handlers:**
```typescript
const handleTableState = (message: WSMessage) => { /* 60+ lines */ };
const handleGameUpdate = (message: WSMessage) => { /* calls handleTableState */ };
const handleGameComplete = (message: WSMessage) => { /* 20+ lines */ };
const handleError = (message: WSMessage) => { /* 1 line */ };
const handleTournamentPaused = (message: WSMessage) => { /* 5 lines */ };
const handleTournamentResumed = (message: WSMessage) => { /* 5 lines */ };
const handleTournamentComplete = (message: WSMessage) => { /* 5 lines */ };
```

**Should be split into:**
1. `GameStateManager` - handleTableState, handleGameUpdate, tableState
2. `PlayerActionsHandler` - handleAction, pendingAction, raiseAmount
3. `TournamentListener` - tournament-related handlers, tournamentId
4. `DebugConsole` - consoleLogs, consoleOpen, addConsoleLog
5. `ModalController` - showHandComplete, showGameComplete

---

## ISSUE #6: Stale Data Fallback in AuthContext

### File: /src/contexts/AuthContext.tsx (Lines 36-76)
```typescript
useEffect(() => {
  const initAuth = async () => {
    const storedToken = getStorageItem(STORAGE_KEYS.AUTH_TOKEN);

    if (storedToken) {
      setToken(storedToken);

      // Try to fetch fresh user data from server
      try {
        const response = await authAPI.getCurrentUser();
        const userData = response.data;
        setUser(userData);  // ✓ Fresh data
        setStorageItem(STORAGE_KEYS.USER_DATA, JSON.stringify(userData));
      } catch (error) {
        // If fetch fails, use stored data as fallback
        const storedUserData = getStorageItem(STORAGE_KEYS.USER_DATA);
        if (storedUserData) {
          try {
            const userData = JSON.parse(storedUserData);
            setUser(userData);  // ❌ STALE data - no timestamp!
          } catch (parseError) {
            // If parsing fails, fall back to basic user info
            const storedUserId = getStorageItem(STORAGE_KEYS.USER_ID);
            const storedUsername = getStorageItem(STORAGE_KEYS.USERNAME);
            if (storedUserId && storedUsername) {
              setUser({
                id: storedUserId,
                username: storedUsername,  // ❌ STALE username!
              });
            }
          }
        }
      }
    }

    setIsLoading(false);
  };

  initAuth();
}, []);
```

**Problems:**
1. No timestamp on stored user data
2. Can't detect if cached data is stale (could be from 1 day ago)
3. If server has updated user balance/status, user sees old data
4. User won't know they're viewing stale information
5. Could lead to race conditions if user data changes frequently

**Better Approach:**
```typescript
// Store with timestamp
const userData = {
  ...response.data,
  _storedAt: Date.now()
};

// Check if older than 5 minutes
const isStale = (Date.now() - userData._storedAt) > 5 * 60 * 1000;
```

---

## ISSUE #7: Missing Type Definitions for WebSocket Payloads

### File: /src/types/index.ts (Lines 114-127)
```typescript
// Current - Generic with any default
export interface WSMessage<T = any> {  // ❌ Defaults to 'any'
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

// Missing all these:
// export interface MatchFoundPayload { ... }
// export interface TableStatePayload { ... }
// export interface GameCompletePayload { ... }
// export interface TournamentPausedPayload { ... }
// etc...
```

### How It Should Look
```typescript
// ✓ Specific interfaces for each payload
export interface MatchFoundPayload {
  table_id: string;
  game_mode: GameMode;
}

export interface TournamentPausedPayload {
  tournament_id: string;
  reason?: string;
}

export interface GameCompletePayload {
  winner_id: string;
  final_chip_count: number;
}

// ✓ Discriminated union for type-safe dispatch
export type WSMessageEvent = 
  | WSMessage<'match_found', MatchFoundPayload>
  | WSMessage<'tournament_paused', TournamentPausedPayload>
  | WSMessage<'game_complete', GameCompletePayload>;
```

Then handlers would be:
```typescript
// ✓ Type-safe!
const handler = (message: WSMessage<'match_found', MatchFoundPayload>) => {
  const { table_id } = message.payload;  // ✓ Compiler knows this exists
};
```

---

## ISSUE #8: Card Type Inconsistency

### File: /src/types/index.ts (Lines 2-26)
```typescript
export interface Player {
  cards?: Card[] | string[];  // ❌ Can be EITHER!
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
```

### Utility Function Must Handle Both
File: /src/utils/index.ts (Lines 4-52)
```typescript
export const parseCard = (card: string | Card): CardObject => {
  // Handle object format
  if (typeof card === 'object' && 'rank' in card && 'suit' in card) {
    return card as CardObject;
  }

  const cardStr = card as string;

  // Handle "Ah", "Kh", "10d" formats
  if (cardStr.length >= 2) {
    const rank = cardStr.slice(0, -1);
    const suitChar = cardStr.slice(-1);
    // ... 15 more lines of parsing
  }

  // Handle "A♠", "K♥" formats
  if (cardStr.includes('♠') || /* ... */) {
    // ... 5 more lines
  }

  // Handle "ace_of_spades" formats
  const parts = cardStr.split('_');
  if (parts.length === 3 && parts[1] === 'of') {
    // ... 10 more lines
  }

  return { rank: '', suit: '', display: cardStr };
};
```

**Better Approach:**
```typescript
// Use a discriminated union
export type Card = 
  | CardObject  // { rank, suit, display }
  | CardString; // "Ah"

// Then parseCard is simple:
export const parseCard = (card: Card): CardObject => {
  if (typeof card === 'object') {
    return card;
  }
  // Just parse the string
};
```

---

## Summary: Which Issues Affect Toasters?

1. ✓ **Issue #1** - Multiple toast calls (direct)
2. ✓ **Issue #2** - Untyped handlers make toast data risky
3. ✓ **Issue #3** - Duplicate handlers affect toast delivery
4. ✗ Issue #4 - Deprecated hook (not used)
5. ✗ Issue #5 - Component size (architectural)
6. ✓ **Issue #6** - Stale data shown in toasts (user sees old names)
7. ✓ **Issue #7** - Missing types makes toast payload unreliable
8. ✗ Issue #8 - Card types (display only)

**Top 3 to Fix:**
1. Fix handler registry to allow multiple handlers per event
2. Add payload type definitions and type all handlers
3. Implement toast deduplication for same event

