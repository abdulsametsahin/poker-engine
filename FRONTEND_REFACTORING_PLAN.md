# Frontend Comprehensive Refactoring Plan

**Created:** 2025-11-14
**Branch:** `claude/fix-toaster-display-01DnuW4t9vv4MELjHgBpA5MS`
**Status:** Ready for Implementation

---

## Executive Summary

This plan addresses critical bugs in the poker engine frontend:
- **Toaster showing multiple times** due to duplicate handlers
- **Stale data** visible without reloading
- **Socket handler conflicts** causing data loss
- **Type safety issues** making refactoring unsafe

The refactoring is organized into 4 phases over 4 weeks, prioritizing critical bugs first.

---

## Critical Issues Identified

### HIGH Priority (Must Fix - Week 1)
1. **Socket handlers override each other** (`/src/contexts/WebSocketContext.tsx:152-158`)
   - Race condition: last registered handler wins
   - Silent data loss when multiple components subscribe
   - **Impact:** GameView stops receiving updates when TournamentDetail mounts

2. **Multiple toast calls for same event** (`GameView.tsx:353`, `TournamentDetail.tsx:219`)
   - Duplicate notifications confuse users
   - **Impact:** User sees "Tournament complete!" 2-3 times

3. **Stale user data fallback** (`/src/contexts/AuthContext.tsx:50-56`)
   - No timestamp on localStorage data
   - User sees outdated balance/status
   - **Impact:** User confusion, data inconsistency

### MEDIUM Priority (Should Fix - Week 2-3)
4. **Untyped message handlers** (12+ instances across pages)
   - All handlers use `any` type
   - Backend changes go undetected
   - **Impact:** Runtime errors, unsafe refactoring

5. **Deprecated WebSocket hook** (`/src/hooks/useWebSocket.ts`)
   - 47 lines of dead code
   - Name collision with real implementation
   - **Impact:** Developer confusion

6. **GameView.tsx too large** (1,155 lines)
   - 13 useState declarations
   - 7 message handlers
   - **Impact:** Hard to maintain/test

7. **Missing WebSocket payload types** (`/src/types/index.ts`)
   - No type definitions for 8+ payload types
   - **Impact:** Type safety defeated

### LOW Priority (Nice to Fix - Week 4+)
8. Card type inconsistency (string vs object)
9. No state management library
10. localStorage used for session state

---

## Phase 1: Critical Bug Fixes (Week 1)

**Goal:** Fix data loss and duplicate toasts

### Task 1.1: Fix Socket Handler Registry
**File:** `/src/contexts/WebSocketContext.tsx`
**Lines:** 152-158
**Priority:** CRITICAL

**Current Problem:**
```typescript
const addMessageHandler = useCallback((type: string, handler: MessageHandler) => {
  messageHandlersRef.current.set(type, handler);  // ❌ Overwrites previous handler
}, []);
```

**Solution:**
```typescript
// Change from Map<string, MessageHandler> to Map<string, MessageHandler[]>
const messageHandlersRef = useRef<Map<string, MessageHandler[]>>(new Map());

const addMessageHandler = useCallback((type: string, handler: MessageHandler) => {
  const handlers = messageHandlersRef.current.get(type) || [];
  handlers.push(handler);
  messageHandlersRef.current.set(type, handlers);

  // Return cleanup function
  return () => {
    const currentHandlers = messageHandlersRef.current.get(type) || [];
    const filtered = currentHandlers.filter(h => h !== handler);
    if (filtered.length === 0) {
      messageHandlersRef.current.delete(type);
    } else {
      messageHandlersRef.current.set(type, filtered);
    }
  };
}, []);

// Update message dispatch
const handleMessage = (data: any) => {
  const handlers = messageHandlersRef.current.get(data.type) || [];
  handlers.forEach(handler => {
    try {
      handler(data);
    } catch (error) {
      console.error(`Error in handler for ${data.type}:`, error);
    }
  });
};
```

