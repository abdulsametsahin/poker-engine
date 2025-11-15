# History Tracking System Refactor - Comprehensive Plan

## Executive Summary

This plan addresses critical issues with action synchronization and history tracking in the poker platform for both cash and tournament games.

### Current Issues
1. **"Processing..." UI stuck**: Frontend shows "Processing call/check/raise..." even after action is processed
2. **Actions not visible to other players**: Player actions saved to DB but not broadcast in real-time
3. **Incomplete history**: Only player actions tracked, missing hand lifecycle events (cards dealt, flop/turn/river, showdown, winners)
4. **No persistent history UI**: History stored only in localStorage, not synced from server
5. **No hand history API**: Cannot fetch past hands or detailed action history

---

## Part 1: Root Cause Analysis

### Issue 1: "Processing..." Synchronization Problem

**Location**: `/home/user/poker-engine/platform/frontend/src/pages/GameView.tsx:787`

**Current Behavior**:
```typescript
// Line 64: pendingAction state set when user acts
setPendingAction({ type: action, timestamp: Date.now() })

// Lines 213-222: Cleared only when:
// 1. Turn changes away from us, OR
// 2. action_sequence advances while we still have turn
if (pendingAction) {
  if (currentTurn !== currentUserId) {
    setPendingAction(null);
  } else if (newState.action_sequence > lastActionSequence) {
    setPendingAction(null);
  }
}
```

**Problem**:
- Race condition: action_sequence may not increment fast enough
- In some edge cases (multiple quick actions, network latency), neither condition triggers
- No explicit acknowledgment from server that action was processed

### Issue 2: Actions Not Visible to Other Players

**Current Flow**:
1. Player sends action via WebSocket: `game_action` event
2. Backend processes in `ProcessGameAction()` (events.go:188)
3. Action saved to DB (events.go:256-268)
4. Backend broadcasts `game_update` with full state
5. **Problem**: `game_update` includes only `last_action` on player object, no dedicated action event

**What's Missing**:
- No separate "player_action_confirmed" event
- Other players only see action via `last_action` field changes
- History panel relies on detecting `last_action` changes (fragile)

### Issue 3: Incomplete Event Tracking

**Currently Tracked** (in `hand_actions` table):
- Player actions: fold, check, call, raise, allin
- Betting round: preflop, flop, turn, river
- Amount

**NOT Tracked**:
- Hand started event
- Cards dealt to players (private - only visible to player)
- Community cards revealed (flop, turn, river)
- Showdown event
- Winner determination
- Hand completion
- Pot transfers
- Player timeouts/auto-actions

---

## Part 2: Solution Architecture

### Database Schema Changes

#### New Table: `game_events`
Tracks ALL game events for complete hand history reconstruction.

```sql
CREATE TABLE game_events (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  hand_id BIGINT NOT NULL,
  table_id VARCHAR(36) NOT NULL,
  event_type ENUM(
    'hand_started',
    'cards_dealt',
    'blinds_posted',
    'player_action',
    'round_advanced',
    'showdown',
    'hand_complete',
    'player_timeout'
  ) NOT NULL,
  user_id VARCHAR(36),  -- NULL for table-wide events
  betting_round ENUM('preflop', 'flop', 'turn', 'river'),
  action_type VARCHAR(20),  -- fold, check, call, raise, allin
  amount INT DEFAULT 0,
  metadata JSON,  -- Flexible storage for event-specific data
  sequence_number INT NOT NULL,  -- Order of events within hand
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_hand (hand_id),
  INDEX idx_table (table_id, created_at),
  INDEX idx_sequence (hand_id, sequence_number),

  FOREIGN KEY (hand_id) REFERENCES hands(id) ON DELETE CASCADE
);
```

