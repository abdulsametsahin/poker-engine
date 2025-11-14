import { Page, expect } from '@playwright/test';

/**
 * Helper functions for authentication in E2E tests
 */

export interface UserCredentials {
  username: string;
  password: string;
  email?: string;
}

/**
 * Create a new user account
 */
export async function createUser(page: Page, credentials: UserCredentials) {
  const { username, password, email } = credentials;

  await page.goto('/login');

  // Switch to register tab
  const registerTab = page.locator('button:has-text("Register"), [data-testid="register-tab"]');
  if (await registerTab.isVisible()) {
    await registerTab.click();
  }

  // Fill registration form
  await page.fill('[data-testid="username-input"], input[name="username"]', username);
  await page.fill('[data-testid="email-input"], input[name="email"]', email || `${username}@test.com`);
  await page.fill('[data-testid="password-input"], input[name="password"]', password);

  // Submit
  await page.click('[data-testid="register-button"], button[type="submit"]:has-text("Register")');

  // Wait for navigation to lobby
  await page.waitForURL('**/lobby', { timeout: 10000 });

  return { username, password, email: email || `${username}@test.com` };
}

/**
 * Login with existing user credentials
 */
export async function loginUser(page: Page, username: string, password: string) {
  await page.goto('/login');

  // Fill login form
  await page.fill('[data-testid="username-input"], input[name="username"]', username);
  await page.fill('[data-testid="password-input"], input[name="password"]', password);

  // Submit
  await page.click('[data-testid="login-button"], button[type="submit"]:has-text("Login")');

  // Wait for navigation to lobby
  await page.waitForURL('**/lobby', { timeout: 10000 });
}

/**
 * Logout current user
 */
export async function logoutUser(page: Page) {
  // Look for logout button in header/menu
  const logoutButton = page.locator('[data-testid="logout-button"], button:has-text("Logout")');

  if (await logoutButton.isVisible()) {
    await logoutButton.click();
  } else {
    // Try opening menu first
    const menuButton = page.locator('[data-testid="user-menu"], [data-testid="profile-menu"]');
    if (await menuButton.isVisible()) {
      await menuButton.click();
      await page.click('[data-testid="logout-button"], button:has-text("Logout")');
    }
  }

  // Wait for redirect to login
  await page.waitForURL('**/login', { timeout: 5000 });
}

/**
 * Check if user is authenticated
 */
export async function isAuthenticated(page: Page): Promise<boolean> {
  try {
    // Check if we're on lobby page or game page
    const url = page.url();
    return url.includes('/lobby') || url.includes('/game') || url.includes('/tournaments');
  } catch {
    return false;
  }
}
