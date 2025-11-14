# Component-Specific Socket Updates & QA Automation Plan

**Created:** 2025-11-14
**Branch:** `claude/fix-toaster-display-01DnuW4t9vv4MELjHgBpA5MS`
**Status:** Ready for Implementation

---

## Executive Summary

Based on comprehensive frontend analysis, this plan addresses:
1. **Component-specific socket updates** - Ensure every page/component receives real-time updates
2. **QA automation testing** - E2E tests for heads-up, 3-way, and 50-player tournaments

**Key Findings:**
- Only 4/6 pages use WebSocket
- 8+ socket message types are missing
- TournamentDetail inefficiently calls REST API on socket events
- Chat system not implemented
- No optimistic updates

**Timeline:** 3-4 days of focused work

---

## Part 1: Component-Specific Socket Update Refactoring

### Phase 1: Critical Fixes (Day 1 - 6-8 hours)

#### Task 1.1: Add table_id Filtering to GameView
**File:** `platform/frontend/src/pages/GameView.tsx`
**Priority:** HIGH
**Effort:** 30 minutes

**Problem:**
- GameView handlers catch ALL table_state messages from all tables
- Causes unnecessary re-renders and potential bugs

**Solution:**
```typescript
const handleTableState = (message: WSMessage<TableStatePayload>) => {
  // Filter by table_id
  if (message.payload.table_id !== tableId) {
    console.log(`Ignoring table_state for ${message.payload.table_id}, current: ${tableId}`);
    return;
  }

  // Process update
  const newState: TableState = {
    // ... existing code
  };
  setTableState(newState);
};

// Apply same pattern to all handlers:
// - handleGameUpdate
// - handleGameComplete
// - handleTournamentPaused (check tournament_id)
// - handleTournamentResumed (check tournament_id)
// - handleTournamentComplete (check tournament_id)
```

**Testing:**
1. Open two game tables in different browser tabs
2. Make action in Table A
3. Verify Table B doesn't update
4. Check console for "Ignoring" messages

---

#### Task 1.2: Implement Chat Socket System
**Files:**
- `platform/frontend/src/pages/GameView.tsx`
- `platform/frontend/src/components/game/ChatPanel.tsx`
- `platform/frontend/src/types/index.ts`

**Priority:** HIGH
**Effort:** 2-3 hours

**Step 1: Add ChatMessagePayload Type**
```typescript
// src/types/index.ts
export interface ChatMessagePayload {
  table_id: string;
  user_id: string;
  username: string;
  message: string;
  timestamp: string;
}

// Add to WSMessageType
export type WSMessageType =
  | ... // existing types
  | 'chat_message';
```

**Step 2: Add chat_message Handler in GameView**
```typescript
// src/pages/GameView.tsx
const handleChatMessage = (message: WSMessage<ChatMessagePayload>) => {
  // Only process messages for this table
  if (message.payload.table_id !== tableId) return;

  const newMessage = {
    id: Date.now().toString(),
    userId: message.payload.user_id,
    username: message.payload.username,
    message: message.payload.message,
    timestamp: new Date(message.payload.timestamp),
  };

  setChatMessages(prev => [...prev, newMessage]);
};

// Register handler
const cleanup8 = addMessageHandler('chat_message', handleChatMessage);
```

**Step 3: Update ChatPanel to Send Messages**
```typescript
// src/components/game/ChatPanel.tsx
import { useWebSocket } from '../../contexts/WebSocketContext';

const { sendMessage } = useWebSocket();

const handleSendMessage = (text: string) => {
  if (!text.trim() || !tableId || !user) return;

  // Send to server
  sendMessage({
    type: 'chat_message',
    payload: {
      table_id: tableId,
      user_id: user.id,
      username: user.username,
      message: text.trim(),
      timestamp: new Date().toISOString(),
    }
  });

  // Optimistic update (add to local state immediately)
  const newMessage = {
    id: `temp-${Date.now()}`,
    userId: user.id,
    username: user.username,
    message: text.trim(),
    timestamp: new Date(),
  };
  onSendMessage(newMessage);
};
```

**Note:** Backend must implement `chat_message` broadcast to table subscribers