**Testing:**
1. Open GameView, register handlers
2. Navigate to TournamentDetail
3. Trigger `tournament_paused` event
4. Verify BOTH components update correctly

**Affected Files:**
- `/src/contexts/WebSocketContext.tsx` (modify handler registry)
- `/src/pages/GameView.tsx` (update to use cleanup function)
- `/src/pages/TournamentDetail.tsx` (update to use cleanup function)
- `/src/pages/Tournaments.tsx` (update to use cleanup function)
- `/src/pages/Lobby.tsx` (update to use cleanup function)

---

### Task 1.2: Implement Toast Deduplication
**File:** `/src/contexts/ToastContext.tsx`
**Lines:** 30-40
**Priority:** CRITICAL

**Current Problem:**
- Multiple components show same toast independently
- No deduplication mechanism

**Solution:**
```typescript
interface ToastMessage {
  id: string;
  message: string;
  severity: AlertColor;
  timestamp: number;
}

// Add deduplication with time window
const recentToasts = useRef<Set<string>>(new Set());

const showToast = (message: string, severity: AlertColor) => {
  // Create hash of message + severity
  const toastKey = `${severity}:${message}`;

  // Check if shown recently (within 2 seconds)
  if (recentToasts.current.has(toastKey)) {
    console.log(`Toast deduplicated: ${message}`);
    return;
  }

  // Add to queue
  setToasts(prev => [...prev, {
    id: Date.now().toString(),
    message,
    severity,
    timestamp: Date.now()
  }]);

  // Mark as shown
  recentToasts.current.add(toastKey);

  // Remove from dedup set after 2 seconds
  setTimeout(() => {
    recentToasts.current.delete(toastKey);
  }, 2000);
};
```

**Testing:**
1. Trigger `tournament_complete` with multiple components mounted
2. Verify only ONE toast appears
3. Verify toast appears again if triggered after 2 seconds

**Affected Files:**
- `/src/contexts/ToastContext.tsx` (add deduplication logic)

---

### Task 1.3: Fix Stale Data in AuthContext
**File:** `/src/contexts/AuthContext.tsx`
**Lines:** 50-56
**Priority:** HIGH

**Current Problem:**
```typescript
const storedUserData = getStorageItem(STORAGE_KEYS.USER_DATA);
if (storedUserData) {
  const userData = JSON.parse(storedUserData);
  setUser(userData);  // ❌ No timestamp check!
}
```

**Solution:**
```typescript
interface StoredUserData {
  user: User;
  _storedAt: number;
}

// When storing
const dataToStore: StoredUserData = {
  user: userData,
  _storedAt: Date.now()
};
setStorageItem(STORAGE_KEYS.USER_DATA, JSON.stringify(dataToStore));

// When retrieving
const storedUserData = getStorageItem(STORAGE_KEYS.USER_DATA);
if (storedUserData) {
  const parsed: StoredUserData = JSON.parse(storedUserData);
  const age = Date.now() - parsed._storedAt;
  const FIVE_MINUTES = 5 * 60 * 1000;

  if (age < FIVE_MINUTES) {
    setUser(parsed.user);
  } else {
    console.warn('Stored user data is stale, forcing fresh fetch');
    // Show warning to user
    showWarning('Loading fresh data...');
    // Could retry API call or force logout
  }
}
```

**Testing:**
1. Login and store user data
2. Manually set `_storedAt` to old timestamp
3. Refresh page
4. Verify warning shown and fresh data fetched

**Affected Files:**
- `/src/contexts/AuthContext.tsx` (add timestamp checks)
- `/src/types/index.ts` (add StoredUserData interface)

---

### Task 1.4: Remove Deprecated WebSocket Hook
**File:** `/src/hooks/useWebSocket.ts`
**Priority:** MEDIUM (Quick win)

**Action:**
- Delete entire file (47 lines of dead code)
- Verify no imports exist (already confirmed via grep)

**Testing:**
- Run build: `npm run build`
- Verify no import errors
- Search codebase: `grep -r "from.*hooks/useWebSocket"`

**Affected Files:**
- DELETE: `/src/hooks/useWebSocket.ts`

