import { test, expect } from "@playwright/test";

test("home renders brand", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByText("OrbitalOps Console")).toBeVisible();
});
