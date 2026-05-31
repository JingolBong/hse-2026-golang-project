import {defineConfig, devices} from "@playwright/test";

// Drives the running app end-to-end against a real backend. The stack is
// external (bring it up via `docker compose up`, or `ng serve` + a dockerized
// backend) and E2E_BASE_URL points at the served frontend. See e2e/README.md.
export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: !!process.env["CI"],
  retries: process.env["CI"] ? 1 : 0,
  reporter: [["list"], ["html", {open: "never"}]],
  use: {
    baseURL: process.env["E2E_BASE_URL"] || "http://localhost:4200",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
  projects: [{name: "chromium", use: {...devices["Desktop Chrome"]}}],
});