**Metadata JSON Examples**:
```json
// hand_started
{
  "dealer_position": 0,
  "small_blind_position": 1,
  "big_blind_position": 2,
  "small_blind_amount": 10,
  "big_blind_amount": 20
}

// player_action
{
  "player_name": "John",
  "current_bet": 100,
  "pot_after": 450
}

// round_advanced
{
  "new_round": "flop",
  "community_cards": ["Ah", "Kd", "Qs"],
  "pot": 450
}

// showdown
{
  "players_showing": [
    {"user_id": "123", "hand": ["As", "Ks"], "hand_rank": "Pair of Aces"}
  ]
}

// hand_complete
{
  "winners": [
    {"user_id": "123", "amount": 450, "hand_rank": "Pair of Aces"}
  ],
  "final_pot": 450,
  "final_community_cards": ["Ah", "Kd", "Qs", "Jc", "Tc"]
}
```

#### Enhanced `hands` Table
Add fields for better querying and display.

```sql
ALTER TABLE hands
  ADD COLUMN betting_rounds_reached ENUM('preflop', 'flop', 'turn', 'river', 'showdown') DEFAULT 'preflop',
  ADD COLUMN num_players INT DEFAULT 0,
  ADD COLUMN hand_summary TEXT;  -- Human-readable summary for UI
```

#### Keep `hand_actions` Table
Maintain backward compatibility and faster queries for action-only history.

---

### Backend Changes

#### 1. Event Tracking Service
**New File**: `/platform/backend/internal/server/history/tracker.go`

```go
type HistoryTracker struct {
  db *db.DB
  mu sync.RWMutex
  handSequences map[int64]int  // hand_id -> next sequence number
}

func (h *HistoryTracker) RecordEvent(
  handID int64,
  tableID string,
  eventType string,
  userID *string,
  bettingRound *string,
  actionType *string,
  amount int,
  metadata map[string]interface{},
) error {
  // Get next sequence number for this hand
  seq := h.getNextSequence(handID)

  // Marshal metadata to JSON
  metadataJSON, _ := json.Marshal(metadata)

  // Insert event
  event := models.GameEvent{
    HandID:         handID,
    TableID:        tableID,
    EventType:      eventType,
    UserID:         userID,
    BettingRound:   bettingRound,
    ActionType:     actionType,
    Amount:         amount,
    Metadata:       string(metadataJSON),
    SequenceNumber: seq,
  }

  return h.db.Create(&event).Error
}
```

#### 2. Enhanced Event Handler
**Modify**: `/platform/backend/internal/server/events/events.go`

Add event recording for ALL engine events:

```go
func HandleGameEvent(event pokerModels.Event, database *db.DB, bridge *game.GameBridge, broadcast func(string), historyTracker *history.HistoryTracker) {
  tableID := event.TableID

  switch event.Event {
  case "handStart":
    game.CreateHandRecord(bridge, database, tableID, event)

    // NEW: Record hand_started event
    handID, _ := bridge.GetCurrentHandID(tableID)
    data := event.Data.(map[string]interface{})
    historyTracker.RecordEvent(handID, tableID, "hand_started", nil, nil, nil, 0, data)

  case "playerAction":
    // Already saves to hand_actions table
    // NEW: Also save to game_events for complete history
    data := event.Data.(map[string]interface{})
    handID, _ := bridge.GetCurrentHandID(tableID)
    userID := data["playerId"].(string)
    action := data["action"].(string)
    amount := data["amount"].(int)
    bettingRound := getCurrentBettingRound(bridge, tableID)

    metadata := map[string]interface{}{
      "player_name": data["playerName"],
      "current_bet": data["currentBet"],
      "pot_after": data["potAfter"],
    }

    historyTracker.RecordEvent(
      handID, tableID, "player_action",
      &userID, &bettingRound, &action, amount, metadata,
    )

    // NEW: Broadcast action confirmation event to ALL players
    broadcastActionConfirmation(bridge, tableID, data)

  case "roundAdvanced":
    handID, _ := bridge.GetCurrentHandID(tableID)
    data := event.Data.(map[string]interface{})
    newRound := data["round"].(string)

    metadata := map[string]interface{}{
      "new_round": newRound,
      "community_cards": data["communityCards"],
      "pot": data["pot"],
    }

    historyTracker.RecordEvent(
      handID, tableID, "round_advanced",
      nil, &newRound, nil, 0, metadata,
    )

  case "handComplete":
    game.UpdateHandRecord(bridge, database, tableID, event)

    // NEW: Record showdown and hand_complete events
    handID, _ := bridge.GetCurrentHandID(tableID)
    data := event.Data.(map[string]interface{})

    // Record showdown if players showed cards
    if showdownData, ok := data["showdown"]; ok {
      historyTracker.RecordEvent(
        handID, tableID, "showdown",
        nil, nil, nil, 0, showdownData.(map[string]interface{}),
      )
    }

    // Record hand completion
    historyTracker.RecordEvent(
      handID, tableID, "hand_complete",
      nil, nil, nil, 0, data,
    )

  // ... other cases
  }

  broadcast(tableID)
}
```