---

## Phase 2: Type Safety (Week 2)

**Goal:** Add comprehensive type definitions to prevent runtime errors

### Task 2.1: Create WebSocket Payload Type Definitions
**File:** `/src/types/index.ts`
**Priority:** MEDIUM

**Add Missing Payload Interfaces:**
```typescript
// Tournament Event Payloads
export interface TournamentPausedPayload {
  tournament_id: string;
  reason?: string;
  paused_at: string;
}

export interface TournamentResumedPayload {
  tournament_id: string;
  resumed_at: string;
}

export interface TournamentCompletePayload {
  tournament_id: string;
  winner_id: string;
  winner_name: string;
  final_standings: Array<{
    player_id: string;
    player_name: string;
    position: number;
    prize: number;
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

// Game Event Payloads
export interface MatchFoundPayload {
  table_id: string;
  game_mode: GameMode;
  players: string[];
}

export interface TableStatePayload {
  table_id: string;
  game_state: GameState;
  players: Player[];
  pot: number;
  community_cards: Card[];
  current_player_id?: string;
  action_timeout?: number;
}

export interface GameUpdatePayload extends TableStatePayload {
  last_action?: PlayerAction;
}

export interface GameCompletePayload {
  table_id: string;
  winner_id: string;
  winner_name: string;
  final_chip_count: number;
  hand_history: HandHistory;
}

export interface PlayerActionPayload {
  table_id: string;
  player_id: string;
  action: ActionType;
  amount?: number;
  timestamp: string;
}

export interface ErrorPayload {
  code: string;
  message: string;
  details?: any;
}

// Update WSMessage to use discriminated union
export type WSMessageEvent =
  | { type: 'tournament_paused'; payload: TournamentPausedPayload }
  | { type: 'tournament_resumed'; payload: TournamentResumedPayload }
  | { type: 'tournament_complete'; payload: TournamentCompletePayload }
  | { type: 'player_eliminated'; payload: PlayerEliminatedPayload }
  | { type: 'blind_level_increased'; payload: BlindLevelIncreasedPayload }
  | { type: 'match_found'; payload: MatchFoundPayload }
  | { type: 'table_state'; payload: TableStatePayload }
  | { type: 'game_update'; payload: GameUpdatePayload }
  | { type: 'game_complete'; payload: GameCompletePayload }
  | { type: 'player_action'; payload: PlayerActionPayload }
  | { type: 'error'; payload: ErrorPayload };
```

**Affected Files:**
- `/src/types/index.ts` (add all payload interfaces)

---

### Task 2.2: Type All Message Handlers
**Priority:** MEDIUM

**Update Handler Signatures:**

**File:** `/src/pages/GameView.tsx`
```typescript
// Before
const handleTournamentPaused = (message: WSMessage) => { ... }

// After
const handleTournamentPaused = (message: { type: 'tournament_paused'; payload: TournamentPausedPayload }) => {
  const { tournament_id, reason } = message.payload;  // ✓ Type-safe!
  // ...
}
```

**Files to Update:**
1. `/src/pages/GameView.tsx` - 7 handlers
2. `/src/pages/TournamentDetail.tsx` - 5 handlers
3. `/src/pages/Tournaments.tsx` - 3 handlers
4. `/src/pages/Lobby.tsx` - 2 handlers
5. `/src/components/MultiTableView.tsx` - 2 handlers

**Testing:**
1. Rename a payload property in types
2. Verify TypeScript errors in handlers
3. Update handlers
4. Verify build passes

---

### Task 2.3: Update WebSocketContext to Use Typed Messages
**File:** `/src/contexts/WebSocketContext.tsx`
**Priority:** MEDIUM

**Update Context Interface:**
```typescript
interface WebSocketContextType {
  isConnected: boolean;
  sendMessage: (message: any) => void;
  addMessageHandler: <T extends WSMessageEvent['type']>(
    type: T,
    handler: (message: Extract<WSMessageEvent, { type: T }>) => void
  ) => () => void;
  removeMessageHandler: (type: string) => void;
}
```

