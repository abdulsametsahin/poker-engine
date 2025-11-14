# Frontend Codebase Analysis

## Directory Structure

```
/home/user/poker-engine/platform/frontend/
├── src/
│   ├── components/
│   │   ├── common/        # Reusable UI components
│   │   ├── game/          # Game-specific components
│   │   └── modals/        # Modal dialogs
│   ├── contexts/          # React Context (Auth, Toast, WebSocket)
│   ├── hooks/             # Custom React hooks
│   ├── pages/             # Page-level components
│   ├── services/          # API calls (api.ts)
│   ├── types/             # TypeScript type definitions
│   ├── utils/             # Utility functions
│   ├── App.tsx            # Main App component
│   └── index.tsx          # Entry point
├── public/
├── package.json
└── tsconfig.json
```

**Key Stats:**
- 61 TypeScript files total
- Largest file: GameView.tsx (1,155 lines)
- Total frontend code: ~12,700 lines

---

## 1. TOASTER IMPLEMENTATION ANALYSIS

### Libraries Used
- **Material-UI (MUI) Snackbar + Alert**: No external toast library like react-toastify or sonner
- Location: `/home/user/poker-engine/platform/frontend/src/contexts/ToastContext.tsx`

### Toast System Architecture

**ToastContext.tsx** (119 lines):
```typescript
interface ToastContextType {
  showToast: (message: string, type?: AlertColor) => void;
  showSuccess: (message: string) => void;
  showError: (message: string) => void;
  showWarning: (message: string) => void;
  showInfo: (message: string) => void;
}
```

**Key Features:**
- Queue-based toast system (FIFO) - handles one toast at a time
- Uses MUI Snackbar with custom Alert styling
- Duration: 2000ms (TOAST_DURATION from constants)
- Position: Top center with slide transition

### Toast Trigger Points

Found 50+ toast calls across the application:

| File | Location | Count | Context |
|------|----------|-------|---------|
| GameView.tsx | Pages | 6 | Game completion, tournament events |
| Lobby.tsx | Pages | 6 | Matchmaking, table joins |
| Login.tsx | Pages | 8 | Authentication, validation |
| TournamentDetail.tsx | Pages | 10 | Tournament management |
| Tournaments.tsx | Pages | 2 | Tournament listing |

### ISSUE #1: Multiple Toast Calls for Same Events

**Problem:** Tournament completion triggers multiple toasts in different files

In GameView.tsx (line 353):
```typescript
const handleTournamentComplete = (message: WSMessage) => {
  showSuccess(`Tournament complete! Winner: ${message.payload.winner_name}`);
};
```

In TournamentDetail.tsx (line 219):
```typescript
const handleTournamentComplete = (message: any) => {
  showSuccess(`Tournament complete! Winner: ${message.payload.winner_name}`);
};
```

Both files listen to the same 'tournament_complete' WebSocket event and show toasts. If a user is on both screens simultaneously (unlikely but possible), duplicate toasts appear.

### ISSUE #2: Untyped Message Payloads in Toast Handlers

Multiple handlers use `any` type:

**Tournaments.tsx (lines 110-134):**
```typescript
const handleTournamentPaused = (message: any) => {  // ❌ 'any' type
  const tournamentId = message.payload?.tournament_id;
};

const handleTournamentResumed = (message: any) => {  // ❌ 'any' type
  const tournamentId = message.payload?.tournament_id;
};
```

**TournamentDetail.tsx (lines 160-220):**
```typescript
const handleTournamentUpdate = (message: any) => {  // ❌ 'any' type
const handleTournamentStarted = (message: any) => {  // ❌ 'any' type
const handleTournamentPaused = (message: any) => {   // ❌ 'any' type
```

**Lobby.tsx (line 200):**
```typescript
const handler = (message: any) => {  // ❌ 'any' type
  const { table_id } = message.payload;
};
```

---

## 2. SOCKET IMPLEMENTATION ANALYSIS

### Socket Architecture

**Location:** `/home/user/poker-engine/platform/frontend/src/contexts/WebSocketContext.tsx` (187 lines)

Uses native WebSocket API with custom provider pattern.

### Socket Setup Flow

```
App.tsx
  └─ AuthProvider
      └─ ToastProvider
          └─ WebSocketProvider ← Custom context
              └─ All Routes
```

### WebSocket Connection Details

**Connection Code (lines 65-126):**
```typescript
const connect = useCallback(() => {
  if (!isAuthenticated || !token) return;

  const wsUrl = API.BASE_URL.replace('http://', 'ws://').replace('https://', 'wss://').replace('/api', '');
  const ws = new WebSocket(`${wsUrl}/ws?token=${token}`);

  ws.onopen = () => {
    setIsConnected(true);
    reconnectAttemptRef.current = 0;
    startHeartbeat();
  };

  ws.onclose = () => {
    setIsConnected(false);
    clearHeartbeat();
    // Reconnect logic
  };

  ws.onmessage = (event) => {
    const message: WSMessage = JSON.parse(event.data);
    setLastMessage(message);
    const handler = messageHandlersRef.current.get(message.type);
    if (handler) {
      handler(message);
    }
  };
}, [isAuthenticated, token, ...]);
```