#### 3. Action Confirmation Broadcast
**New Function**: Send explicit confirmation to action initiator

```go
func broadcastActionConfirmation(bridge *game.GameBridge, tableID string, actionData map[string]interface{}) {
  userID := actionData["playerId"].(string)

  // Send confirmation to the player who acted
  confirmMsg := map[string]interface{}{
    "type": "action_confirmed",
    "payload": map[string]interface{}{
      "user_id": userID,
      "action": actionData["action"],
      "amount": actionData["amount"],
      "success": true,
    },
  }

  bridge.Mu.RLock()
  if client, exists := bridge.Clients[userID]; exists {
    // Send to specific player
    sendToClient(client, confirmMsg)
  }
  bridge.Mu.RUnlock()

  // Also broadcast action to ALL players for history
  actionBroadcast := map[string]interface{}{
    "type": "player_action_broadcast",
    "payload": actionData,
  }

  broadcastToTable(bridge, tableID, actionBroadcast)
}
```

#### 4. New API Endpoints
**Add to**: `/platform/backend/internal/server/handlers/tables.go`

```go
// GET /api/hands/:handId/history
// Returns complete event history for a hand
func GetHandHistory(c *gin.Context) {
  handID := c.Param("handId")

  var events []models.GameEvent
  err := database.Where("hand_id = ?", handID).
    Order("sequence_number ASC").
    Find(&events).Error

  if err != nil {
    c.JSON(500, gin.H{"error": "Failed to fetch hand history"})
    return
  }

  // Enrich with player names, card descriptions, etc.
  enrichedHistory := enrichHandHistory(events)

  c.JSON(200, gin.H{
    "hand_id": handID,
    "events": enrichedHistory,
  })
}

// GET /api/tables/:tableId/hands
// Returns all hands for a table with summaries
func GetTableHands(c *gin.Context) {
  tableID := c.Param("tableId")

  var hands []models.Hand
  err := database.Where("table_id = ?", tableID).
    Order("started_at DESC").
    Limit(50).
    Find(&hands).Error

  if err != nil {
    c.JSON(500, gin.H{"error": "Failed to fetch table hands"})
    return
  }

  c.JSON(200, gin.H{
    "table_id": tableID,
    "hands": hands,
  })
}

// GET /api/tables/:tableId/current-hand/history
// Returns real-time history for current active hand
func GetCurrentHandHistory(c *gin.Context) {
  tableID := c.Param("tableId")

  // Get current hand ID from bridge
  handID, exists := bridge.GetCurrentHandID(tableID)
  if !exists {
    c.JSON(404, gin.H{"error": "No active hand"})
    return
  }

  var events []models.GameEvent
  err := database.Where("hand_id = ?", handID).
    Order("sequence_number ASC").
    Find(&events).Error

  if err != nil {
    c.JSON(500, gin.H{"error": "Failed to fetch history"})
    return
  }

  c.JSON(200, gin.H{
    "hand_id": handID,
    "events": enrichHandHistory(events),
  })
}
```