**Affected Files:**
- `/src/contexts/WebSocketContext.tsx` (update types)

---

## Phase 3: Component Refactoring (Week 3)

**Goal:** Break down large components, improve maintainability

### Task 3.1: Split GameView.tsx into Smaller Components

**Current:** 1,155 lines, 13 useState declarations

**Target Architecture:**
```
GameView/
  ├── index.tsx (main component, 200 lines)
  ├── GameStateManager.tsx (handle game state updates)
  ├── PlayerActionsPanel.tsx (action buttons, raise slider)
  ├── TournamentListener.tsx (tournament event handlers)
  ├── DebugConsole.tsx (console logs, debug UI)
  └── GameModals.tsx (hand complete, game complete modals)
```

**Step 1: Create Component Structure**
```bash
mkdir -p src/pages/GameView
```

**Step 2: Extract DebugConsole (Easiest)**
```typescript
// src/pages/GameView/DebugConsole.tsx
interface DebugConsoleProps {
  logs: ConsoleLog[];
  isOpen: boolean;
  onToggle: () => void;
}

export const DebugConsole: React.FC<DebugConsoleProps> = ({ logs, isOpen, onToggle }) => {
  // Move console rendering logic here
};
```

**Step 3: Extract TournamentListener**
```typescript
// src/pages/GameView/TournamentListener.tsx
interface TournamentListenerProps {
  tournamentId: string | null;
  onTournamentUpdate: (id: string) => void;
}

export const TournamentListener: React.FC<TournamentListenerProps> = ({
  tournamentId,
  onTournamentUpdate
}) => {
  const { addMessageHandler } = useWebSocket();
  const { showSuccess, showInfo } = useToast();

  // Move tournament handlers here
  useEffect(() => {
    const cleanup1 = addMessageHandler('tournament_paused', handlePaused);
    const cleanup2 = addMessageHandler('tournament_resumed', handleResumed);
    const cleanup3 = addMessageHandler('tournament_complete', handleComplete);

    return () => {
      cleanup1();
      cleanup2();
      cleanup3();
    };
  }, [tournamentId]);

  return null; // Invisible component
};
```

**Step 4: Extract GameStateManager**
**Step 5: Extract PlayerActionsPanel**
**Step 6: Extract GameModals**

**Testing:**
- Load game, verify all functionality works
- Check network tab for duplicate API calls
- Verify no memory leaks (mount/unmount 10 times)

**Affected Files:**
- `/src/pages/GameView.tsx` → `/src/pages/GameView/index.tsx`
- CREATE: `/src/pages/GameView/GameStateManager.tsx`
- CREATE: `/src/pages/GameView/PlayerActionsPanel.tsx`
- CREATE: `/src/pages/GameView/TournamentListener.tsx`
- CREATE: `/src/pages/GameView/DebugConsole.tsx`
- CREATE: `/src/pages/GameView/GameModals.tsx`

---

### Task 3.2: Implement Proper Cache Invalidation
**Priority:** MEDIUM

**Problem:**
- Tournament data fetched in multiple places
- No coordination between fetches
- Stale data visible

**Solution - Add Timestamp to Fetched Data:**
```typescript
interface CachedData<T> {
  data: T;
  fetchedAt: number;
  expiresAt: number;
}

const tournamentCache = new Map<string, CachedData<Tournament>>();

const fetchTournamentData = async (id: string, forceRefresh = false) => {
  const cached = tournamentCache.get(id);

  if (!forceRefresh && cached && Date.now() < cached.expiresAt) {
    return cached.data;
  }

  const response = await tournamentAPI.getTournament(id);
  const data = response.data;

  tournamentCache.set(id, {
    data,
    fetchedAt: Date.now(),
    expiresAt: Date.now() + (5 * 60 * 1000) // 5 minutes
  });

  return data;
};

// Invalidate on socket updates
addMessageHandler('tournament_paused', (message) => {
  tournamentCache.delete(message.payload.tournament_id);
});
```