---

#### Task 1.3: Add Table Event Listeners to Lobby
**File:** `platform/frontend/src/pages/Lobby.tsx`
**Priority:** HIGH
**Effort:** 3 hours

**Problem:**
- Active tables list never updates after initial load
- Users don't see new tables created
- Past games list is static

**Solution - Add 5 New Handlers:**

```typescript
// 1. Table Created
const handleTableCreated = (message: WSMessage<TableCreatedPayload>) => {
  const newTable = {
    id: message.payload.table_id,
    game_mode: message.payload.game_mode,
    small_blind: message.payload.small_blind,
    big_blind: message.payload.big_blind,
    current_players: message.payload.current_players,
    max_players: message.payload.max_players,
    status: 'waiting',
    created_at: new Date().toISOString(),
  };

  setActiveTables(prev => [newTable, ...prev]);
  showInfo(`New ${message.payload.game_mode} table created!`);
};

// 2. Table Updated
const handleTableUpdated = (message: WSMessage<TableUpdatedPayload>) => {
  setActiveTables(prev => prev.map(table =>
    table.id === message.payload.table_id
      ? { ...table, ...message.payload }
      : table
  ));
};

// 3. Table Completed
const handleTableCompleted = (message: WSMessage<TableCompletedPayload>) => {
  const completedTable = activeTables.find(t => t.id === message.payload.table_id);

  if (completedTable) {
    // Move to past tables
    setPastTables(prev => [
      {
        ...completedTable,
        status: 'completed',
        winner_id: message.payload.winner_id,
        winner_name: message.payload.winner_name,
        completed_at: new Date().toISOString(),
      },
      ...prev
    ]);

    // Remove from active
    setActiveTables(prev => prev.filter(t => t.id !== message.payload.table_id));
  }
};

// 4. Player Joined Table
const handleTablePlayerJoined = (message: WSMessage<TablePlayerJoinedPayload>) => {
  setActiveTables(prev => prev.map(table =>
    table.id === message.payload.table_id
      ? { ...table, current_players: (table.current_players || 0) + 1 }
      : table
  ));
};

// 5. Player Left Table
const handleTablePlayerLeft = (message: WSMessage<TablePlayerLeftPayload>) => {
  setActiveTables(prev => prev.map(table =>
    table.id === message.payload.table_id
      ? { ...table, current_players: Math.max(0, (table.current_players || 1) - 1) }
      : table
  ));
};

// Register all handlers
const cleanup2 = addMessageHandler('table_created', handleTableCreated);
const cleanup3 = addMessageHandler('table_updated', handleTableUpdated);
const cleanup4 = addMessageHandler('table_completed', handleTableCompleted);
const cleanup5 = addMessageHandler('table_player_joined', handleTablePlayerJoined);
const cleanup6 = addMessageHandler('table_player_left', handleTablePlayerLeft);
```

**New Payload Types Needed:**
```typescript
// src/types/index.ts
export interface TableCreatedPayload {
  table_id: string;
  game_mode: GameMode;
  small_blind: number;
  big_blind: number;
  current_players: number;
  max_players: number;
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
  reason?: string;
}
```

---

#### Task 1.4: Refactor TournamentDetail to Use Socket Payloads
**File:** `platform/frontend/src/pages/TournamentDetail.tsx`
**Priority:** HIGH
**Effort:** 2 hours

**Problem:**
- Currently calls `fetchTournamentData()` REST API on every socket event
- Inefficient: Socket event → API call → State update
- Should be: Socket event → State update

**Solution:**

```typescript
// Before:
const handlePlayerEliminated = (message: { payload: { tournament_id: string } }) => {
  if (message.payload?.tournament_id === id) {
    fetchTournamentData(); // ❌ REST API call
  }
};

// After:
const handlePlayerEliminated = (message: WSMessage<PlayerEliminatedPayload>) => {
  if (message.payload?.tournament_id !== id) return;

  // Update from socket payload directly
  setPlayers(prev => prev.map(player =>
    player.user_id === message.payload.player_id
      ? { ...player, status: 'eliminated', position: message.payload.position }
      : player
  ));

  setStandings(prev => {
    const updated = [...prev];
    updated[message.payload.position - 1] = {
      position: message.payload.position,
      player_name: message.payload.player_name,
      chips: 0,
      status: 'eliminated',
    };
    return updated;
  });

  showInfo(`${message.payload.player_name} eliminated in position ${message.payload.position}`);
};
```

