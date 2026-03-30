import { test, expect } from '@playwright/test';

/**
 * E2E Tests for Dashboard VM Management
 * 
 * Tests that the VM dashboard routes are accessible.
 */

test.describe('Dashboard - VM Management', () => {
  test('VM dashboard route should be accessible', async ({ page }) => {
    const response = await page.goto('/dashboard/-vms');
    
    // Check that the page responded
    expect(response?.ok() || response?.status() === 200).toBeTruthy();
  });

  test('should have valid HTML response for VM dashboard', async ({ page }) => {
    await page.goto('/dashboard/-vms');
    
    // Check that we got HTML content
    const contentType = await page.evaluate(() => document.contentType);
    expect(contentType).toContain('html');
  });

  test('should handle navigation to VM dashboard', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    // Navigate to VM dashboard
    await page.goto('/dashboard/-vms');
    await page.waitForTimeout(1000);
    
    // URL should contain dashboard path OR redirect to auth (GitHub login)
    const currentUrl = page.url();
    expect(currentUrl.includes('/dashboard/') || currentUrl.includes('github.com/login')).toBeTruthy();
  });
});