---

### Frontend Changes

#### 1. Fix "Processing..." Synchronization
**Modify**: `/platform/frontend/src/pages/GameView.tsx`

**Change 1**: Add action confirmation listener

```typescript
// Add new state for confirmed actions
const [confirmedActions, setConfirmedActions] = useState<Set<string>>(new Set());

// In WebSocket message handler, add new case
useEffect(() => {
  if (!ws) return;

  ws.onmessage = (event) => {
    const message = JSON.parse(event.data);

    switch (message.type) {
      case 'action_confirmed':
        // IMMEDIATE confirmation from server
        const { user_id, action } = message.payload;
        if (user_id === currentUserId && pendingAction?.type === action) {
          addConsoleLog('ACTION', `Action ${action} confirmed by server`, 'success');
          setPendingAction(null);  // Clear immediately
        }
        break;

      case 'player_action_broadcast':
        // Add to history for all players
        addActionToHistory(message.payload);
        break;

      case 'game_update':
        // ... existing code
        break;
    }
  };
}, [ws, pendingAction, currentUserId]);
```

**Change 2**: Add timeout fallback (safety net)

```typescript
// When setting pendingAction, also set a timeout
const handleAction = (action: string, amount?: number) => {
  const actionId = `${action}_${Date.now()}`;
  setPendingAction({ type: action, timestamp: Date.now(), id: actionId });

  // Safety: Clear after 5 seconds even if no confirmation
  setTimeout(() => {
    setPendingAction(prev => {
      if (prev?.id === actionId) {
        addConsoleLog('ACTION', `Action ${action} auto-cleared after timeout`, 'warning');
        return null;
      }
      return prev;
    });
  }, 5000);

  // Send action to server
  ws.send(JSON.stringify({
    type: 'game_action',
    payload: { action, amount, request_id: actionId }
  }));
};
```

#### 2. Enhanced History Panel
**Modify**: `/platform/frontend/src/components/game/HistoryPanel.tsx`

```typescript
interface GameEvent {
  id: string;
  event_type: 'hand_started' | 'player_action' | 'round_advanced' | 'showdown' | 'hand_complete';
  user_id?: string;
  player_name?: string;
  action?: string;
  amount?: number;
  betting_round?: string;
  metadata?: any;
  timestamp: Date;
  sequence_number: number;
}

interface HistoryPanelProps {
  tableId: string;
  events: GameEvent[];  // Changed from simple history entries
  isLive?: boolean;  // Whether showing current hand or past hand
}

export const HistoryPanel: React.FC<HistoryPanelProps> = ({ tableId, events, isLive = true }) => {
  return (
    <Box>
      <Stack spacing={0.5}>
        {events.map((event) => (
          <HistoryEventItem key={event.id} event={event} />
        ))}
      </Stack>
    </Box>
  );
};

const HistoryEventItem: React.FC<{ event: GameEvent }> = ({ event }) => {
  switch (event.event_type) {
    case 'hand_started':
      return (
        <Box sx={{ px: 1.5, py: 1, borderRadius: RADIUS.sm, background: `${COLORS.info.main}20` }}>
          <Typography variant="caption" sx={{ color: COLORS.info.main, fontSize: 10, fontWeight: 700 }}>
            🃏 NEW HAND #{event.metadata.hand_number}
          </Typography>
        </Box>
      );

    case 'player_action':
      return (
        <Box sx={{ px: 1.5, py: 1, borderRadius: RADIUS.sm, background: `${COLORS.background.secondary}80` }}>
          <Stack direction="row" justifyContent="space-between">
            <Typography variant="caption" sx={{ color: COLORS.text.primary, fontSize: 11, fontWeight: 600 }}>
              {event.player_name}
            </Typography>
            <Typography variant="caption" sx={{ color: COLORS.text.secondary, fontSize: 9 }}>
              {event.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            </Typography>
          </Stack>
          <Typography variant="caption" sx={{ color: getActionColor(event.action), fontSize: 10, fontWeight: 700 }}>
            {event.action.toUpperCase()}
            {event.amount > 0 && ` $${event.amount}`}
          </Typography>
        </Box>
      );

    case 'round_advanced':
      const { new_round, community_cards } = event.metadata;
      return (
        <Box sx={{ px: 1.5, py: 1, borderRadius: RADIUS.sm, background: `${COLORS.warning.main}20` }}>
          <Typography variant="caption" sx={{ color: COLORS.warning.main, fontSize: 10, fontWeight: 700 }}>
            📊 {new_round.toUpperCase()}: {community_cards.join(' ')}
          </Typography>
        </Box>
      );

    case 'showdown':
      return (
        <Box sx={{ px: 1.5, py: 1, borderRadius: RADIUS.sm, background: `${COLORS.accent.main}20` }}>
          <Typography variant="caption" sx={{ color: COLORS.accent.main, fontSize: 10, fontWeight: 700 }}>
            🎯 SHOWDOWN
          </Typography>
          {event.metadata.players_showing.map((p: any) => (
            <Typography key={p.user_id} variant="caption" sx={{ color: COLORS.text.secondary, fontSize: 9 }}>
              {p.player_name}: {p.hand_rank}
            </Typography>
          ))}
        </Box>
      );

    case 'hand_complete':
      const winners = event.metadata.winners;
      return (
        <Box sx={{ px: 1.5, py: 1, borderRadius: RADIUS.sm, background: `${COLORS.success.main}20` }}>
          <Typography variant="caption" sx={{ color: COLORS.success.main, fontSize: 10, fontWeight: 700 }}>
            🏆 {winners.map((w: any) => w.player_name).join(', ')} won ${event.metadata.final_pot}
          </Typography>
        </Box>
      );

    default:
      return null;
  }
};
```

