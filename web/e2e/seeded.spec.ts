import {test, expect} from "@playwright/test";

// Opt-in deep happy-path. Runs only when E2E_SEED_PROJECT is set to the name of
// a project that is already tracked in the DB (seeded, or added via the UI), so
// the suite stays green on an empty database.
const SEED = process.env["E2E_SEED_PROJECT"];

test.describe("seeded read happy-path (opt-in)", () => {
  test.skip(!SEED, "set E2E_SEED_PROJECT to a tracked project name to run");

  test("open a tracked project and see its stats", async ({page}) => {
    await page.goto("/myprojects");

    const card = page.locator("app-my-project", {hasText: SEED!});
    await expect(card).toBeVisible({timeout: 20000});

    // The default (non-settings) view shows the stats table and the action button.
    await expect(card.getByText("Общее количество задач")).toBeVisible();
    await expect(card.getByRole("button", {name: "Посмотреть"})).toBeVisible();
  });
});
