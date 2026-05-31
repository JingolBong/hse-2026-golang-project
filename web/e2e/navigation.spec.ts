import {test, expect} from "@playwright/test";

// App shell + client-side routing. No backend data required: the page headings
// render regardless of API responses, so this stays deterministic.
test.describe("app shell & client-side routing", () => {
  test("home renders with header navigation", async ({page}) => {
    await page.goto("/");
    await expect(page.getByRole("heading", {name: "Jira analyzer"})).toBeVisible();

    const nav = page.locator("header nav");
    for (const link of ["Проекты", "Задачи", "Сравнение", "Мои проекты"]) {
      await expect(nav.getByRole("link", {name: link})).toBeVisible();
    }
  });

  const routes = [
    {link: "Проекты", path: "/projects", heading: "Проекты"},
    {link: "Задачи", path: "/issues", heading: "Задачи"},
    {link: "Сравнение", path: "/compare", heading: "Сравнение"},
    {link: "Мои проекты", path: "/myprojects", heading: "Мои проекты"},
  ];

  for (const r of routes) {
    test(`navigates to ${r.path} via "${r.link}"`, async ({page}) => {
      await page.goto("/");
      await page.locator("header nav").getByRole("link", {name: r.link}).click();
      await expect(page).toHaveURL(new RegExp(`${r.path}$`));
      await expect(page.getByRole("heading", {name: r.heading})).toBeVisible();
    });
  }
});
