# Key Frontend Files Reference

## Critical Files by Purpose

### Context Providers (App-level State)
| File | Lines | Purpose | Issues |
|------|-------|---------|--------|
| `/src/contexts/AuthContext.tsx` | 127 | User auth & token mgmt | Stale data fallback (Issue #7) |
| `/src/contexts/ToastContext.tsx` | 119 | Toast notifications | Queue-based (single at a time) |
| `/src/contexts/WebSocketContext.tsx` | 187 | WebSocket management | Handler registry overwrites (Issue #3) |

### Custom Hooks
| File | Lines | Purpose | Issues |
|------|-------|---------|--------|
| `/src/hooks/useWebSocket.ts` | 47 | **DEPRECATED** - Old WS hook | Dead code, uses localStorage (Issue #4) |
| `/src/hooks/useResponsiveSize.ts` | TBD | Responsive sizing | - |

### Services
| File | Lines | Purpose | Details |
|------|-------|---------|---------|
| `/src/services/api.ts` | 63 | HTTP API client | Auth, tables, matchmaking, tournaments |

### Type Definitions
| File | Lines | Issues |
|------|-------|--------|
| `/src/types/index.ts` | 168 | `any` types in WSMessage, missing payload types |

### Utilities
| File | Lines | Purpose |
|------|-------|---------|
| `/src/utils/index.ts` | 306 | Cards, formatting, validation, storage helpers |

### Pages (Large Components)
| File | Lines | Socket Handlers | useState Count | Issues |
|------|-------|-----------------|----------------|--------|
| `/src/pages/GameView.tsx` | 1,155 | 7 | 13 | Bloated, multiple handlers (Issue #6, #3) |
| `/src/pages/TournamentDetail.tsx` | 895 | 7 | Multiple | Duplicate handlers (Issue #3) |
| `/src/pages/Lobby.tsx` | 620 | 1 | Multiple | Untyped handler (Issue #2) |
| `/src/pages/Tournaments.tsx` | 432 | 3 | Multiple | Untyped handlers (Issue #2) |
| `/src/pages/Login.tsx` | 451 | 0 | Multiple | Validation toasts |
| `/src/pages/Settings.tsx` | 194 | 0 | - | - |

### Components
| File | Lines | Connections |
|------|-------|-------------|
| `/src/components/MultiTableView.tsx` | 344 | Uses WebSocket & Router |
| `/src/components/game/PokerTable.tsx` | 416 | Core game rendering |
| `/src/components/game/PlayerSeat.tsx` | 378 | Player UI rendering |
| `/src/components/common/Button.tsx` | 264 | Button variants |
| `/src/components/modals/HandCompleteDisplay.tsx` | 253 | Hand result modal |
| `/src/components/modals/WinnerDisplay.tsx` | 278 | Winner announcement |
| `/src/components/modals/GameCompleteDisplay.tsx` | 312 | Game end modal |
| `/src/components/game/ConsolePanel.tsx` | 222 | Debug logging UI |

---

## Problem-to-File Mapping

### Issue #1: Multiple Toast Calls
- `/src/pages/GameView.tsx` - line 353
- `/src/pages/TournamentDetail.tsx` - line 219
- Both handle 'tournament_complete' event

### Issue #2: Untyped Message Payloads
- `/src/pages/Tournaments.tsx` - lines 110-134 (all `any`)
- `/src/pages/TournamentDetail.tsx` - lines 160-220 (all `any`)
- `/src/pages/Lobby.tsx` - line 200 (all `any`)

### Issue #3: Duplicate Socket Handlers
- `/src/pages/GameView.tsx` - lines 358-364 (registers 7 handlers)
- `/src/pages/TournamentDetail.tsx` - lines 223-239 (registers 7 handlers)
- `/src/pages/Tournaments.tsx` - lines 145-147 (registers 3 handlers)
- `/src/components/MultiTableView.tsx` - lines 62-63 (registers 2 handlers)

### Issue #4: Deprecated Hook
- `/src/hooks/useWebSocket.ts` - entire file (47 lines)
- NOT imported anywhere (dead code)

### Issue #6: Bloated Component
- `/src/pages/GameView.tsx` - 1,155 lines total
- Should be split into: GameState, PlayerActions, Tournament, Console subcomponents

### Issue #7: Stale Data
- `/src/contexts/AuthContext.tsx` - lines 50-56 (fallback to localStorage)
- `/src/services/api.ts` - no cache headers

### Issue #8: Loose TypeScript
- `/src/types/index.ts` - line 114 (WSMessage<T = any>)

### Issue #9: Missing Types
- `/src/types/index.ts` - missing all payload interfaces
- Affects all handler functions across pages

### Issue #10: Card Type Inconsistency
- `/src/types/index.ts` - lines 8, 22-26 (Card[] | string[])
- `/src/utils/index.ts` - lines 4-52 (parseCard handles both)

---

## Data Flow Diagram

```
Entry Point
  ↓
App.tsx
  ↓
Auth (localStorage) → AuthProvider
  ↓
ToastProvider (MUI Snackbar)
  ↓
WebSocketProvider (native WebSocket)
  ├─ GameView.tsx ←─ Socket Events → WebSocketContext
  │   ├─ useState[13] → local state
  │   ├─ localStorage[game_history_*]
  │   └─ showToast() → ToastContext
  │
  ├─ TournamentDetail.tsx ←─ Socket Events
  │   ├─ fetchTournamentData() → API
  │   └─ showToast() → ToastContext
  │
  ├─ Tournaments.tsx ←─ Socket Events
  │   ├─ fetchTournaments() → API
  │   └─ showToast() → ToastContext
  │
  └─ Lobby.tsx ←─ Socket Events
      ├─ matchmakingAPI.join() → API
      └─ showToast() → ToastContext
```

---

## Dependencies Analysis

### External Dependencies
```json
{
  "@mui/material": "^7.3.5",      // UI components & Snackbar
  "@mui/icons-material": "^7.3.5", // Icons
  "react": "^19.2.0",             // Core framework
  "react-router-dom": "^6.28.0",  // Routing
  "axios": "^1.13.2",             // HTTP client
  "typescript": "^4.9.5"          // Type checking
}
```

### No External Toast Library
- Custom implementation using MUI Snackbar
- Could use react-toastify or sonner for better UX

### No State Management Library
- No Redux, Zustand, Recoil, or similar
- Only Context API + useState
- Leads to prop drilling and state inconsistency

---

## File Size Distribution

```
GameView.tsx .......... 1,155 lines (30.8%)
TournamentDetail.tsx ... 895 lines (23.9%)
Lobby.tsx ............. 620 lines (16.6%)
PokerTable.tsx ........ 416 lines (11.1%)
components total ...... 3,500+ lines
pages total ........... 3,700+ lines
Total ................. ~12,700 lines
```

**Analysis:** GameView is 30% of the entire frontend by itself!