**Key Features:**
- Token-based authentication (via query parameter)
- Exponential backoff reconnection (1.5x multiplier, max 30s)
- Heartbeat pings every 25 seconds
- Event-driven message dispatch using Map-based handler registry

### Message Handler Registration Pattern

**Dynamic Handler System:**

Each page/component registers handlers in useEffect:

```typescript
useEffect(() => {
  const handleTableState = (message: WSMessage) => { /* ... */ };
  const handleGameComplete = (message: WSMessage) => { /* ... */ };
  
  addMessageHandler('table_state', handleTableState);
  addMessageHandler('game_complete', handleGameComplete);
  
  return () => {
    removeMessageHandler('table_state');
    removeMessageHandler('game_complete');
  };
}, [addMessageHandler, removeMessageHandler, /* deps */]);
```

### Socket Event Listeners Across App

| File | Event Type | Count | Notes |
|------|-----------|-------|-------|
| GameView.tsx | table_state, game_update, game_complete, error, tournament_* | 7 | Most complex handler |
| TournamentDetail.tsx | tournament_*, blind_level_increased, player_eliminated | 7 | Real-time tournament updates |
| Tournaments.tsx | tournament_paused, tournament_resumed, tournament_update | 3 | Tournament list updates |
| Lobby.tsx | match_found | 1 | Matchmaking found event |
| MultiTableView.tsx | table_state, game_update | 2 | Multi-table display |

### ISSUE #3: Duplicate Socket Handlers for Same Events

**Problem:** 'tournament_paused', 'tournament_resumed', and 'tournament_complete' handled in multiple places

**TournamentDetail.tsx (lines 223-239):**
```typescript
addMessageHandler('tournament_paused', handleTournamentPaused);
addMessageHandler('tournament_resumed', handleTournamentResumed);
addMessageHandler('tournament_complete', handleTournamentComplete);
```

**GameView.tsx (lines 362-364):**
```typescript
addMessageHandler('tournament_paused', handleTournamentPaused);
addMessageHandler('tournament_resumed', handleTournamentResumed);
addMessageHandler('tournament_complete', handleTournamentComplete);
```

The WebSocket handler registry uses `Map.set()`, which means the second registration overwrites the first. This is problematic because:
1. Only the last registered handler executes
2. Creates unpredictable behavior based on which component mounts last
3. Makes debugging difficult

### ISSUE #4: Deprecated Hook Implementation

**Location:** `/home/user/poker-engine/platform/frontend/src/hooks/useWebSocket.ts` (47 lines)

A SECOND, simpler WebSocket implementation exists and is duplicated:

```typescript
const WS_URL = process.env.REACT_APP_WS_URL || 'ws://localhost:8080/ws';

export interface WSMessage {
  type: string;
  payload: any;  // ❌ 'any' type
}

export const useWebSocket = () => {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<WSMessage | null>(null);
  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('token');  // ❌ Reads from localStorage directly
    if (!token) return;

    ws.current = new WebSocket(`${WS_URL}?token=${token}`);
    // Basic setup without heartbeat or reconnection logic
  }, []);

  return { isConnected, lastMessage, send };
};
```

**Problems:**
1. NOT imported anywhere currently (dead code)
2. Reads token from localStorage instead of Auth context
3. Lacks reconnection logic
4. Missing heartbeat mechanism
5. No exported payload types
6. TypeScript interface uses `any` for payload

### ISSUE #5: Missing Reconnection in Some Handlers

**GameView.tsx:** Has 7 message handlers but no cleanup error handling
**Lobby.tsx:** Single match_found handler with no error fallback

---

## 3. DATA MANAGEMENT ANALYSIS

### State Management Architecture

**No centralized state management** (no Redux, Zustand, etc.)

Uses composition of:
- **Context API:** Auth, Toast, WebSocket
- **Local Component State:** useState
- **localStorage:** Token persistence

### GameView.tsx State Bloat

**13 useState declarations (lines 48-72):**
```typescript
const [tableState, setTableState] = useState<TableState | null>(null);
const [raiseAmount, setRaiseAmount] = useState(0);
const [showHandComplete, setShowHandComplete] = useState(false);
const [showGameComplete, setShowGameComplete] = useState(false);
const [gameMode, setGameMode] = useState<string>('heads_up');
const [consoleOpen, setConsoleOpen] = useState(false);
const [consoleLogs, setConsoleLogs] = useState<any[]>([]);
const [tournamentId, setTournamentId] = useState<string | null>(null);
const [pendingAction, setPendingAction] = useState<PendingAction | null>(null);
const [lastActionSequence, setLastActionSequence] = useState<number>(0);
const [history, setHistory] = useState<any[]>([...]);
const [chatMessages, setChatMessages] = useState<any[]>([]);
```

**File Size:** 1,155 lines (component too large)

