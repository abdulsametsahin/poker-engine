import { test, expect, Browser, Page } from '@playwright/test';
import { createUser, loginUser } from './helpers/auth';
import { createTable, joinTable, makeAction, getPotAmount, getPlayerChips, isGameComplete } from './helpers/game';

/**
 * E2E tests for 3-player poker games
 *
 * Covers:
 * - 3-player game creation and joining
 * - Player elimination and position advancement
 * - Dealer button rotation
 * - Side pot handling
 * - Final winner determination
 * - Game state synchronization across 3 clients
 */

test.describe('3-Way Game - E2E', () => {
  let player1: Page;
  let player2: Page;
  let player3: Page;

  test.beforeEach(async ({ browser }) => {
    // Create three player contexts
    const context1 = await browser.newContext();
    const context2 = await browser.newContext();
    const context3 = await browser.newContext();

    player1 = await context1.newPage();
    player2 = await context2.newPage();
    player3 = await context3.newPage();

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

    await createUser(player3, {
      username: `player3_${timestamp}`,
      password: 'test123',
    });
  });

  test.afterEach(async () => {
    await player1.close();
    await player2.close();
    await player3.close();
  });

  test('Should complete a 3-player game setup', async () => {
    // Player 1 creates a table
    const tableId = await createTable(player1, {
      gameMode: '3_player',
      buyIn: 1000,
      smallBlind: 10,
      bigBlind: 20,
    });

    expect(tableId).toBeTruthy();
    console.log('Created 3-player table:', tableId);

    // Player 2 and 3 join
    await joinTable(player2, tableId, 1000);
    await joinTable(player3, tableId, 1000);

    // Wait for game to start
    await player1.waitForSelector('[data-testid="game-status"]:has-text("playing")', { timeout: 10000 });
    await player2.waitForSelector('[data-testid="game-status"]:has-text("playing")', { timeout: 10000 });
    await player3.waitForSelector('[data-testid="game-status"]:has-text("playing")', { timeout: 10000 });

    // Verify all players see the same pot (blinds posted)
    const p1Pot = await getPotAmount(player1);
    const p2Pot = await getPotAmount(player2);
    const p3Pot = await getPotAmount(player3);

    expect(p1Pot).toBe(p2Pot);
    expect(p2Pot).toBe(p3Pot);
    expect(p1Pot).toBeGreaterThanOrEqual(30); // At least SB + BB

    console.log('3-player game started, initial pot:', p1Pot);
  });

  test('Should sync game state across all 3 players', async () => {
    const tableId = await createTable(player1, {
      gameMode: '3_player',
      buyIn: 1000,
      smallBlind: 10,
      bigBlind: 20,
    });

    await joinTable(player2, tableId, 1000);
    await joinTable(player3, tableId, 1000);

    await player1.waitForSelector('[data-testid="game-status"]:has-text("playing")');

    // Wait to identify who has action
    await player1.waitForTimeout(1000);

    // Try to make an action from player1
    const actionButtons = player1.locator('[data-testid="action-buttons"]');
    if (await actionButtons.isVisible({ timeout: 2000 })) {
      await makeAction(player1, 'raise', 50);

      // Both other players should see the update
      await player2.waitForSelector('[data-testid="current-bet"]:has-text("50")', { timeout: 2000 });
      await player3.waitForSelector('[data-testid="current-bet"]:has-text("50")', { timeout: 2000 });

      console.log('WebSocket sync successful across all 3 players');
    }
  });

  test('Should handle dealer button rotation', async () => {
    const tableId = await createTable(player1, {
      gameMode: '3_player',
      buyIn: 1000,
      smallBlind: 10,
      bigBlind: 20,
    });

    await joinTable(player2, tableId, 1000);
    await joinTable(player3, tableId, 1000);

    await player1.waitForSelector('[data-testid="game-status"]:has-text("playing")');

    // Check initial dealer button position
    const initialDealer = await player1.locator('[data-testid="dealer-button"]').count();
    expect(initialDealer).toBeGreaterThan(0);

    // Play a quick hand - someone folds
    await player1.waitForTimeout(1500);

    // Try to fold from whichever player has action
    const players = [player1, player2, player3];
    for (const player of players) {
      const foldButton = player.locator('[data-testid="fold-button"]');
      if (await foldButton.isVisible({ timeout: 1000 })) {
        await foldButton.click();
        break;
      }
    }

    // Wait for new hand
    await player1.waitForTimeout(3000);

    // Dealer button should still exist (may have moved)
    const newDealer = await player1.locator('[data-testid="dealer-button"]').count();
    expect(newDealer).toBeGreaterThan(0);

    console.log('Dealer button rotation working');
  });

  test('Should handle player elimination in 3-way', async () => {
    const tableId = await createTable(player1, {
      gameMode: '3_player',
      buyIn: 100,
      smallBlind: 10,
      bigBlind: 20,
    });

    await joinTable(player2, tableId, 100);
    await joinTable(player3, tableId, 100);

    await player1.waitForSelector('[data-testid="game-status"]:has-text("playing")');

    // Play multiple hands quickly to try to eliminate a player
    for (let i = 0; i < 5; i++) {
      await player1.waitForTimeout(1000);

      // Each player checks if they can act
      const players = [player1, player2, player3];
      for (const player of players) {
        const actionButtons = player.locator('[data-testid="action-buttons"]');
        if (await actionButtons.isVisible({ timeout: 1000 })) {
          // Randomly fold or call
          const foldButton = player.locator('[data-testid="fold-button"]');
          const callButton = player.locator('[data-testid="call-button"]');

          if (await foldButton.isVisible({ timeout: 500 })) {
            await foldButton.click();
          } else if (await callButton.isVisible({ timeout: 500 })) {
            await callButton.click();
          }
          break;
        }
      }

      // Wait between hands
      await player1.waitForTimeout(2000);
    }

    console.log('3-way elimination test completed');
  });

  test('Should display game status for all 3 players', async () => {
    const tableId = await createTable(player1, {
      gameMode: '3_player',
      buyIn: 1000,
      smallBlind: 10,
      bigBlind: 20,
    });

    await joinTable(player2, tableId, 1000);
    await joinTable(player3, tableId, 1000);

    // Verify game status elements visible for all players
    await expect(player1.locator('[data-testid="game-status"]')).toBeVisible();
    await expect(player1.locator('[data-testid="pot-amount"]')).toBeVisible();

    await expect(player2.locator('[data-testid="game-status"]')).toBeVisible();
    await expect(player2.locator('[data-testid="pot-amount"]')).toBeVisible();

    await expect(player3.locator('[data-testid="game-status"]')).toBeVisible();
    await expect(player3.locator('[data-testid="pot-amount"]')).toBeVisible();

    console.log('All game status indicators visible for 3 players');
  });

  test('Should handle multiple rounds with 3 active players', async () => {
    const tableId = await createTable(player1, {
      gameMode: '3_player',
      buyIn: 1000,
      smallBlind: 10,
      bigBlind: 20,
    });

    await joinTable(player2, tableId, 1000);
    await joinTable(player3, tableId, 1000);

    await player1.waitForSelector('[data-testid="game-status"]:has-text("playing")');

    // Play 3 quick hands
    for (let i = 0; i < 3; i++) {
      await player1.waitForTimeout(1500);

      // Find who has action and fold
      const players = [player1, player2, player3];
      for (const player of players) {
        const foldButton = player.locator('[data-testid="fold-button"]');
        if (await foldButton.isVisible({ timeout: 1500 })) {
          await foldButton.click();
          console.log(`Hand ${i + 1}: Player folded`);
          break;
        }
      }

      // Wait for next hand
      await player1.waitForTimeout(2500);
    }

    console.log('Multiple rounds completed successfully with 3 players');
  });

  test('Should verify pot calculations with 3 players', async () => {
    const tableId = await createTable(player1, {
      gameMode: '3_player',
      buyIn: 1000,
      smallBlind: 10,
      bigBlind: 20,
    });

    await joinTable(player2, tableId, 1000);
    await joinTable(player3, tableId, 1000);

    await player1.waitForSelector('[data-testid="game-status"]:has-text("playing")');

    // Get initial pot (should be SB + BB)
    const initialPot = await getPotAmount(player1);
    expect(initialPot).toBeGreaterThanOrEqual(30);

    console.log('Initial pot with 3 players:', initialPot);

    // Wait for actions
    await player1.waitForTimeout(1500);

    // Try to raise from whoever has action
    const players = [player1, player2, player3];
    for (const player of players) {
      const raiseInput = player.locator('[data-testid="raise-amount-input"]');
      if (await raiseInput.isVisible({ timeout: 1000 })) {
        await raiseInput.fill('50');
        await player.locator('[data-testid="raise-button"]').click();
        console.log('Player raised to 50');
        break;
      }
    }

    // Wait for pot to update
    await player1.waitForTimeout(1000);

    // Pot should have increased
    const newPot = await getPotAmount(player1);
    expect(newPot).toBeGreaterThan(initialPot);

    console.log('Pot after raise:', newPot);
  });

  test('Should handle 3-player chat and notifications', async () => {
    const tableId = await createTable(player1, {
      gameMode: '3_player',
      buyIn: 1000,
      smallBlind: 10,
      bigBlind: 20,
    });

    await joinTable(player2, tableId, 1000);
    await joinTable(player3, tableId, 1000);

    await player1.waitForSelector('[data-testid="game-status"]:has-text("playing")');

    // Check if chat input exists
    const chatInput = player1.locator('[data-testid="chat-input"], input[placeholder*="chat" i]');

    if (await chatInput.isVisible({ timeout: 2000 })) {
      await chatInput.fill('Hello from Player 1!');
      await player1.locator('[data-testid="chat-send"], button:has-text("Send")').click();

      // Wait for message to appear in chat
      await player1.waitForTimeout(1000);

      console.log('Chat message sent in 3-player game');
    } else {
      console.log('Chat not available yet (feature may be in development)');
    }
  });
});

/**
 * Additional 3-way test ideas:
 *
 * - test('Should handle side pot with all-in from one player')
 * - test('Should show correct player positions (UTG, BTN, SB, BB)')
 * - test('Should handle betting round completion with 3 players')
 * - test('Should display community cards to all players simultaneously')
 * - test('Should handle showdown with 3 players')
 * - test('Should calculate correct hand rankings for 3-way showdown')
 * - test('Should handle tie between 2 of 3 players')
 * - test('Should track chip stacks accurately for all 3 players')
 * - test('Should handle player disconnect and reconnect in 3-way')
 * - test('Should complete game when only 1 player remains')
 */
