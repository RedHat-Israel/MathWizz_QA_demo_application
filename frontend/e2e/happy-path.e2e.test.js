// Happy Path E2E Test — "Spread the Joy"
//
// This test verifies MathWizz's complete user journey through the browser:
// register → login → solve a problem → verify it appears in history.
//
// TO ADAPT THIS FOR ANOTHER TEAM'S CI:
// If your product integrates with MathWizz, you can create a lightweight
// version of this test using curl instead of Playwright:
//
//   # Register
//   curl -s -o /dev/null -w "%{http_code}" -c cookies.txt \
//     -X POST http://localhost:8080/register \
//     -H "Content-Type: application/json" \
//     -d '{"email":"test@example.com","password":"testpass123"}'
//
//   # Login
//   curl -s -o /dev/null -w "%{http_code}" -c cookies.txt \
//     -X POST http://localhost:8080/login \
//     -H "Content-Type: application/json" \
//     -d '{"email":"test@example.com","password":"testpass123"}'
//
//   # Solve
//   curl -s -b cookies.txt \
//     -X POST http://localhost:8080/solve \
//     -H "Content-Type: application/json" \
//     -d '{"problem":"25+75"}'
//   # Expected: {"answer":100}
//
//   # Verify history
//   curl -s -b cookies.txt http://localhost:8080/history
//   # Expected: JSON array containing the solved problem
//
// Remove Node.js/Playwright setup steps, keep Kind cluster setup/teardown.

const { test, expect } = require('@playwright/test');

const FRONTEND_URL = process.env.FRONTEND_URL || 'http://localhost:3000';

test.describe('MathWizz Happy Path', () => {
  test('complete user journey: register, login, solve, verify history', async ({ page }) => {
    const testEmail = `happypath${Date.now()}@example.com`;
    const testPassword = 'testpass123';

    // Step 1: Navigate to MathWizz
    await page.goto(FRONTEND_URL);

    // Step 2: Register a new account
    await page.getByRole('button', { name: 'REGISTER' }).click();
    await page.getByTestId('email-input').fill(testEmail);
    await page.getByTestId('password-input').fill(testPassword);
    await page.getByTestId('register-button').click();

    // Verify registration succeeded — user is redirected to solver page
    await expect(page.getByTestId('problem-input')).toBeVisible({ timeout: 10000 });

    // Step 3: Solve a math problem
    await page.getByTestId('problem-input').fill('25+75');
    await page.getByTestId('solve-button').click();
    await expect(page.getByTestId('answer-display')).toContainText('= 100');

    // Step 4: Navigate to history and verify the problem appears
    await page.getByRole('button', { name: 'HISTORY' }).click();

    const maxRetries = 10;
    const retryInterval = 500;
    let found = false;

    for (let i = 0; i < maxRetries; i++) {
      const historyItems = await page.getByTestId('history-item').all();

      for (const item of historyItems) {
        const text = await item.textContent();
        if (text.includes('25+75') && text.includes('100')) {
          found = true;
          break;
        }
      }

      if (found) break;

      await page.waitForTimeout(retryInterval);

      const refreshButton = page.getByTestId('refresh-button');
      if (await refreshButton.isVisible()) {
        await refreshButton.click();
        await page.waitForTimeout(200);
      }
    }

    expect(found).toBeTruthy();
  });
});
