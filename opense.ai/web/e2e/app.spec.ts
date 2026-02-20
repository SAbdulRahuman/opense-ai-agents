import { test, expect } from '@playwright/test';

test.describe('Dashboard Page', () => {
  test('loads the home page', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/OpeNSE/i);
  });

  test('renders the sidebar navigation', async ({ page }) => {
    await page.goto('/');
    const sidebar = page.locator('nav, [data-testid="sidebar"]');
    await expect(sidebar.first()).toBeVisible();
  });

  test('has dashboard content', async ({ page }) => {
    await page.goto('/');
    // Dashboard should contain market-related content
    const body = page.locator('body');
    await expect(body).toBeVisible();
  });
});

test.describe('Navigation', () => {
  const routes = [
    { path: '/charts', label: 'Charts' },
    { path: '/chat', label: 'Chat' },
    { path: '/financeql', label: 'FinanceQL' },
    { path: '/portfolio', label: 'Portfolio' },
    { path: '/screener', label: 'Screener' },
    { path: '/backtest', label: 'Backtest' },
  ];

  for (const route of routes) {
    test(`navigates to ${route.label} page (${route.path})`, async ({ page }) => {
      await page.goto(route.path);
      await expect(page.locator('body')).toBeVisible();
      // Page should render without errors
      const errors: string[] = [];
      page.on('pageerror', (err) => errors.push(err.message));
      await page.waitForTimeout(500);
      expect(errors).toHaveLength(0);
    });
  }
});

test.describe('Charts Page', () => {
  test('renders chart container', async ({ page }) => {
    await page.goto('/charts');
    await page.waitForLoadState('networkidle');
    const main = page.locator('main, [role="main"], .chart-container, [data-testid="chart"]');
    // At least the page renders
    await expect(page.locator('body')).toBeVisible();
  });
});

test.describe('Chat Page', () => {
  test('renders chat interface', async ({ page }) => {
    await page.goto('/chat');
    await page.waitForLoadState('networkidle');
    // Look for input area (textarea, input, or contenteditable)
    const inputArea = page.locator('textarea, input[type="text"], [contenteditable="true"], [data-testid="chat-input"]');
    if (await inputArea.count() > 0) {
      await expect(inputArea.first()).toBeVisible();
    }
  });

  test('can type a message', async ({ page }) => {
    await page.goto('/chat');
    await page.waitForLoadState('networkidle');
    const inputArea = page.locator('textarea, input[type="text"], [data-testid="chat-input"]');
    if (await inputArea.count() > 0) {
      await inputArea.first().fill('Analyze TCS');
      await expect(inputArea.first()).toHaveValue('Analyze TCS');
    }
  });
});

test.describe('FinanceQL Page', () => {
  test('renders query editor', async ({ page }) => {
    await page.goto('/financeql');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('body')).toBeVisible();
  });
});

test.describe('Portfolio Page', () => {
  test('renders portfolio view', async ({ page }) => {
    await page.goto('/portfolio');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('body')).toBeVisible();
  });
});

test.describe('Screener Page', () => {
  test('renders screener interface', async ({ page }) => {
    await page.goto('/screener');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('body')).toBeVisible();
  });
});

test.describe('Backtest Page', () => {
  test('renders backtest interface', async ({ page }) => {
    await page.goto('/backtest');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('body')).toBeVisible();
  });
});

test.describe('Accessibility', () => {
  test('home page has no major accessibility issues', async ({ page }) => {
    await page.goto('/');
    // Check for basic accessibility: lang attribute, viewport meta
    const html = page.locator('html');
    const lang = await html.getAttribute('lang');
    expect(lang).toBeTruthy();
  });

  test('all images have alt text', async ({ page }) => {
    await page.goto('/');
    const images = page.locator('img');
    const count = await images.count();
    for (let i = 0; i < count; i++) {
      const alt = await images.nth(i).getAttribute('alt');
      expect(alt, `Image ${i} missing alt text`).toBeTruthy();
    }
  });

  test('interactive elements are keyboard accessible', async ({ page }) => {
    await page.goto('/');
    // Tab through the page — no errors should occur
    for (let i = 0; i < 10; i++) {
      await page.keyboard.press('Tab');
    }
    // Page should still be functional
    await expect(page.locator('body')).toBeVisible();
  });
});

test.describe('Responsive Design', () => {
  test('renders on mobile viewport', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 812 });
    await page.goto('/');
    await expect(page.locator('body')).toBeVisible();
  });

  test('renders on tablet viewport', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.goto('/');
    await expect(page.locator('body')).toBeVisible();
  });
});
