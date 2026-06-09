import { readFileSync } from "node:fs";
import { join } from "node:path";

export type ThemeConfig = {
  siteTitle: string;
  brand: string;
  accent: {
    name: AccentName;
    color: string;
    soft: string;
  };
  homeTitle: string;
  homeEyebrow: string;
  homeDescription: string;
  layout: {
    albumGridColumns: number;
    albumGridColumnsTablet: number;
    albumGridColumnsMobile: number;
    photoGridColumns: number;
    photoGridColumnsTablet: number;
    photoGridColumnsMobile: number;
  };
  nav: {
    albums: string;
  };
  header: {
    enabled: boolean;
  };
  footer: {
    text: string;
    poweredBy: boolean;
  };
  favicon: {
    href: string;
    type: string;
  };
};

type AccentName = "pink" | "coral" | "mint" | "blue" | "lavender" | "lemon" | "red" | "beige";

type RawThemeConfig = {
  title?: string;
  params?: {
    brand?: string;
    accent?: string;
    home_title?: string;
    home_eyebrow?: string;
    home_description?: string;
    album_grid_columns?: number;
    photo_grid_columns?: number;
    nav?: {
      albums?: string;
    };
    header?: {
      enabled?: boolean;
    };
    footer?: {
      text?: string;
      powered_by?: boolean;
    };
    favicon?: {
      href?: string;
      type?: string;
    };
  };
};

type TomlObject = Record<string, string | number | boolean | TomlObject>;

const accents: Record<AccentName, { color: string; soft: string }> = {
  pink: { color: "#ff6fae", soft: "#ffe3ef" },
  coral: { color: "#ff6b5f", soft: "#ffe6e2" },
  mint: { color: "#35c99b", soft: "#dff8ef" },
  blue: { color: "#4da3ff", soft: "#e3f1ff" },
  lavender: { color: "#9b7cff", soft: "#eee8ff" },
  lemon: { color: "#f4c430", soft: "#fff5c7" },
  red: { color: "#f04438", soft: "#ffe4e0" },
  beige: { color: "#c88a5a", soft: "#f4e8dd" },
};

const defaults: ThemeConfig = {
  siteTitle: "albam",
  brand: "albam",
  accent: {
    name: "pink",
    ...accents.pink,
  },
  homeTitle: "Your Albums",
  homeEyebrow: "SELF-HOSTED PHOTO ALBUM",
  homeDescription: "写真をディレクトリごとに，シンプルで可愛いグリッドとして眺められるアルバムです。",
  layout: {
    albumGridColumns: 5,
    albumGridColumnsTablet: 3,
    albumGridColumnsMobile: 2,
    photoGridColumns: 6,
    photoGridColumnsTablet: 3,
    photoGridColumnsMobile: 2,
  },
  nav: {
    albums: "Albums",
  },
  header: {
    enabled: true,
  },
  footer: {
    text: "",
    poweredBy: true,
  },
  favicon: {
    href: "/favicon.svg",
    type: "image/svg+xml",
  },
};

export function getThemeConfig(): ThemeConfig {
  const raw = parseThemeToml(readFileSync(join(process.cwd(), "theme.toml"), "utf8"));
  const params = raw.params ?? {};
  const albumGridColumns = numberValue(params.album_grid_columns, defaults.layout.albumGridColumns, 1, 8);
  const photoGridColumns = numberValue(params.photo_grid_columns, defaults.layout.photoGridColumns, 1, 10);
  const accentName = accentValue(params.accent, defaults.accent.name);
  const accent = accents[accentName];

  return {
    siteTitle: stringValue(raw.title, defaults.siteTitle),
    brand: stringValue(params.brand, defaults.brand),
    accent: {
      name: accentName,
      ...accent,
    },
    homeTitle: stringValue(params.home_title, defaults.homeTitle),
    homeEyebrow: stringValue(params.home_eyebrow, defaults.homeEyebrow),
    homeDescription: stringValue(params.home_description, defaults.homeDescription),
    layout: {
      albumGridColumns,
      albumGridColumnsTablet: Math.min(albumGridColumns, defaults.layout.albumGridColumnsTablet),
      albumGridColumnsMobile: Math.min(albumGridColumns, defaults.layout.albumGridColumnsMobile),
      photoGridColumns,
      photoGridColumnsTablet: Math.min(photoGridColumns, defaults.layout.photoGridColumnsTablet),
      photoGridColumnsMobile: Math.min(photoGridColumns, defaults.layout.photoGridColumnsMobile),
    },
    nav: {
      albums: stringValue(params.nav?.albums, defaults.nav.albums),
    },
    header: {
      enabled: booleanValue(params.header?.enabled, defaults.header.enabled),
    },
    footer: {
      text: stringValue(params.footer?.text, defaults.footer.text),
      poweredBy: booleanValue(params.footer?.powered_by, defaults.footer.poweredBy),
    },
    favicon: {
      href: stringValue(params.favicon?.href, defaults.favicon.href),
      type: stringValue(params.favicon?.type, defaults.favicon.type),
    },
  };
}

export function formatPageTitle(...parts: string[]) {
  return parts.filter((part) => part !== "").join(" | ");
}

function stringValue(value: unknown, fallback: string) {
  return typeof value === "string" ? value : fallback;
}

function numberValue(value: unknown, fallback: number, min: number, max: number) {
  return typeof value === "number" && Number.isInteger(value)
    ? Math.min(Math.max(value, min), max)
    : fallback;
}

function booleanValue(value: unknown, fallback: boolean) {
  return typeof value === "boolean" ? value : fallback;
}

function accentValue(value: unknown, fallback: AccentName): AccentName {
  return typeof value === "string" && value in accents ? (value as AccentName) : fallback;
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

    const stringMatch = trimmed.match(/^([A-Za-z0-9_-]+)\s*=\s*"((?:\\"|[^"])*)"$/);
    if (stringMatch) {
      assignValue(parsed, section, stringMatch[1], stringMatch[2].replace(/\\"/g, '"'));
      continue;
    }

    const numberMatch = trimmed.match(/^([A-Za-z0-9_-]+)\s*=\s*([0-9]+)$/);
    if (numberMatch) {
      assignValue(parsed, section, numberMatch[1], Number.parseInt(numberMatch[2], 10));
      continue;
    }

    const booleanMatch = trimmed.match(/^([A-Za-z0-9_-]+)\s*=\s*(true|false)$/);
    if (booleanMatch) {
      assignValue(parsed, section, booleanMatch[1], booleanMatch[2] === "true");
    }
  }

  return parsed as RawThemeConfig;
}

function assignValue(target: TomlObject, section: string[], key: string, value: string | number | boolean) {
  let current = target;

  for (const segment of section) {
    const next = current[segment];
    if (!next || typeof next !== "object") {
      current[segment] = {};
    }
    current = current[segment] as TomlObject;
  }

  current[key] = value;
}