**Apply Same Pattern to:**
- `handleBlindIncrease` - Update blind levels from payload
- `handleTournamentUpdate` - Update tournament from payload.tournament
- `handleTournamentComplete` - Update standings from payload.final_standings

**Update Backend Payloads:**
Ensure backend sends complete data in socket payloads:
```typescript
// Backend should send:
{
  type: 'player_eliminated',
  payload: {
    tournament_id: string,
    player_id: string,
    player_name: string,
    position: number,
    eliminated_by: string,
    updated_standings: [...],  // Include full standings
  }
}
```

---

### Phase 2: Enhancement (Day 2 - 4-6 hours)

#### Task 2.1: Add Tournament Event Listeners to Tournaments.tsx
**File:** `platform/frontend/src/pages/Tournaments.tsx`
**Priority:** MEDIUM
**Effort:** 2 hours

```typescript
// Add 3 new handlers
const handleTournamentCreated = (message: WSMessage<TournamentCreatedPayload>) => {
  const newTournament = {
    id: message.payload.tournament_id,
    name: message.payload.name,
    buy_in: message.payload.buy_in,
    starting_chips: message.payload.starting_chips,
    max_players: message.payload.max_players,
    current_players: 1, // Creator
    status: 'registering',
    created_at: new Date().toISOString(),
  };

  setTournaments(prev => [newTournament, ...prev]);
  showSuccess(`New tournament "${message.payload.name}" created!`);
};

const handleTournamentStarted = (message: WSMessage<{ tournament_id: string }>) => {
  setTournaments(prev => prev.map(t =>
    t.id === message.payload.tournament_id
      ? { ...t, status: 'in_progress' }
      : t
  ));
};

const handleTournamentPlayerRegistered = (message: WSMessage<TournamentPlayerRegisteredPayload>) => {
  setTournaments(prev => prev.map(t =>
    t.id === message.payload.tournament_id
      ? { ...t, current_players: (t.current_players || 0) + 1 }
      : t
  ));
};
```

**New Payload Types:**
```typescript
export interface TournamentCreatedPayload {
  tournament_id: string;
  name: string;
  buy_in: number;
  starting_chips: number;
  max_players: number;
  created_at: string;
}

export interface TournamentPlayerRegisteredPayload {
  tournament_id: string;
  user_id: string;
  username: string;
}
```

---

#### Task 2.2: Add Optimistic Updates for Player Actions
**File:** `platform/frontend/src/pages/GameView.tsx`
**Priority:** MEDIUM
**Effort:** 3 hours

**Current Flow:**
1. User clicks "Call" button
2. Send action to server
3. Wait for server response
4. Update UI

**New Flow (Optimistic):**
1. User clicks "Call" button
2. **Update UI immediately** (optimistic)
3. Send action to server
4. Confirm with server response (or rollback if error)

**Implementation:**
```typescript
const handleAction = useCallback((action: string, amount?: number) => {
  if (pendingAction) return;
  if (!isMyTurn) return;

  const requestId = generateRequestId();

  // OPTIMISTIC UPDATE
  setTableState(prev => {
    if (!prev) return prev;

    // Update current player's state immediately
    const updatedPlayers = prev.players.map(p =>
      p.user_id === currentUserId
        ? {
            ...p,
            last_action: action as PlayerAction,
            last_action_amount: amount,
            current_bet: action === 'call' ? prev.current_bet : amount,
            chips: p.chips - (amount || 0),
          }
        : p
    );

    return {
      ...prev,
      players: updatedPlayers,
      pot: prev.pot + (amount || 0),
    };
  });

  // Set pending (for rollback if needed)
  setPendingAction({
    type: action,
    amount,
    requestId,
    timestamp: Date.now(),
  });

  // Send to server
  sendMessage({
    type: 'game_action',
    payload: {
      table_id: tableId,
      action: action as PlayerAction,
      amount,
      request_id: requestId,
    }
  });

  // Timeout to rollback if no confirmation
  setTimeout(() => {
    if (pendingAction?.requestId === requestId) {
      console.warn('Action not confirmed, rolling back');
      // Fetch fresh state from server
      sendMessage({ type: 'subscribe_table', payload: { table_id: tableId } });
    }
  }, 5000);
}, [/* deps */]);

// In handleTableState - confirm optimistic update
const handleTableState = (message: WSMessage<TableStatePayload>) => {
  if (message.payload.table_id !== tableId) return;

  // Clear pending action if confirmed
  if (pendingAction && message.payload.action_sequence > lastActionSequence) {
    setPendingAction(null);
  }

  setTableState(message.payload);
  setLastActionSequence(message.payload.action_sequence);
};
```

