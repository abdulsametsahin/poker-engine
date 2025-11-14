import { Page, expect } from '@playwright/test';

/**
 * Helper functions for game interactions in E2E tests
 */

export interface CreateTableOptions {
  gameMode: 'heads_up' | '3_player' | '6_player';
  buyIn: number;
  smallBlind: number;
  bigBlind: number;
}

/**
 * Create a new poker table
 */
export async function createTable(page: Page, options: CreateTableOptions): Promise<string> {
  await page.goto('/lobby');

  // Click create table button
  const createButton = page.locator('[data-testid="create-table-button"], button:has-text("Create Table")');
  await createButton.click();

  // Fill form
  await page.selectOption('[data-testid="game-mode-select"], select[name="gameMode"]', options.gameMode);
  await page.fill('[data-testid="buy-in-input"], input[name="buyIn"]', options.buyIn.toString());
  await page.fill('[data-testid="small-blind-input"], input[name="smallBlind"]', options.smallBlind.toString());
  await page.fill('[data-testid="big-blind-input"], input[name="bigBlind"]', options.bigBlind.toString());

  // Submit
  await page.click('[data-testid="create-button"], button[type="submit"]:has-text("Create")');

  // Wait for navigation to game
  await page.waitForURL('**/game/**', { timeout: 10000 });

  // Extract table ID from URL
  const url = page.url();
  const tableId = url.split('/').pop() || '';

  return tableId;
}

/**
 * Join an existing table
 */
export async function joinTable(page: Page, tableId: string, buyIn: number) {
  await page.goto(`/game/${tableId}`);

  // Fill buy-in amount
  const buyInInput = page.locator('[data-testid="buy-in-input"], input[name="buyIn"]');

  if (await buyInInput.isVisible()) {
    await buyInInput.fill(buyIn.toString());
    await page.click('[data-testid="join-button"], button:has-text("Join")');
  }

  // Wait for game to load
  await page.waitForSelector('[data-testid="poker-table"], .poker-table', { timeout: 10000 });
}

/**
 * Make a game action (fold, check, call, raise, all-in)
 */
export async function makeAction(page: Page, action: string, amount?: number) {
  // Wait for action buttons to be visible
  await page.waitForSelector('[data-testid="action-buttons"], .action-buttons', { timeout: 5000 });

  if (amount && action === 'raise') {
    // Fill raise amount
    const raiseInput = page.locator('[data-testid="raise-amount-input"], input[name="raiseAmount"]');
    await raiseInput.fill(amount.toString());
  }

  // Click action button
  const actionButton = page.locator(`[data-testid="${action}-button"], button:has-text("${action}", "i")`).first();
  await actionButton.click();

  // Wait a bit for action to process
  await page.waitForTimeout(500);
}

/**
 * Wait for your turn
 */
export async function waitForTurn(page: Page, timeout = 30000) {
  // Wait for action buttons to be enabled
  await page.waitForSelector('[data-testid="action-buttons"]:not([disabled])', { timeout });
}

/**
 * Get current pot amount
 */
export async function getPotAmount(page: Page): Promise<number> {
  const potElement = page.locator('[data-testid="pot-amount"], .pot-amount');
  const text = await potElement.textContent();
  const amount = parseInt(text?.replace(/[^0-9]/g, '') || '0');
  return amount;
}

/**
 * Get player chip count
 */
export async function getPlayerChips(page: Page, seatNumber?: number): Promise<number> {
  let selector = '[data-testid="player-chips"], .player-chips';

  if (seatNumber !== undefined) {
    selector = `[data-testid="player-${seatNumber}-chips"]`;
  }

  const chipsElement = page.locator(selector).first();
  const text = await chipsElement.textContent();
  const chips = parseInt(text?.replace(/[^0-9]/g, '') || '0');
  return chips;
}

/**
 * Check if game is complete
 */
export async function isGameComplete(page: Page): Promise<boolean> {
  const completeModal = page.locator('[data-testid="game-complete-modal"], .game-complete-modal');
  return await completeModal.isVisible({ timeout: 1000 }).catch(() => false);
}

/**
 * Get winner name
 */
export async function getWinner(page: Page): Promise<string | null> {
  const winnerElement = page.locator('[data-testid="winner-name"], .winner-name');

  if (await winnerElement.isVisible()) {
    return await winnerElement.textContent();
  }

  return null;
}

/**
 * Quick match - join matchmaking queue
 */
export async function quickMatch(page: Page, gameMode: string) {
  await page.goto('/lobby');

  // Click quick match for specific game mode
  const quickMatchButton = page.locator(`[data-testid="quick-match-${gameMode}"], button:has-text("${gameMode}")`);
  await quickMatchButton.click();

  // Wait for match found modal
  await page.waitForSelector('[data-testid="match-found-modal"]', { timeout: 30000 });

  // Wait for countdown and automatic navigation
  await page.waitForURL('**/game/**', { timeout: 10000 });

  // Extract table ID
  const url = page.url();
  return url.split('/').pop() || '';
}