#### 3. Real-Time History Updates
**Add to GameView.tsx**:

```typescript
const [currentHandEvents, setCurrentHandEvents] = useState<GameEvent[]>([]);

// Fetch current hand history on mount and when hand changes
useEffect(() => {
  if (!tableId) return;

  fetch(`/api/tables/${tableId}/current-hand/history`)
    .then(res => res.json())
    .then(data => {
      if (data.events) {
        setCurrentHandEvents(data.events);
      }
    })
    .catch(err => console.error('Failed to fetch hand history:', err));
}, [tableId, tableState?.current_hand?.hand_number]);

// Add events from WebSocket broadcasts
useEffect(() => {
  if (!ws) return;

  ws.onmessage = (event) => {
    const message = JSON.parse(event.data);

    if (message.type === 'player_action_broadcast') {
      // Add to current hand events
      const newEvent: GameEvent = {
        id: `${Date.now()}_${message.payload.playerId}`,
        event_type: 'player_action',
        user_id: message.payload.playerId,
        player_name: message.payload.playerName,
        action: message.payload.action,
        amount: message.payload.amount,
        betting_round: message.payload.bettingRound,
        metadata: message.payload,
        timestamp: new Date(),
        sequence_number: currentHandEvents.length,
      };

      setCurrentHandEvents(prev => [...prev, newEvent]);
    }

    // Similar for other event types (round_advanced, hand_complete, etc.)
  };
}, [ws, currentHandEvents]);
```

#### 4. Hand History Viewer (New Component)
**New File**: `/platform/frontend/src/components/game/HandHistoryViewer.tsx`