---

### Phase 3: QA Automation Testing (Day 3-4)

#### Overview

Create comprehensive E2E test suite using:
- **Playwright** for browser automation
- **Test fixtures** for user/game setup
- **Mock backend** or real backend
- **Parallel execution** for speed

#### Setup Test Infrastructure

**Install Playwright:**
```bash
cd platform/frontend
npm install -D @playwright/test
npx playwright install
```

**Create Test Structure:**
```
platform/frontend/tests/
├── e2e/
│   ├── heads-up.spec.ts
│   ├── three-way.spec.ts
│   ├── tournament-50-players.spec.ts
│   └── helpers/
│       ├── auth.ts
│       ├── game.ts
│       ├── tournament.ts
│       └── websocket.ts
├── fixtures/
│   ├── users.json
│   ├── tables.json
│   └── tournaments.json
└── playwright.config.ts
```

---

#### Test Suite 1: Heads-Up Game E2E
**File:** `tests/e2e/heads-up.spec.ts`
**Scenarios:** 10 test cases
**Estimated Time:** 2-3 hours to write

```typescript
import { test, expect } from '@playwright/test';
import { loginUser, createUser } from './helpers/auth';
import { createTable, joinTable, makeAction } from './helpers/game';

test.describe('Heads-Up Game - E2E', () => {
  let player1, player2;

  test.beforeEach(async ({ browser }) => {
    // Create two players
    player1 = await browser.newPage();
    player2 = await browser.newPage();

    await createUser(player1, { username: 'player1', password: 'pass123' });
    await createUser(player2, { username: 'player2', password: 'pass123' });

    await loginUser(player1, 'player1', 'pass123');
    await loginUser(player2, 'player2', 'pass123');
  });

  test.afterEach(async () => {
    await player1.close();
    await player2.close();
  });

  test('Should complete full heads-up game', async () => {
    // Player 1 creates table
    const tableId = await createTable(player1, {
      gameMode: 'heads_up',
      buyIn: 1000,
      smallBlind: 10,
      bigBlind: 20,
    });

    // Player 2 joins
    await joinTable(player2, tableId, 1000);

    // Wait for game to start
    await expect(player1.locator('[data-testid="game-status"]')).toContainText('playing');
    await expect(player2.locator('[data-testid="game-status"]')).toContainText('playing');

    // Verify both players see same table state
    const p1Pot = await player1.locator('[data-testid="pot-amount"]').textContent();
    const p2Pot = await player2.locator('[data-testid="pot-amount"]').textContent();
    expect(p1Pot).toBe(p2Pot);

    // Play first hand
    await makeAction(player1, 'call'); // Small blind calls
    await makeAction(player2, 'check'); // Big blind checks

    // Verify pot updated
    await expect(player1.locator('[data-testid="pot-amount"]')).toContainText('30');

    // Play until game completion
    // ... continue game logic
  });

  test('Should handle player all-in', async () => {
    // ... test all-in scenario
  });

  test('Should display winner correctly', async () => {
    // ... test winner display
  });

  test('Should sync game state via WebSocket', async () => {
    // Verify real-time updates
    await makeAction(player1, 'raise', 50);

    // Player 2 should see update immediately
    await expect(player2.locator('[data-testid="current-bet"]')).toContainText('50', {
      timeout: 1000 // Should update within 1 second
    });
  });

  test('Should handle disconnection and reconnection', async () => {
    // Disconnect player 1
    await player1.context().setOffline(true);

    // Wait 2 seconds
    await player1.waitForTimeout(2000);

    // Reconnect
    await player1.context().setOffline(false);

    // Should restore game state
    await expect(player1.locator('[data-testid="game-status"]')).toContainText('playing');
  });
});
```