**Alternative - Create TournamentContext:**
```typescript
// src/contexts/TournamentContext.tsx
interface TournamentContextType {
  tournaments: Map<string, Tournament>;
  fetchTournament: (id: string) => Promise<Tournament>;
  invalidateTournament: (id: string) => void;
}
```

**Affected Files:**
- CREATE: `/src/contexts/TournamentContext.tsx` OR
- `/src/pages/TournamentDetail.tsx` (add caching logic)
- `/src/pages/Tournaments.tsx` (use shared cache)

---

## Phase 4: Architecture Improvements (Week 4+)

**Goal:** Long-term maintainability and scalability

### Task 4.1: Fix Card Type Inconsistency
**File:** `/src/types/index.ts`
**Priority:** LOW

**Current:**
```typescript
export interface Player {
  cards?: Card[] | string[];  // ❌ Ambiguous
}
```

**Solution:**
```typescript
export interface Card {
  rank: string;
  suit: string;
}

export type CardRepresentation = Card | string;

export interface Player {
  cards?: Card[];  // Always use Card objects internally
}

// Parser handles conversion
export const parseCard = (input: CardRepresentation): Card => {
  if (typeof input === 'object') return input;
  // Parse string format
};
```

**Affected Files:**
- `/src/types/index.ts` (update Card types)
- `/src/utils/index.ts` (simplify parseCard)

---

### Task 4.2: Consider State Management Library
**Priority:** LOW (Research phase)

**Current Architecture:**
- Context API + useState
- Props drilling in some places
- State duplicated across components

**Options:**
1. **Zustand** (Recommended)
   - Lightweight (3KB)
   - Simple API
   - No boilerplate
   - Good TypeScript support

2. **Redux Toolkit**
   - More boilerplate
   - Better for very large apps
   - Excellent DevTools

3. **Jotai** (Atomic state)
   - Very lightweight
   - Bottom-up approach
   - Great for derived state

**Example with Zustand:**
```typescript
// src/store/gameStore.ts
import create from 'zustand';

interface GameState {
  tableState: TableState | null;
  tournamentId: string | null;
  setTableState: (state: TableState) => void;
  setTournamentId: (id: string) => void;
}

export const useGameStore = create<GameState>((set) => ({
  tableState: null,
  tournamentId: null,
  setTableState: (state) => set({ tableState: state }),
  setTournamentId: (id) => set({ tournamentId: id }),
}));
```

**Decision Point:**
- If Context API works well after refactoring → Don't add library
- If still experiencing state issues → Add Zustand

**Affected Files:**
- TBD based on decision

---

### Task 4.3: Add Error Boundaries
**Priority:** LOW

**Problem:**
- Unhandled errors crash entire app
- No graceful degradation

**Solution:**
```typescript
// src/components/ErrorBoundary.tsx
class ErrorBoundary extends React.Component<Props, State> {
  state = { hasError: false, error: null };

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error boundary caught:', error, errorInfo);
    // Log to error tracking service
  }

  render() {
    if (this.state.hasError) {
      return <ErrorFallback error={this.state.error} />;
    }
    return this.props.children;
  }
}
```

**Affected Files:**
- CREATE: `/src/components/ErrorBoundary.tsx`
- `/src/App.tsx` (wrap routes)

---

### Task 4.4: Improve localStorage Usage
**Priority:** LOW

**Current Issues:**
- Session state in localStorage (persists across sessions)
- No size limits
- No encryption for sensitive data

**Solutions:**
1. Use sessionStorage for temporary data
2. Add size monitoring
3. Encrypt sensitive data
4. Add expiration timestamps

**Affected Files:**
- `/src/utils/storage.ts` (enhance storage utilities)

---

## Testing Strategy

### Unit Tests
```bash
# Test socket handler registry
npm test -- WebSocketContext.test.tsx

# Test toast deduplication
npm test -- ToastContext.test.tsx

# Test type safety
npm run type-check
```

