import { getRuntimeThemeContent } from "./api";
import type { RuntimeThemeContent } from "./api";

export async function applyRuntimeContent(options: { page?: "home" | "album" } = {}) {
  try {
    const content = await getRuntimeThemeContent();
    applySharedContent(content);

    if (options.page === "home") {
      applyHomeContent(content);
    }

    return content;
  } catch (error) {
    console.error(error);
    return undefined;
  }
}

export function titlePrefix(content: RuntimeThemeContent | undefined, fallback: string) {
  if (!content) {
    return fallback;
  }

  return content.brand === "" ? "" : content.siteTitle;
}

function applySharedContent(content: RuntimeThemeContent) {
  const brand = document.querySelector<HTMLAnchorElement>("[data-theme-brand]");
  const brandText = document.querySelector<HTMLElement>("[data-theme-brand-text]");
  const footerCopyright = document.querySelector<HTMLElement>("[data-theme-footer-copyright]");
  const footerText = document.querySelector<HTMLElement>("[data-theme-footer-text]");
  const footerContentSeparator = document.querySelector<HTMLElement>("[data-theme-footer-content-separator]");
  const footerPoweredSeparator = document.querySelector<HTMLElement>("[data-theme-footer-powered-separator]");
  const hasCopyright = content.copyright !== "";
  const hasFooterText = content.footerText !== "";

  if (brand) {
    brand.hidden = content.brand === "";
    brand.setAttribute("aria-label", `${content.brand} home`);
  }
  if (brandText) {
    brandText.textContent = content.brand;
  }
  if (footerCopyright) {
    footerCopyright.hidden = !hasCopyright;
    footerCopyright.textContent = content.copyright;
  }
  if (footerText) {
    footerText.hidden = !hasFooterText;
    footerText.textContent = content.footerText;
  }
  if (footerContentSeparator) {
    footerContentSeparator.hidden = !(hasCopyright && hasFooterText);
  }
  if (footerPoweredSeparator) {
    footerPoweredSeparator.hidden = !(hasCopyright || hasFooterText);
  }
}

function applyHomeContent(content: RuntimeThemeContent) {
  document.title = formatPageTitle(titlePrefix(content, ""), content.homeTitle);

  const homePage = document.querySelector<HTMLElement>(".home-page");
  homePage?.style.setProperty("--home-title-length", String(Math.max(Array.from(content.homeTitle).length, 1)));

  for (const element of document.querySelectorAll<HTMLElement>("[data-theme-home-eyebrow]")) {
    element.textContent = content.homeEyebrow;
  }
  for (const element of document.querySelectorAll<HTMLElement>("[data-theme-home-description]")) {
    element.textContent = content.homeDescription;
  }

  const homeTitle = document.querySelector<HTMLElement>("[data-theme-home-title]");
  if (homeTitle) {
    homeTitle.textContent = content.homeTitle;
  }
}

function formatPageTitle(...parts: string[]) {
  return parts.filter((part) => part !== "").join(" | ");
}