---

#### Test Suite 2: 3-Way Game E2E
**File:** `tests/e2e/three-way.spec.ts`
**Scenarios:** 12 test cases
**Estimated Time:** 3-4 hours

```typescript
import { test, expect } from '@playwright/test';

test.describe('3-Way Game - E2E', () => {
  let player1, player2, player3;

  test.beforeEach(async ({ browser }) => {
    // Create three players
    player1 = await browser.newPage();
    player2 = await browser.newPage();
    player3 = await browser.newPage();

    // Setup players...
  });

  test('Should complete 3-way game with multiple hands', async () => {
    // Create table
    const tableId = await createTable(player1, {
      gameMode: '3_player',
      buyIn: 1000,
      smallBlind: 10,
      bigBlind: 20,
    });

    // Join players
    await joinTable(player2, tableId, 1000);
    await joinTable(player3, tableId, 1000);

    // Verify all players see each other
    await expect(player1.locator('[data-testid="player-count"]')).toContainText('3');
    await expect(player2.locator('[data-testid="player-count"]')).toContainText('3');
    await expect(player3.locator('[data-testid="player-count"]')).toContainText('3');

    // Play hand with all three
    await makeAction(player1, 'fold');
    await makeAction(player2, 'raise', 40);
    await makeAction(player3, 'call');

    // Verify pot
    await expect(player1.locator('[data-testid="pot-amount"]')).toContainText('90');
  });

  test('Should handle player elimination in 3-way', async () => {
    // Play until one player eliminated
    // ... test logic
  });

  test('Should rotate dealer button correctly', async () => {
    // Verify dealer rotation
    // ... test logic
  });

  test('Should handle side pots correctly', async () => {
    // Test complex pot splitting
    // ... test logic
  });
});
```

---

#### Test Suite 3: 50-Player Tournament E2E
**File:** `tests/e2e/tournament-50-players.spec.ts`
**Scenarios:** 15 test cases
**Estimated Time:** 6-8 hours (complex!)

```typescript
import { test, expect } from '@playwright/test';
import { createTournament, registerForTournament } from './helpers/tournament';

test.describe('50-Player Tournament - E2E', () => {
  test('Should create and register 50 players', async ({ browser }) => {
    // Create tournament
    const admin = await browser.newPage();
    await loginUser(admin, 'admin', 'admin123');

    const tournamentId = await createTournament(admin, {
      name: '50-Player Test Tournament',
      buyIn: 100,
      startingChips: 10000,
      maxPlayers: 50,
      minPlayers: 2,
      blindLevels: [
        { small: 10, big: 20, duration: 300 },
        { small: 20, big: 40, duration: 300 },
        { small: 50, big: 100, duration: 300 },
      ],
    });

    // Create 50 players in parallel
    const players = [];
    for (let i = 1; i <= 50; i++) {
      players.push(
        browser.newPage().then(async page => {
          await createUser(page, { username: `player${i}`, password: 'pass123' });
          await loginUser(page, `player${i}`, 'pass123');
          await registerForTournament(page, tournamentId, 100);
          return page;
        })
      );
    }

    // Wait for all registrations
    await Promise.all(players);

    // Verify tournament shows 50 players
    await expect(admin.locator('[data-testid="player-count"]')).toContainText('50/50');
  });

  test('Should start tournament when all registered', async () => {
    // ... test auto-start logic
  });

  test('Should handle blind level increases', async () => {
    // Verify blind increases every 5 minutes
    // ... test logic
  });

  test('Should eliminate players correctly', async () => {
    // Track eliminations
    // Verify standings update
    // ... test logic
  });

  test('Should distribute prizes to top finishers', async () => {
    // Run full tournament
    // Verify prize distribution
    // ... test logic
  });

  test('Should handle tournament pause/resume', async () => {
    // Admin pauses tournament
    await admin.click('[data-testid="pause-tournament"]');

    // All players should see pause modal
    const player1 = players[0];
    await expect(player1.locator('[data-testid="tournament-paused-modal"]')).toBeVisible();

    // Resume
    await admin.click('[data-testid="resume-tournament"]');
    await expect(player1.locator('[data-testid="tournament-paused-modal"]')).not.toBeVisible();
  });

  test('Should maintain table balance (8-10 players per table)', async () => {
    // ... test table balancing logic
  });

  test('Should handle concurrent games across multiple tables', async () => {
    // Verify multiple tables run simultaneously
    // ... test logic
  });
});
```