```typescript
// Modal/drawer component to view past hands
export const HandHistoryViewer: React.FC<{ tableId: string }> = ({ tableId }) => {
  const [hands, setHands] = useState<Hand[]>([]);
  const [selectedHand, setSelectedHand] = useState<number | null>(null);
  const [handEvents, setHandEvents] = useState<GameEvent[]>([]);

  // Fetch all hands for this table
  useEffect(() => {
    fetch(`/api/tables/${tableId}/hands`)
      .then(res => res.json())
      .then(data => setHands(data.hands));
  }, [tableId]);

  // Fetch events for selected hand
  const loadHandDetails = (handId: number) => {
    fetch(`/api/hands/${handId}/history`)
      .then(res => res.json())
      .then(data => {
        setSelectedHand(handId);
        setHandEvents(data.events);
      });
  };

  return (
    <Drawer>
      <Box>
        <Typography variant="h6">Hand History</Typography>

        {/* Hand list */}
        <List>
          {hands.map(hand => (
            <ListItem key={hand.id} onClick={() => loadHandDetails(hand.id)}>
              <ListItemText
                primary={`Hand #${hand.hand_number}`}
                secondary={`Pot: $${hand.pot_amount} | ${hand.num_players} players`}
              />
            </ListItem>
          ))}
        </List>

        {/* Selected hand details */}
        {selectedHand && (
          <Box>
            <Typography variant="subtitle1">Hand #{selectedHand} Details</Typography>
            <HistoryPanel tableId={tableId} events={handEvents} isLive={false} />
          </Box>
        )}
      </Box>
    </Drawer>
  );
};
```

---

## Part 3: Implementation Checklist

### Phase 1: Database & Backend Foundation (Week 1)
- [ ] 1.1 Create `game_events` table migration
- [ ] 1.2 Add `GameEvent` model to `models.go`
- [ ] 1.3 Implement `HistoryTracker` service
- [ ] 1.4 Add history tracker initialization in `main.go`
- [ ] 1.5 Write unit tests for HistoryTracker

### Phase 2: Event Recording (Week 1-2)
- [ ] 2.1 Integrate HistoryTracker into `HandleGameEvent()`
- [ ] 2.2 Record `hand_started` events
- [ ] 2.3 Record `player_action` events (dual-write to game_events + hand_actions)
- [ ] 2.4 Record `round_advanced` events (flop, turn, river)
- [ ] 2.5 Record `showdown` events
- [ ] 2.6 Record `hand_complete` events
- [ ] 2.7 Record `player_timeout` events
- [ ] 2.8 Test event recording with integration tests

### Phase 3: Action Confirmation (Week 2)
- [ ] 3.1 Implement `broadcastActionConfirmation()` function
- [ ] 3.2 Send `action_confirmed` message to action initiator
- [ ] 3.3 Send `player_action_broadcast` to all table players
- [ ] 3.4 Test with multiple concurrent players

### Phase 4: History API (Week 2)
- [ ] 4.1 Implement `GET /api/hands/:handId/history`
- [ ] 4.2 Implement `GET /api/tables/:tableId/hands`
- [ ] 4.3 Implement `GET /api/tables/:tableId/current-hand/history`
- [ ] 4.4 Add authentication/authorization checks
- [ ] 4.5 Test API endpoints with Postman/curl

### Phase 5: Frontend Synchronization Fix (Week 3)
- [ ] 5.1 Add `action_confirmed` WebSocket listener
- [ ] 5.2 Clear `pendingAction` immediately on confirmation
- [ ] 5.3 Add timeout fallback (5-second auto-clear)
- [ ] 5.4 Add request_id to action payloads
- [ ] 5.5 Test with slow network conditions
- [ ] 5.6 Test with rapid consecutive actions

### Phase 6: Frontend History Display (Week 3)
- [ ] 6.1 Update HistoryPanel to accept GameEvent[]
- [ ] 6.2 Implement HistoryEventItem component
- [ ] 6.3 Add rendering for all event types
- [ ] 6.4 Fetch current hand history on load
- [ ] 6.5 Update history from WebSocket broadcasts
- [ ] 6.6 Test history display with all event types

### Phase 7: Hand History Viewer (Week 4)
- [ ] 7.1 Create HandHistoryViewer component
- [ ] 7.2 Add hand list with summaries
- [ ] 7.3 Add hand detail view
- [ ] 7.4 Add "View History" button to GameView
- [ ] 7.5 Style and polish UI
- [ ] 7.6 Test with multiple completed hands

### Phase 8: Tournament Support (Week 4)
- [ ] 8.1 Verify event recording works for tournament tables
- [ ] 8.2 Add tournament-specific events (blinds_increased, player_eliminated)
- [ ] 8.3 Test history tracking in tournament context
- [ ] 8.4 Add tournament-specific history display elements

### Phase 9: Testing & QA (Week 5)
- [ ] 9.1 End-to-end testing: cash games
- [ ] 9.2 End-to-end testing: tournament games
- [ ] 9.3 Test with 2, 3, 6, 9 player tables
- [ ] 9.4 Test network failure scenarios
- [ ] 9.5 Load testing: 100+ concurrent hands
- [ ] 9.6 Test database performance with 10k+ events
- [ ] 9.7 Fix any bugs found during testing

### Phase 10: Optimization & Polish (Week 5-6)
- [ ] 10.1 Add database indexes for common queries
- [ ] 10.2 Implement event pagination for large hand histories
- [ ] 10.3 Add caching for frequently accessed hands
- [ ] 10.4 Optimize WebSocket broadcast performance
- [ ] 10.5 Add loading states and error handling
- [ ] 10.6 Write documentation for history system

---

## Part 4: Success Criteria

### Critical Requirements
1. ✅ "Processing..." clears within 500ms of action confirmation
2. ✅ All player actions visible to all players in real-time
3. ✅ Complete hand history (all events) saved to database
4. ✅ History panel shows all events for current hand
5. ✅ Can view past hand details via UI
6. ✅ Works for both cash and tournament games

### Performance Requirements
- Action confirmation latency: < 200ms (p95)
- History API response time: < 500ms for 50-hand list
- History detail load time: < 300ms for typical hand (30-50 events)
- Database writes: No blocking of game logic
- WebSocket broadcast: < 100ms to all table clients

### Data Integrity Requirements
- No missed events (100% capture rate)
- Events ordered correctly (sequence numbers)
- Hand history reproducible from events alone
- No orphaned events (all reference valid hand_id)

---

## Part 5: Rollout Plan

### Development Environment Testing (1 week)
- Deploy to dev server
- Manual testing by dev team
- Automated test suite execution

### Staging Environment Testing (1 week)
- Deploy to staging with production data volume
- Load testing with simulated players
- Beta user testing (invite 10-20 users)
- Monitor error rates and performance

### Production Rollout (Phased)
1. **Phase 1**: Enable event recording only (no UI changes)
   - Monitor database load
   - Verify event capture completeness

2. **Phase 2**: Enable action confirmation messages
   - Monitor "Processing..." fix effectiveness
   - Check for any synchronization regressions

3. **Phase 3**: Enable enhanced history display
   - Roll out to 10% of users
   - Gather feedback
   - Monitor frontend performance

4. **Phase 4**: Full rollout
   - Enable for all users
   - Monitor for 1 week
   - Address any issues

### Rollback Plan
- Feature flags for each component
- Database rollback scripts prepared
- Can disable new features without code deploy
- Monitoring alerts for error rate spikes

---

## Part 6: Monitoring & Observability

### Metrics to Track
- Action confirmation latency (p50, p95, p99)
- Event recording success rate
- History API response times
- Database query performance
- WebSocket message delivery rate

### Logging
- All event recordings (DEBUG level)
- Action confirmations (INFO level)
- History API requests (INFO level)
- Errors and failures (ERROR level)

### Alerts
- Event recording failure rate > 1%
- Action confirmation latency > 1s (p95)
- History API error rate > 5%
- Database connection pool exhaustion

---

## Part 7: Future Enhancements

### Post-Launch Improvements
1. **Hand Replay**: Animated replay of past hands
2. **Statistics Dashboard**: Win rate, action frequencies, profit/loss
3. **Hand Sharing**: Share interesting hands via URL
4. **Export**: Download hand history as text/JSON
5. **Search/Filter**: Find hands by criteria (won, lost, specific cards)
6. **Advanced Analytics**:
   - Showdown percentage
   - Aggression factor
   - VPIP (Voluntarily Put In Pot)
   - PFR (Pre-Flop Raise)

### Technical Debt to Address
1. **Migrate localStorage history**: Move to server-side storage
2. **Archive old events**: Move events older than 30 days to cold storage
3. **Optimize JSON queries**: Use PostgreSQL JSONB if migrating from MySQL
4. **Add event compression**: Reduce storage for large tournaments

---

## Part 8: Questions & Decisions Needed

### Open Questions
1. **Data retention**: How long to keep hand history? (Suggest: 1 year for cash, permanent for tournaments)
2. **Privacy**: Should players see hole cards of folded opponents? (Suggest: No, unless showdown)
3. **UI placement**: Where to put "Hand History" button? (Suggest: GameSidebar tab)
4. **Event granularity**: Track individual card deals or just final hands? (Suggest: Final hands only for privacy)
5. **Real-time vs polling**: Use WebSocket for all updates or allow polling fallback? (Suggest: WebSocket primary, polling for reconnect)

### Decisions Made
- ✅ Use MySQL JSON column for flexible metadata
- ✅ Dual-write to `game_events` and `hand_actions` during transition
- ✅ Send separate confirmation message instead of relying on state changes
- ✅ Store events at table level (all players see same events)
- ✅ Sequence numbers for correct ordering

---

## Part 9: Estimated Effort

| Phase | Tasks | Estimated Hours | Dependencies |
|-------|-------|-----------------|--------------|
| 1. Database & Backend Foundation | 5 | 16h | None |
| 2. Event Recording | 8 | 24h | Phase 1 |
| 3. Action Confirmation | 4 | 12h | Phase 2 |
| 4. History API | 5 | 16h | Phase 2 |
| 5. Frontend Sync Fix | 6 | 16h | Phase 3 |
| 6. Frontend History Display | 6 | 20h | Phase 4, 5 |
| 7. Hand History Viewer | 5 | 20h | Phase 6 |
| 8. Tournament Support | 4 | 12h | Phase 2-7 |
| 9. Testing & QA | 7 | 24h | Phase 8 |
| 10. Optimization & Polish | 6 | 16h | Phase 9 |

**Total Estimated Effort**: 176 hours (~4-5 weeks for 1 developer, ~2-3 weeks for 2 developers)

---

## Part 10: Risk Assessment

### High Risk Items
1. **Database performance**: Large volume of events could slow down queries
   - **Mitigation**: Proper indexing, event archival strategy, load testing

2. **WebSocket message flood**: Too many broadcasts could overwhelm clients
   - **Mitigation**: Batch events, rate limiting, message compression

3. **Race conditions**: Event ordering could get mixed up under high concurrency
   - **Mitigation**: Sequence numbers, database constraints, comprehensive testing

### Medium Risk Items
1. **Migration complexity**: Existing tables in production
   - **Mitigation**: Zero-downtime migration, backward compatibility

2. **Frontend state management**: Complex synchronization logic
   - **Mitigation**: Clear state machine, comprehensive unit tests

### Low Risk Items
1. **API design changes**: New endpoints shouldn't affect existing functionality
2. **UI changes**: Additive only, existing UI remains functional

---

## Summary

This comprehensive plan addresses all identified issues:

1. ✅ **"Processing..." bug**: Fixed via explicit action confirmation messages
2. ✅ **Actions not visible**: Fixed via `player_action_broadcast` to all players
3. ✅ **Incomplete history**: Fixed via `game_events` table tracking all event types
4. ✅ **No persistent history**: Fixed via history API and server-side storage
5. ✅ **Cash & tournament support**: Designed to work for both game types

The implementation is phased to minimize risk and allow for iterative testing and feedback.
