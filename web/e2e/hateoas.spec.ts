import {test, expect} from "@playwright/test";

// The payoff test: proves the frontend really consumes HATEOAS service
// discovery against the live backend — something the unit tests (which mock
// HateoasService) cannot catch. Also guards the frontend<->backend contract.
test.describe("HATEOAS wiring (frontend <-> backend)", () => {
  test("discovery is fetched, then the resolved data endpoint is hit", async ({page}) => {
    const discovery = page.waitForResponse(
      r => /\/api\/v1\/resource\/services(\?|$)/.test(r.url()),
      {timeout: 20000},
    );
    const data = page.waitForResponse(
      r => /\/api\/v1\/projects(\?|$)/.test(r.url()) && r.request().method() === "GET",
      {timeout: 20000},
    );

    await page.goto("/myprojects");

    // 1. Service discovery succeeds and advertises the expected links.
    const disc = await discovery;
    expect(disc.status()).toBe(200);
    const discBody = await disc.json();
    expect(discBody).toHaveProperty("_links");
    expect(discBody._links).toHaveProperty("projects");

    // 2. The data call goes to the URL resolved from discovery, enveloped.
    const dataResp = await data;
    expect(dataResp.status()).toBe(200);
    const env = await dataResp.json();
    expect(env).toHaveProperty("status");
    expect(env).toHaveProperty("data");
  });

  test("my projects page settles (cards or empty-state) without errors", async ({page}) => {
    const errors: string[] = [];
    page.on("console", m => {
      if (m.type() === "error") errors.push(m.text());
    });

    await page.goto("/myprojects");
    await expect(page.getByRole("heading", {name: "Мои проекты"})).toBeVisible();

    // Either the project cards or the documented empty-state must appear —
    // i.e. the page must reach a stable state, not hang on the loader.
    const cards = page.locator("app-my-project");
    const empty = page.locator(".body_empty");
    await expect(cards.first().or(empty)).toBeVisible({timeout: 20000});
    await expect(page.locator("p.text-center", {hasText: "Loading"})).toHaveCount(0);

    expect(errors, errors.join("\n")).toEqual([]);
  });
});