---

#### Test Helpers

**File:** `tests/e2e/helpers/auth.ts`
```typescript
import { Page } from '@playwright/test';

export async function createUser(page: Page, { username, password }) {
  await page.goto('http://localhost:3000/login');
  await page.click('[data-testid="register-tab"]');
  await page.fill('[data-testid="username-input"]', username);
  await page.fill('[data-testid="email-input"]', `${username}@test.com`);
  await page.fill('[data-testid="password-input"]', password);
  await page.click('[data-testid="register-button"]');
  await page.waitForURL('**/lobby');
}

export async function loginUser(page: Page, username: string, password: string) {
  await page.goto('http://localhost:3000/login');
  await page.fill('[data-testid="username-input"]', username);
  await page.fill('[data-testid="password-input"]', password);
  await page.click('[data-testid="login-button"]');
  await page.waitForURL('**/lobby');
}
```

**File:** `tests/e2e/helpers/game.ts`
```typescript
export async function createTable(page: Page, options) {
  await page.goto('http://localhost:3000/lobby');
  await page.click('[data-testid="create-table-button"]');
  await page.selectOption('[data-testid="game-mode-select"]', options.gameMode);
  await page.fill('[data-testid="buy-in-input"]', options.buyIn.toString());
  await page.fill('[data-testid="small-blind-input"]', options.smallBlind.toString());
  await page.fill('[data-testid="big-blind-input"]', options.bigBlind.toString());
  await page.click('[data-testid="create-button"]');

  // Extract table ID from URL
  await page.waitForURL('**/game/**');
  const url = page.url();
  return url.split('/').pop();
}

export async function joinTable(page: Page, tableId: string, buyIn: number) {
  await page.goto(`http://localhost:3000/game/${tableId}`);
  await page.fill('[data-testid="buy-in-input"]', buyIn.toString());
  await page.click('[data-testid="join-button"]');
}

export async function makeAction(page: Page, action: string, amount?: number) {
  if (amount) {
    await page.fill('[data-testid="raise-amount-input"]', amount.toString());
  }
  await page.click(`[data-testid="${action}-button"]`);
}
```

---

#### Playwright Configuration

**File:** `playwright.config.ts`
```typescript
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',

  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
  ],

  webServer: {
    command: 'npm run start',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
  },
});
```

---

## Implementation Timeline

### Day 1 (6-8 hours)
- ✅ Task 1.1: Add table_id filtering (30 min)
- ✅ Task 1.2: Implement chat system (3 hrs)
- ✅ Task 1.3: Add Lobby event listeners (3 hrs)
- ✅ Task 1.4: Refactor TournamentDetail (2 hrs)

### Day 2 (4-6 hours)
- ✅ Task 2.1: Add tournament listeners (2 hrs)
- ✅ Task 2.2: Implement optimistic updates (3 hrs)
- ✅ Update all type definitions (1 hr)

### Day 3 (6-8 hours)
- ✅ Setup Playwright infrastructure (1 hr)
- ✅ Write heads-up E2E tests (3 hrs)
- ✅ Write 3-way E2E tests (3 hrs)

### Day 4 (6-8 hours)
- ✅ Write 50-player tournament tests (6 hrs)
- ✅ Run full test suite and fix issues (2 hrs)

**Total Estimated Time:** 22-30 hours (3-4 days)

---

## Backend Requirements

**New Socket Message Types Needed:**

```go
// Chat
type ChatMessage struct {
    TableID   string `json:"table_id"`
    UserID    string `json:"user_id"`
    Username  string `json:"username"`
    Message   string `json:"message"`
    Timestamp string `json:"timestamp"`
}

