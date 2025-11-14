import { Page, expect } from '@playwright/test';

/**
 * Helper functions for tournament interactions in E2E tests
 */

export interface CreateTournamentOptions {
  name: string;
  buyIn: number;
  startingChips: number;
  maxPlayers: number;
  minPlayers: number;
  blindLevels?: Array<{
    small: number;
    big: number;
    duration: number;
  }>;
}

/**
 * Create a new tournament
 */
export async function createTournament(page: Page, options: CreateTournamentOptions): Promise<string> {
  await page.goto('/tournaments');

  // Click create tournament button
  const createButton = page.locator('[data-testid="create-tournament-button"], button:has-text("Create Tournament")');
  await createButton.click();

  // Fill form
  await page.fill('[data-testid="tournament-name-input"], input[name="name"]', options.name);
  await page.fill('[data-testid="buy-in-input"], input[name="buyIn"]', options.buyIn.toString());
  await page.fill('[data-testid="starting-chips-input"], input[name="startingChips"]', options.startingChips.toString());
  await page.fill('[data-testid="max-players-input"], input[name="maxPlayers"]', options.maxPlayers.toString());
  await page.fill('[data-testid="min-players-input"], input[name="minPlayers"]', options.minPlayers.toString());

  // Submit
  await page.click('[data-testid="create-button"], button[type="submit"]:has-text("Create")');

  // Wait for success message with tournament code
  await page.waitForSelector('text=/Tournament Code:/i', { timeout: 10000 });

  // Extract tournament ID from URL or response
  const tournamentCards = page.locator('[data-testid^="tournament-"], .tournament-card');
  const firstCard = tournamentCards.first();
  const tournamentId = await firstCard.getAttribute('data-tournament-id') || '';

  return tournamentId;
}

/**
 * Register for a tournament
 */
export async function registerForTournament(page: Page, tournamentId: string, buyIn: number) {
  await page.goto(`/tournaments/${tournamentId}`);

  // Click register button
  const registerButton = page.locator('[data-testid="register-button"], button:has-text("Register")');

  if (await registerButton.isVisible()) {
    await registerButton.click();

    // Confirm buy-in if needed
    const confirmButton = page.locator('[data-testid="confirm-button"], button:has-text("Confirm")');

    if (await confirmButton.isVisible({ timeout: 2000 })) {
      await confirmButton.click();
    }

    // Wait for registration success
    await page.waitForSelector('text=/Registered/i, [data-testid="registered-badge"]', { timeout: 5000 });
  }
}

/**
 * Join tournament by code
 */
export async function joinTournamentByCode(page: Page, code: string, buyIn: number) {
  await page.goto('/tournaments');

  // Click join by code
  const joinButton = page.locator('[data-testid="join-by-code-button"], button:has-text("Join by Code")');
  await joinButton.click();

  // Enter code
  await page.fill('[data-testid="tournament-code-input"], input[name="code"]', code);
  await page.click('[data-testid="join-button"], button:has-text("Join")');

  // Wait for navigation to tournament detail
  await page.waitForURL('**/tournaments/**', { timeout: 10000 });

  // Register
  await page.click('[data-testid="register-button"], button:has-text("Register")');

  // Wait for registration success
  await page.waitForSelector('text=/Registered/i', { timeout: 5000 });
}

/**
 * Start a tournament (creator only)
 */
export async function startTournament(page: Page) {
  const startButton = page.locator('[data-testid="start-tournament-button"], button:has-text("Start Tournament")');

  if (await startButton.isVisible()) {
    await startButton.click();

    // Confirm if needed
    const confirmButton = page.locator('[data-testid="confirm-start-button"], button:has-text("Confirm")');

    if (await confirmButton.isVisible({ timeout: 2000 })) {
      await confirmButton.click();
    }

    // Wait for tournament to start
    await page.waitForSelector('text=/In Progress/i, [data-status="in_progress"]', { timeout: 10000 });
  }
}

/**
 * Pause a tournament (creator only)
 */
export async function pauseTournament(page: Page) {
  const pauseButton = page.locator('[data-testid="pause-tournament-button"], button:has-text("Pause")');
  await pauseButton.click();

  // Wait for paused state
  await page.waitForSelector('[data-testid="tournament-paused-modal"], text=/Paused/i', { timeout: 5000 });
}

/**
 * Resume a tournament (creator only)
 */
export async function resumeTournament(page: Page) {
  const resumeButton = page.locator('[data-testid="resume-tournament-button"], button:has-text("Resume")');
  await resumeButton.click();

  // Wait for resumed state
  await page.waitForSelector('text=/In Progress/i', { timeout: 5000 });
}

/**
 * Get tournament player count
 */
export async function getTournamentPlayerCount(page: Page): Promise<number> {
  const countElement = page.locator('[data-testid="player-count"], .player-count');
  const text = await countElement.textContent();

  // Extract number from "X/Y" format
  const match = text?.match(/(\d+)/);
  return match ? parseInt(match[1]) : 0;
}

/**
 * Get tournament status
 */
export async function getTournamentStatus(page: Page): Promise<string> {
  const statusElement = page.locator('[data-testid="tournament-status"], .tournament-status');
  const status = await statusElement.textContent();
  return status?.toLowerCase().trim() || '';
}

/**
 * Check if player is eliminated
 */
export async function isPlayerEliminated(page: Page): Promise<boolean> {
  const eliminatedBadge = page.locator('[data-testid="eliminated-badge"], text=/Eliminated/i');
  return await eliminatedBadge.isVisible({ timeout: 1000 }).catch(() => false);
}

/**
 * Get tournament standings
 */
export async function getTournamentStandings(page: Page): Promise<Array<{ position: number; name: string; chips: number }>> {
  const standingsRows = page.locator('[data-testid="standings-row"], .standings-row');
  const count = await standingsRows.count();

  const standings = [];

  for (let i = 0; i < count; i++) {
    const row = standingsRows.nth(i);
    const position = parseInt(await row.locator('[data-testid="position"]').textContent() || '0');
    const name = await row.locator('[data-testid="player-name"]').textContent() || '';
    const chipsText = await row.locator('[data-testid="chips"]').textContent() || '0';
    const chips = parseInt(chipsText.replace(/[^0-9]/g, ''));

    standings.push({ position, name, chips });
  }

  return standings;
}
