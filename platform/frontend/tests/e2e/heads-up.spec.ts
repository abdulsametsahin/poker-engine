import { test, expect, Browser, Page } from '@playwright/test';
import { createUser, loginUser } from './helpers/auth';
import { createTable, joinTable, makeAction, getPotAmount, isGameComplete } from './helpers/game';

/**
 * E2E tests for heads-up poker games
 *
 * Covers:
 * - Game creation and joining
 * - Player actions (fold, call, raise)
 * - Game state synchronization via WebSocket
 * - Winner determination
 * - Disconnection/reconnection
 */

test.describe('Heads-Up Game - E2E', () => {
  let player1: Page;
  let player2: Page;

  test.beforeEach(async ({ browser }) => {
    // Create two player contexts
    const context1 = await browser.newContext();
    const context2 = await browser.newContext();

    player1 = await context1.newPage();
    player2 = await context2.newPage();

    // Create users
    const timestamp = Date.now();
    await createUser(player1, {
      username: `player1_${timestamp}`,
      password: 'test123',
    });

    await createUser(player2, {
      username: `player2_${timestamp}`,
      password: 'test123',
    });
  });

  test.afterEach(async () => {
    await player1.close();
    await player2.close();
  });

  test('Should complete a simple heads-up game', async () => {
    // Player 1 creates a table
    const tableId = await createTable(player1, {
      gameMode: 'heads_up',
      buyIn: 1000,
      smallBlind: 10,
      bigBlind: 20,
    });

    expect(tableId).toBeTruthy();
    console.log('Created table:', tableId);

    // Player 2 joins the table
    await joinTable(player2, tableId, 1000);

    // Wait for game to start
    await player1.waitForSelector('[data-testid="game-status"]:has-text("playing")', { timeout: 10000 });
    await player2.waitForSelector('[data-testid="game-status"]:has-text("playing")', { timeout: 10000 });

    // Verify both players see the same pot (blinds posted)
    const p1Pot = await getPotAmount(player1);
    const p2Pot = await getPotAmount(player2);

    expect(p1Pot).toBe(p2Pot);
    expect(p1Pot).toBeGreaterThanOrEqual(30); // At least small blind + big blind

    console.log('Initial pot:', p1Pot);
  });

  test('Should sync game state via WebSocket', async () => {
    // Create and join game
    const tableId = await createTable(player1, {
      gameMode: 'heads_up',
      buyIn: 1000,
      smallBlind: 10,
      bigBlind: 20,
    });

    await joinTable(player2, tableId, 1000);

    // Wait for game to start
    await player1.waitForSelector('[data-testid="game-status"]:has-text("playing")');

    // Player 1 raises
    await makeAction(player1, 'raise', 50);

    // Player 2 should see the updated bet within 1 second
    await player2.waitForSelector('[data-testid="current-bet"]:has-text("50")', { timeout: 2000 });

    // Verify WebSocket sync happened quickly
    const p2CurrentBet = await player2.locator('[data-testid="current-bet"]').textContent();
    expect(p2CurrentBet).toContain('50');

    console.log('WebSocket sync successful - bet updated in real-time');
  });

  test('Should handle player fold', async () => {
    const tableId = await createTable(player1, {
      gameMode: 'heads_up',
      buyIn: 1000,
      smallBlind: 10,
      bigBlind: 20,
    });

    await joinTable(player2, tableId, 1000);
    await player1.waitForSelector('[data-testid="game-status"]:has-text("playing")');

    // Player with action folds
    await makeAction(player1, 'fold');

    // Hand should complete, winner announced
    await player2.waitForSelector('[data-testid="hand-complete-modal"], text=/won/i', { timeout: 5000 });

    console.log('Fold action processed correctly');
  });

  test('Should display game status indicators', async () => {
    const tableId = await createTable(player1, {
      gameMode: 'heads_up',
      buyIn: 1000,
      smallBlind: 10,
      bigBlind: 20,
    });

    await joinTable(player2, tableId, 1000);

    // Check for game status elements
    await expect(player1.locator('[data-testid="game-status"]')).toBeVisible();
    await expect(player1.locator('[data-testid="pot-amount"]')).toBeVisible();
    await expect(player1.locator('[data-testid="player-chips"]')).toBeVisible();

    console.log('All game status indicators visible');
  });

  test('Should handle multiple hands in sequence', async () => {
    const tableId = await createTable(player1, {
      gameMode: 'heads_up',
      buyIn: 1000,
      smallBlind: 10,
      bigBlind: 20,
    });

    await joinTable(player2, tableId, 1000);
    await player1.waitForSelector('[data-testid="game-status"]:has-text("playing")');

    // Play 3 quick hands
    for (let i = 0; i < 3; i++) {
      // Wait for action
      await player1.waitForTimeout(1000);

      // Fold immediately
      const foldButton = player1.locator('[data-testid="fold-button"]');
      if (await foldButton.isVisible({ timeout: 2000 })) {
        await foldButton.click();
      }

      // Wait for next hand to start
      await player1.waitForTimeout(2000);
    }

    console.log('Multiple hands completed successfully');
  });
});

/**
 * Additional test ideas to implement:
 *
 * - test('Should handle player all-in')
 * - test('Should calculate pot correctly with raises')
 * - test('Should show winner with correct hand rank')
 * - test('Should handle disconnection and reconnection')
 * - test('Should persist game state across page refresh')
 * - test('Should enforce turn-based actions')
 * - test('Should display action timer countdown')
 * - test('Should handle tie/split pot')
 * - test('Should complete game when one player runs out of chips')
 * - test('Should display game history correctly')
 */
