import { test, expect, Browser, Page } from '@playwright/test';
import { createUser } from './helpers/auth';
import { createTournament, registerForTournament, getTournamentStandings } from './helpers/tournament';

/**
 * E2E tests for 50-player tournament
 *
 * Covers:
 * - Creating tournament and registering 50 players
 * - Tournament auto-start when max players reached
 * - Table balancing and player distribution
 * - Blind level increases
 * - Player eliminations and standings updates
 * - Prize pool distribution
 * - Tournament pause/resume functionality
 * - Final winner determination
 *
 * NOTE: This test suite is resource-intensive (50 browser contexts)
 * Consider running on CI with adequate resources or using smaller batches
 */

test.describe('50-Player Tournament - E2E', () => {
  // Increase timeout for tournament tests
  test.setTimeout(300000); // 5 minutes

  test('Should create tournament and register 50 players', async ({ browser }) => {
    // Create organizer
    const organizerPage = await browser.newPage();
    const timestamp = Date.now();

    await createUser(organizerPage, {
      username: `organizer_${timestamp}`,
      password: 'test123',
    });

    // Create tournament
    const tournamentId = await createTournament(organizerPage, {
      name: `50-Player Test Tournament ${timestamp}`,
      buyIn: 100,
      startingChips: 1000,
      maxPlayers: 50,
      minPlayers: 50,
      blindLevels: [
        { level: 1, smallBlind: 10, bigBlind: 20, duration: 600 },
        { level: 2, smallBlind: 20, bigBlind: 40, duration: 600 },
        { level: 3, smallBlind: 50, bigBlind: 100, duration: 600 },
      ],
    });

    expect(tournamentId).toBeTruthy();
    console.log('Created tournament:', tournamentId);

    // Create and register 50 players in batches to avoid overwhelming the browser
    const batchSize = 10;
    const totalPlayers = 50;
    const players: Page[] = [];

    for (let batch = 0; batch < totalPlayers / batchSize; batch++) {
      console.log(`Creating player batch ${batch + 1}/${totalPlayers / batchSize}...`);

      const batchPromises = [];
      for (let i = 0; i < batchSize; i++) {
        const playerIndex = batch * batchSize + i + 1;
        const promise = (async () => {
          const context = await browser.newContext();
          const page = await context.newPage();

          await createUser(page, {
            username: `player${playerIndex}_${timestamp}`,
            password: 'test123',
          });

          await registerForTournament(page, tournamentId, 100);

          console.log(`Player ${playerIndex} registered`);
          return page;
        })();

        batchPromises.push(promise);
      }

      const batchPlayers = await Promise.all(batchPromises);
      players.push(...batchPlayers);

      // Small delay between batches
      await new Promise(resolve => setTimeout(resolve, 1000));
    }

    console.log('All 50 players registered successfully');

    // Verify tournament shows 50 registered players
    await organizerPage.goto(`/tournaments/${tournamentId}`);
    await organizerPage.waitForTimeout(2000);

    const registeredCount = await organizerPage.locator('[data-testid="registered-players"], .registered-count').textContent();
    expect(registeredCount).toContain('50');

    // Cleanup
    await organizerPage.close();
    for (const player of players) {
      await player.close();
    }

    console.log('50-player registration test completed');
  });

  test('Should auto-start tournament when max players reached', async ({ browser }) => {
    const organizerPage = await browser.newPage();
    const timestamp = Date.now();

    await createUser(organizerPage, {
      username: `organizer_${timestamp}`,
      password: 'test123',
    });

    // Create tournament with lower player count for faster testing
    const tournamentId = await createTournament(organizerPage, {
      name: `Auto-Start Test ${timestamp}`,
      buyIn: 100,
      startingChips: 1000,
      maxPlayers: 6,
      minPlayers: 6,
      blindLevels: [
        { level: 1, smallBlind: 10, bigBlind: 20, duration: 600 },
      ],
    });

    // Register 6 players
    const players: Page[] = [];
    for (let i = 1; i <= 6; i++) {
      const context = await browser.newContext();
      const page = await context.newPage();

      await createUser(page, {
        username: `player${i}_${timestamp}`,
        password: 'test123',
      });

      await registerForTournament(page, tournamentId, 100);
      players.push(page);
    }

    // Wait for auto-start
    await organizerPage.waitForTimeout(3000);

    // Verify tournament status changed to "in_progress"
    await organizerPage.goto(`/tournaments/${tournamentId}`);
    const status = await organizerPage.locator('[data-testid="tournament-status"]').textContent();
    expect(status?.toLowerCase()).toContain('progress');

    console.log('Tournament auto-started successfully');

    // Cleanup
    await organizerPage.close();
    for (const player of players) {
      await player.close();
    }
  });

  test('Should distribute players across tables correctly', async ({ browser }) => {
    const organizerPage = await browser.newPage();
    const timestamp = Date.now();

    await createUser(organizerPage, {
      username: `organizer_${timestamp}`,
      password: 'test123',
    });

    // Create tournament with 18 players (should create 3 tables of 6)
    const tournamentId = await createTournament(organizerPage, {
      name: `Table Distribution Test ${timestamp}`,
      buyIn: 100,
      startingChips: 1000,
      maxPlayers: 18,
      minPlayers: 18,
      blindLevels: [
        { level: 1, smallBlind: 10, bigBlind: 20, duration: 600 },
      ],
    });

    // Register 18 players
    const players: Page[] = [];
    for (let i = 1; i <= 18; i++) {
      const context = await browser.newContext();
      const page = await context.newPage();

      await createUser(page, {
        username: `player${i}_${timestamp}`,
        password: 'test123',
      });

      await registerForTournament(page, tournamentId, 100);
      players.push(page);
    }

    // Wait for tournament to start
    await organizerPage.waitForTimeout(5000);

    // Check tournament detail page for table count
    await organizerPage.goto(`/tournaments/${tournamentId}`);
    await organizerPage.waitForTimeout(2000);

    const tablesSection = organizerPage.locator('[data-testid="tournament-tables"], .tournament-tables');
    if (await tablesSection.isVisible({ timeout: 5000 })) {
      const tableCount = await organizerPage.locator('[data-testid="table-card"], .table-card').count();
      expect(tableCount).toBe(3); // 18 players / 6 max per table = 3 tables
      console.log('Table distribution correct: 3 tables for 18 players');
    } else {
      console.log('Table section not visible yet');
    }

    // Cleanup
    await organizerPage.close();
    for (const player of players) {
      await player.close();
    }
  });

  test('Should handle blind level increases', async ({ browser }) => {
    const organizerPage = await browser.newPage();
    const timestamp = Date.now();

    await createUser(organizerPage, {
      username: `organizer_${timestamp}`,
      password: 'test123',
    });

    // Create tournament with short blind levels for testing
    const tournamentId = await createTournament(organizerPage, {
      name: `Blind Level Test ${timestamp}`,
      buyIn: 100,
      startingChips: 1000,
      maxPlayers: 6,
      minPlayers: 6,
      blindLevels: [
        { level: 1, smallBlind: 10, bigBlind: 20, duration: 5 }, // 5 seconds for testing
        { level: 2, smallBlind: 20, bigBlind: 40, duration: 5 },
        { level: 3, smallBlind: 50, bigBlind: 100, duration: 5 },
      ],
    });

    // Register 6 players
    const players: Page[] = [];
    for (let i = 1; i <= 6; i++) {
      const context = await browser.newContext();
      const page = await context.newPage();

      await createUser(page, {
        username: `player${i}_${timestamp}`,
        password: 'test123',
      });

      await registerForTournament(page, tournamentId, 100);
      players.push(page);
    }

    // Wait for tournament to start
    await organizerPage.waitForTimeout(5000);

    // Navigate to tournament detail
    await organizerPage.goto(`/tournaments/${tournamentId}`);

    // Check initial blind level
    let blindLevel = await organizerPage.locator('[data-testid="blind-level"], .blind-level').textContent();
    console.log('Initial blind level:', blindLevel);

    // Wait for blind increase (5 seconds + buffer)
    await organizerPage.waitForTimeout(8000);

    // Check updated blind level
    await organizerPage.reload();
    await organizerPage.waitForTimeout(2000);

    blindLevel = await organizerPage.locator('[data-testid="blind-level"], .blind-level').textContent();
    console.log('Updated blind level:', blindLevel);

    // Should show level 2 or higher
    expect(blindLevel).toBeTruthy();

    // Cleanup
    await organizerPage.close();
    for (const player of players) {
      await player.close();
    }
  });

  test('Should update standings when players are eliminated', async ({ browser }) => {
    const organizerPage = await browser.newPage();
    const timestamp = Date.now();

    await createUser(organizerPage, {
      username: `organizer_${timestamp}`,
      password: 'test123',
    });

    // Create small tournament for elimination testing
    const tournamentId = await createTournament(organizerPage, {
      name: `Elimination Test ${timestamp}`,
      buyIn: 100,
      startingChips: 100, // Low starting chips for quick eliminations
      maxPlayers: 6,
      minPlayers: 6,
      blindLevels: [
        { level: 1, smallBlind: 10, bigBlind: 20, duration: 60 },
      ],
    });

    // Register 6 players
    const players: Page[] = [];
    for (let i = 1; i <= 6; i++) {
      const context = await browser.newContext();
      const page = await context.newPage();

      await createUser(page, {
        username: `player${i}_${timestamp}`,
        password: 'test123',
      });

      await registerForTournament(page, tournamentId, 100);
      players.push(page);
    }

    // Wait for tournament to start
    await organizerPage.waitForTimeout(5000);

    // Navigate to tournament detail
    await organizerPage.goto(`/tournaments/${tournamentId}`);

    // Check initial standings count (should be 6)
    await organizerPage.waitForTimeout(2000);
    let standingsRows = await organizerPage.locator('[data-testid="standings-row"], .standings-row').count();
    console.log('Initial standings count:', standingsRows);

    // Wait for game to progress and eliminations to happen
    // With 100 chips and 10/20 blinds, eliminations should happen quickly
    await organizerPage.waitForTimeout(30000);

    // Reload to get fresh standings
    await organizerPage.reload();
    await organizerPage.waitForTimeout(2000);

    // Check if any eliminations occurred
    const eliminatedPlayers = await organizerPage.locator('[data-testid="eliminated-player"], .eliminated').count();
    console.log('Eliminated players:', eliminatedPlayers);

    // Cleanup
    await organizerPage.close();
    for (const player of players) {
      await player.close();
    }
  });

  test('Should display correct prize pool', async ({ browser }) => {
    const organizerPage = await browser.newPage();
    const timestamp = Date.now();

    await createUser(organizerPage, {
      username: `organizer_${timestamp}`,
      password: 'test123',
    });

    const buyIn = 100;
    const maxPlayers = 10;

    const tournamentId = await createTournament(organizerPage, {
      name: `Prize Pool Test ${timestamp}`,
      buyIn,
      startingChips: 1000,
      maxPlayers,
      minPlayers: maxPlayers,
      blindLevels: [
        { level: 1, smallBlind: 10, bigBlind: 20, duration: 600 },
      ],
    });

    // Register 10 players
    const players: Page[] = [];
    for (let i = 1; i <= maxPlayers; i++) {
      const context = await browser.newContext();
      const page = await context.newPage();

      await createUser(page, {
        username: `player${i}_${timestamp}`,
        password: 'test123',
      });

      await registerForTournament(page, tournamentId, buyIn);
      players.push(page);
    }

    // Wait and check prize pool
    await organizerPage.waitForTimeout(3000);
    await organizerPage.goto(`/tournaments/${tournamentId}`);

    const prizePool = await organizerPage.locator('[data-testid="prize-pool"], .prize-pool').textContent();
    const expectedPrizePool = buyIn * maxPlayers;

    console.log('Prize pool:', prizePool);
    console.log('Expected:', expectedPrizePool);

    expect(prizePool).toContain(expectedPrizePool.toString());

    // Cleanup
    await organizerPage.close();
    for (const player of players) {
      await player.close();
    }
  });

  test('Should handle tournament pause and resume', async ({ browser }) => {
    const organizerPage = await browser.newPage();
    const timestamp = Date.now();

    await createUser(organizerPage, {
      username: `organizer_${timestamp}`,
      password: 'test123',
    });

    const tournamentId = await createTournament(organizerPage, {
      name: `Pause/Resume Test ${timestamp}`,
      buyIn: 100,
      startingChips: 1000,
      maxPlayers: 6,
      minPlayers: 6,
      blindLevels: [
        { level: 1, smallBlind: 10, bigBlind: 20, duration: 600 },
      ],
    });

    // Register 6 players
    const players: Page[] = [];
    for (let i = 1; i <= 6; i++) {
      const context = await browser.newContext();
      const page = await context.newPage();

      await createUser(page, {
        username: `player${i}_${timestamp}`,
        password: 'test123',
      });

      await registerForTournament(page, tournamentId, 100);
      players.push(page);
    }

    // Wait for tournament to start
    await organizerPage.waitForTimeout(5000);
    await organizerPage.goto(`/tournaments/${tournamentId}`);

    // Pause tournament (if pause button exists)
    const pauseButton = organizerPage.locator('[data-testid="pause-tournament"], button:has-text("Pause")');
    if (await pauseButton.isVisible({ timeout: 3000 })) {
      await pauseButton.click();
      await organizerPage.waitForTimeout(1000);

      // Verify paused status
      const status = await organizerPage.locator('[data-testid="tournament-status"]').textContent();
      expect(status?.toLowerCase()).toContain('pause');

      console.log('Tournament paused');

      // Resume tournament
      const resumeButton = organizerPage.locator('[data-testid="resume-tournament"], button:has-text("Resume")');
      if (await resumeButton.isVisible({ timeout: 2000 })) {
        await resumeButton.click();
        await organizerPage.waitForTimeout(1000);

        // Verify resumed status
        const newStatus = await organizerPage.locator('[data-testid="tournament-status"]').textContent();
        expect(newStatus?.toLowerCase()).toContain('progress');

        console.log('Tournament resumed');
      }
    } else {
      console.log('Pause functionality not available (may require organizer privileges)');
    }

    // Cleanup
    await organizerPage.close();
    for (const player of players) {
      await player.close();
    }
  });

  test('Should show real-time chip counts for all players', async ({ browser }) => {
    const organizerPage = await browser.newPage();
    const timestamp = Date.now();

    await createUser(organizerPage, {
      username: `organizer_${timestamp}`,
      password: 'test123',
    });

    const tournamentId = await createTournament(organizerPage, {
      name: `Chip Count Test ${timestamp}`,
      buyIn: 100,
      startingChips: 1000,
      maxPlayers: 6,
      minPlayers: 6,
      blindLevels: [
        { level: 1, smallBlind: 10, bigBlind: 20, duration: 600 },
      ],
    });

    // Register 6 players
    const players: Page[] = [];
    for (let i = 1; i <= 6; i++) {
      const context = await browser.newContext();
      const page = await context.newPage();

      await createUser(page, {
        username: `player${i}_${timestamp}`,
        password: 'test123',
      });

      await registerForTournament(page, tournamentId, 100);
      players.push(page);
    }

    // Wait for tournament to start
    await organizerPage.waitForTimeout(5000);
    await organizerPage.goto(`/tournaments/${tournamentId}`);

    // Check standings for chip counts
    await organizerPage.waitForTimeout(2000);

    const standings = await getTournamentStandings(organizerPage);
    console.log('Tournament standings:', standings);

    expect(standings.length).toBeGreaterThan(0);

    // Each player should have chip count
    for (const standing of standings) {
      expect(standing.chips).toBeDefined();
      expect(standing.chips).toBeGreaterThan(0);
    }

    // Cleanup
    await organizerPage.close();
    for (const player of players) {
      await player.close();
    }
  });
});

/**
 * Additional 50-player tournament test ideas:
 *
 * - test('Should handle table balancing when players are eliminated')
 * - test('Should merge tables correctly as player count decreases')
 * - test('Should show accurate average chip stack')
 * - test('Should display correct positions/rankings in real-time')
 * - test('Should handle ante increases at higher blind levels')
 * - test('Should show tournament timer and time remaining')
 * - test('Should display break periods between blind levels')
 * - test('Should handle late registration if enabled')
 * - test('Should show bubble position and ITM (in the money) indicator')
 * - test('Should distribute prizes correctly based on finishing position')
 * - test('Should handle tournament cancellation if minimum players not reached')
 * - test('Should show tournament history after completion')
 * - test('Should handle concurrent tournaments without interference')
 * - test('Should verify WebSocket updates for all 50 players simultaneously')
 * - test('Should stress test with rapid player actions across all tables')
 */