// Table Events
type TableCreated struct {
    TableID        string `json:"table_id"`
    GameMode       string `json:"game_mode"`
    SmallBlind     int    `json:"small_blind"`
    BigBlind       int    `json:"big_blind"`
    CurrentPlayers int    `json:"current_players"`
    MaxPlayers     int    `json:"max_players"`
}

type TablePlayerJoined struct {
    TableID  string `json:"table_id"`
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Seat     int    `json:"seat"`
}

type TableCompleted struct {
    TableID    string `json:"table_id"`
    WinnerID   string `json:"winner_id"`
    WinnerName string `json:"winner_name"`
}

// Tournament Events
type TournamentCreated struct {
    TournamentID  string `json:"tournament_id"`
    Name          string `json:"name"`
    BuyIn         int    `json:"buy_in"`
    StartingChips int    `json:"starting_chips"`
    MaxPlayers    int    `json:"max_players"`
}

type TournamentPlayerRegistered struct {
    TournamentID string `json:"tournament_id"`
    UserID       string `json:"user_id"`
    Username     string `json:"username"`
}
```

**Backend Implementation Checklist:**
- [ ] Implement chat_message broadcast to table subscribers
- [ ] Add table_created event on table creation
- [ ] Add table_player_joined/left events
- [ ] Add tournament_created event
- [ ] Add tournament_player_registered event
- [ ] Include full data in socket payloads (not just IDs)
- [ ] Implement server-side message scoping by table_id/tournament_id

---

## Testing Strategy

### Unit Tests
```bash
# Test individual components
npm test -- --testPathPattern="components/game/ChatPanel.test.tsx"
```

### Integration Tests
```bash
# Test WebSocket context
npm test -- --testPathPattern="contexts/WebSocketContext.test.tsx"
```

### E2E Tests
```bash
# Run all E2E tests
npx playwright test

# Run specific test suite
npx playwright test heads-up.spec.ts

# Run with UI
npx playwright test --ui

# Run in headed mode (see browser)
npx playwright test --headed
```

### Load Testing
```bash
# Test 50-player tournament performance
npx playwright test tournament-50-players.spec.ts --workers=1
```

---

## Success Metrics

**Component Updates:**
- ✅ All pages have real-time socket updates
- ✅ No stale data after 1 second of socket event
- ✅ Chat messages sync across all players

**QA Automation:**
- ✅ 100% test pass rate on heads-up games
- ✅ 100% test pass rate on 3-way games
- ✅ 100% test pass rate on 50-player tournaments
- ✅ Average test execution time < 5 minutes
- ✅ No flaky tests

**Performance:**
- ✅ Socket message latency < 100ms
- ✅ UI update latency < 200ms
- ✅ Zero unnecessary re-renders
- ✅ 50-player tournament runs without crashes

---

## Rollout Plan

### Week 1: Component Updates
1. Implement Phase 1 critical fixes
2. Test manually with 2-3 players
3. Deploy to staging

### Week 2: Enhancement + Testing
4. Implement Phase 2 enhancements
5. Write E2E test suites
6. Run tests in CI/CD

### Week 3: Production
7. Deploy to production
8. Monitor socket performance
9. Fix any issues

---

## Documentation Updates

**Files to Update:**
- `README.md` - Add testing instructions
- `WEBSOCKET_EVENTS.md` - Document all socket events
- `TESTING.md` - E2E testing guide
- `CONTRIBUTING.md` - How to run tests before PR

---

## Questions for Backend Team

1. Can we add these new socket message types to backend?
2. Should chat messages be stored in DB or ephemeral?
3. What's the max number of concurrent tournaments supported?
4. Do we need rate limiting on chat messages?
5. Should we implement server-side message scoping or client-side filtering?

---

**Document Status:** Ready for Implementation
**Next Steps:** Begin Day 1 tasks
**Owner:** Frontend Team