### Integration Tests
```bash
# Test multi-component socket handling
npm test -- integration/socket-handlers.test.tsx

# Test stale data detection
npm test -- integration/auth-context.test.tsx
```

### Manual Testing Checklist
- [ ] Load GameView, then TournamentDetail - both receive socket messages
- [ ] Trigger tournament_complete - only ONE toast appears
- [ ] Logout, set old timestamp, login - warning shown
- [ ] Navigate between pages - no duplicate toasts
- [ ] Reconnect socket - handlers still work
- [ ] TypeScript build passes with no `any` warnings

---

## Migration Strategy

### Week 1: Critical Fixes
**Branch:** `claude/fix-toaster-display-01DnuW4t9vv4MELjHgBpA5MS`

1. Fix socket handler registry (Task 1.1)
2. Implement toast deduplication (Task 1.2)
3. Fix stale data (Task 1.3)
4. Remove deprecated hook (Task 1.4)

**Deliverables:**
- No more duplicate toasts
- No more handler overwrites
- Stale data warnings shown

### Week 2: Type Safety
**Branch:** `feature/type-safety-improvements`

1. Create payload type definitions (Task 2.1)
2. Type all message handlers (Task 2.2)
3. Update WebSocketContext types (Task 2.3)

**Deliverables:**
- All handlers strongly typed
- No `any` types in handlers
- TypeScript strict mode enabled

### Week 3: Component Refactoring
**Branch:** `feature/gameview-refactor`

1. Split GameView.tsx (Task 3.1)
2. Implement cache invalidation (Task 3.2)

**Deliverables:**
- GameView.tsx < 300 lines
- 5 focused subcomponents
- Proper cache invalidation

### Week 4+: Architecture
**Branches:** TBD

1. Fix card types (Task 4.1)
2. Research state management (Task 4.2)
3. Add error boundaries (Task 4.3)
4. Improve localStorage (Task 4.4)

**Deliverables:**
- Consistent type system
- Error handling
- Production-ready architecture

---

## Risk Assessment

### High Risk
- **Socket handler registry change** - Core functionality, affects all pages
  - Mitigation: Thorough testing, gradual rollout
  - Rollback plan: Revert commit, use feature flag

### Medium Risk
- **Type system changes** - May break existing code
  - Mitigation: TypeScript will catch issues at compile time
  - Rollback plan: Types are non-breaking, can be gradual

### Low Risk
- **Component splitting** - Isolated changes
  - Mitigation: Component tests
  - Rollback plan: Easy to revert

---

## Success Metrics

### Performance
- [ ] Page load time < 2s
- [ ] Socket message handling < 50ms
- [ ] No memory leaks (constant memory over 10 minutes)

### Code Quality
- [ ] TypeScript strict mode enabled
- [ ] 0 `any` types in handlers
- [ ] Test coverage > 70%
- [ ] Bundle size < 500KB

### User Experience
- [ ] 0 duplicate toasts
- [ ] 0 stale data complaints
- [ ] < 1% error rate in production

---

## Next Steps

1. **Review this plan** with team (30 minutes)
2. **Approve Week 1 tasks** for immediate implementation
3. **Create GitHub issues** for each task
4. **Set up feature branches** for each phase
5. **Begin implementation** of Phase 1

---

## Questions for Team

1. Do we have a staging environment for testing socket changes?
2. What is our deployment process for frontend changes?
3. Should we add feature flags for gradual rollout?
4. Do we have error tracking service (Sentry, etc.)?
5. What is our browser support requirement (affects state management choice)?

---

## References

- [FRONTEND_EXPLORATION_SUMMARY.txt](./FRONTEND_EXPLORATION_SUMMARY.txt)
- [FRONTEND_ANALYSIS.md](./FRONTEND_ANALYSIS.md)
- [FRONTEND_CODE_ISSUES.md](./FRONTEND_CODE_ISSUES.md)
- [FRONTEND_FILES_REFERENCE.md](./FRONTEND_FILES_REFERENCE.md)

---

**Document Status:** Ready for Review
**Next Review:** After Phase 1 completion
**Owner:** Development Team
