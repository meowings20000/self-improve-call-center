const { test, expect } = require('@playwright/test');

test('learner gets immediate feedback and targeted follow-up', async ({ page }) => {
  await page.goto(process.env.APP_URL || 'http://localhost:3000');
  await expect(page.getByRole('link', { name: 'LingoLift home' })).toBeVisible();
  await expect(page.locator('#prompt')).toContainText('Which sentence is correct?');

  await page.getByRole('radio', { name: /She walk to work every day/ }).click();
  await page.getByRole('button', { name: 'Check answer' }).click();
  await expect(page.locator('#feedbackTitle')).toContainText('Good try');
  await expect(page.locator('#answerReveal')).toContainText('She walks to work every day.');
  await expect(page.locator('#hearts')).toHaveText('4');

  await page.getByRole('button', { name: 'Continue' }).click();
  await expect(page.locator('#skillLabel')).toContainText('Grammar agreement');
  await expect(page.locator('#prompt')).toContainText('My brother');

  await page.locator('[data-view="progress"]').first().click();
  await expect(page.locator('#focusTitle')).toContainText('Grammar agreement');
  await expect(page.locator('#skillGrid')).toContainText('Needs focus');
});

test('mobile learner can leave and re-enter the lesson from the skill map', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto(process.env.APP_URL || 'http://localhost:3000');
  await expect(page.locator('#prompt')).toContainText('Which sentence is correct?');

  const menu = page.getByRole('button', { name: 'Open menu' });
  await menu.click();
  await expect(page.locator('[data-view="learn"]')).toBeVisible();
  await page.locator('[data-view="progress"]').first().click();
  await expect(page.locator('#progressView')).toBeVisible();

  await menu.click();
  await page.locator('[data-view="learn"]').first().click();
  await expect(page.locator('#learnView')).toBeVisible();
  await expect(page.locator('#prompt')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Check answer' })).toBeVisible();
});