### ISSUE #6: Data Flow Without Clear Patterns

**Stale Data Scenarios:**

1. **User data in multiple places:**
   - AuthContext: user + token
   - localStorage: AUTH_TOKEN, USER_ID, USERNAME, USER_DATA
   - Passed as props in some places

2. **Tournament data inconsistency:**
   - TournamentDetail.tsx calls `fetchTournamentData()` multiple times
   - Tournaments.tsx has separate tournament list
   - Multiple socket handlers update state independently

3. **Table state management:**
   - tableState lives in GameView
   - Historical game data stored in localStorage with `game_history_${tableId}`
   - No cache invalidation strategy

### ISSUE #7: localStorage Used for Session Data

**AuthContext.tsx (lines 39-70):**
```typescript
const initAuth = async () => {
  const storedToken = getStorageItem(STORAGE_KEYS.AUTH_TOKEN);  // localStorage
  
  if (storedToken) {
    try {
      const response = await authAPI.getCurrentUser();  // Fresh data
      setUser(response.data);
    } catch (error) {
      const storedUserData = getStorageItem(STORAGE_KEYS.USER_DATA);  // Fallback
      if (storedUserData) {
        setUser(JSON.parse(storedUserData));  // Might be stale
      }
    }
  }
};
```

**Problems:**
- Fallback to stale stored user data if API fails
- No timestamp on stored data to detect staleness
- User can see old data if connection drops

---

## 4. TYPESCRIPT TYPE ISSUES

### Type Definition File

**Location:** `/home/user/poker-engine/platform/frontend/src/types/index.ts` (168 lines)

### ISSUE #8: Loose Type for WSMessage Generic

**types/index.ts (line 114):**
```typescript
export interface WSMessage<T = any> {  // ❌ Default 'any' type
  type: WSMessageType;
  payload: T;
}
```

### ISSUE #9: Improperly Typed Payload Functions

All message handlers use `any`:

**TournamentDetail.tsx (lines 160-220):**
```typescript
const handleTournamentUpdate = (message: any) => {  // Should be WSMessage<TournamentUpdatePayload>
  if (message.payload?.tournament?.id === id) { }
};
```

**Lobby.tsx (line 200):**
```typescript
const handler = (message: any) => {  // Should be WSMessage<MatchFoundPayload>
  const { table_id } = message.payload;
};
```

### Missing Type Definitions

**Should exist but don't:**
```typescript
// Missing:
export interface TournamentUpdatePayload { }
export interface TournamentStartedPayload { }
export interface TournamentPausedPayload { }
export interface MatchFoundPayload { }  // Only partial definition
export interface TournamentCompletePayload { }
export interface BlindIncreasePayload { }
export interface PlayerEliminatedPayload { }
export interface ErrorPayload { }
export interface TableStatePayload { }
export interface GameUpdatePayload { }
export interface GameCompletePayload { }
```

### ISSUE #10: Inconsistent Card Type Handling

**types/index.ts (lines 8, 17-26):**
```typescript
export interface Player {
  cards?: Card[] | string[];  // ❌ Can be string OR Card object
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

Cards are sometimes `"Ah"` (string), sometimes `{rank: 'A', suit: '♠'}` (object). Utility functions have to handle both:

**utils/index.ts (lines 4-52):**
```typescript
export const parseCard = (card: string | Card): CardObject => {
  if (typeof card === 'object' && 'rank' in card && 'suit' in card) {
    return card as CardObject;
  }
  // ... handle string formats
};
```

---

## Summary of Critical Issues

### Severity Breakdown

**HIGH Priority (Bugs/Data Loss):**
1. ✗ Duplicate socket handlers override each other (Issue #3)
2. ✗ Stale user data fallback in AuthContext (Issue #7)
3. ✗ Multiple toasts for same event not deduplicated (Issue #1)

**MEDIUM Priority (Code Quality):**
4. ✗ GameView.tsx is 1,155 lines - needs decomposition (Issue #6)
5. ✗ Multiple `any` types in message handlers (Issues #2, #9)
6. ✗ Duplicate WebSocket hook implementation (Issue #4)

**LOW Priority (Best Practices):**
7. ✗ No centralized state management
8. ✗ Inconsistent Card type handling (Issue #10)
9. ✗ Missing payload type definitions (Issue #9)
10. ✗ localStorage used for session state (Issue #7)

---

## Recommendations

### Short Term (Bugs)
1. Fix handler registration to allow multiple handlers per event
2. Add deduplication for toast notifications
3. Remove deprecated useWebSocket hook
4. Properly type all message handlers

### Medium Term (Architecture)
1. Extract GameView into smaller, focused components
2. Create proper type definitions for all WebSocket payloads
3. Consider using a state management library (Redux, Zustand, or Recoil)
4. Implement message handler middleware for logging/debugging

### Long Term (Patterns)
1. Move session storage to memory, not localStorage
2. Implement data cache invalidation strategy
3. Create custom hooks for domain logic (tournaments, tables, matches)
4. Add error boundaries and fallback UI for disconnections

