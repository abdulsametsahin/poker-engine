import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for poker engine E2E tests
 *
 * Test suites:
 * - Heads-up games (2 players)
 * - 3-way games (3 players)
 * - 50-player tournaments
 */
export default defineConfig({
  testDir: './tests/e2e',

  // Run tests in parallel for speed
  fullyParallel: true,

  // Fail CI build on committed .only tests
  forbidOnly: !!process.env.CI,

  // Retry failed tests in CI
  retries: process.env.CI ? 2 : 0,

  // Limit workers in CI to prevent resource exhaustion
  workers: process.env.CI ? 1 : undefined,

  // Reporter to use
  reporter: [
    ['html'],
    ['list'],
    ['json', { outputFile: 'test-results.json' }],
  ],

  // Shared settings for all tests
  use: {
    // Base URL for the app
    baseURL: process.env.BASE_URL || 'http://localhost:3000',

    // Collect trace on first retry for debugging
    trace: 'on-first-retry',

    // Screenshot on failure
    screenshot: 'only-on-failure',

    // Video on failure
    video: 'retain-on-failure',

    // Maximum time for each action
    actionTimeout: 10000,

    // Navigation timeout
    navigationTimeout: 30000,
  },

  // Test projects for different browsers
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },

    // Uncomment to test on other browsers
    // {
    //   name: 'firefox',
    //   use: { ...devices['Desktop Firefox'] },
    // },
    // {
    //   name: 'webkit',
    //   use: { ...devices['Desktop Safari'] },
    // },
  ],

  // Start dev server before running tests
  webServer: {
    command: 'npm run start',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
    timeout: 120000, // 2 minutes to start
  },

  // Global timeout for each test
  timeout: 60000, // 1 minute per test

  // Expect timeout for assertions
  expect: {
    timeout: 10000, // 10 seconds for assertions
  },
});
