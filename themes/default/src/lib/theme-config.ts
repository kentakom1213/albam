import { readFileSync } from "node:fs";
import { join } from "node:path";

export type ThemeConfig = {
  siteTitle: string;
  brand: string;
  homeTitle: string;
  homeEyebrow: string;
  homeDescription: string;
  nav: {
    albums: string;
    tags: string;
    settings: string;
  };
};

type RawThemeConfig = {
  title?: string;
  params?: {
    brand?: string;
    home_title?: string;
    home_eyebrow?: string;
    home_description?: string;
    nav?: {
      albums?: string;
      tags?: string;
      settings?: string;
    };
  };
};

type TomlObject = Record<string, string | TomlObject>;

const defaults: ThemeConfig = {
  siteTitle: "albam",
  brand: "albam",
  homeTitle: "Your Albums",
  homeEyebrow: "SELF-HOSTED PHOTO ALBUM",
  homeDescription: "写真をディレクトリやタグごとに，シンプルで可愛いグリッドとして眺められるアルバムです。",
  nav: {
    albums: "Albums",
    tags: "Tags",
    settings: "Settings",
  },
};

export function getThemeConfig(): ThemeConfig {
  const raw = parseThemeToml(readFileSync(join(process.cwd(), "theme.toml"), "utf8"));
  const params = raw.params ?? {};

  return {
    siteTitle: stringValue(raw.title, defaults.siteTitle),
    brand: stringValue(params.brand, defaults.brand),
    homeTitle: stringValue(params.home_title, defaults.homeTitle),
    homeEyebrow: stringValue(params.home_eyebrow, defaults.homeEyebrow),
    homeDescription: stringValue(params.home_description, defaults.homeDescription),
    nav: {
      albums: stringValue(params.nav?.albums, defaults.nav.albums),
      tags: stringValue(params.nav?.tags, defaults.nav.tags),
      settings: stringValue(params.nav?.settings, defaults.nav.settings),
    },
  };
}

function stringValue(value: unknown, fallback: string) {
  return typeof value === "string" && value !== "" ? value : fallback;
}

function parseThemeToml(source: string): RawThemeConfig {
  const parsed: TomlObject = {};
  let section: string[] = [];

  for (const line of source.split(/\r?\n/)) {
    const trimmed = line.trim();

    if (!trimmed || trimmed.startsWith("#")) {
      continue;
    }

    const sectionMatch = trimmed.match(/^\[([A-Za-z0-9_.-]+)\]$/);
    if (sectionMatch) {
      section = sectionMatch[1].split(".");
      continue;
    }

    const assignmentMatch = trimmed.match(/^([A-Za-z0-9_-]+)\s*=\s*"((?:\\"|[^"])*)"$/);
    if (!assignmentMatch) {
      continue;
    }

    assignValue(parsed, section, assignmentMatch[1], assignmentMatch[2].replace(/\\"/g, '"'));
  }

  return parsed as RawThemeConfig;
}

function assignValue(target: TomlObject, section: string[], key: string, value: string) {
  let current = target;

  for (const segment of section) {
    const next = current[segment];
    if (!next || typeof next === "string") {
      current[segment] = {};
    }
    current = current[segment] as TomlObject;
  }

  current[key] = value;
}
