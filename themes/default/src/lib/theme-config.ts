import { readFileSync } from "node:fs";

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
  };
  features: {
    showHeader: boolean;
    showFooter: boolean;
    showTags: boolean;
    showAlbumCount: boolean;
  };
  footer: {
    copyright: string;
    text: string;
  };
};

type AccentName = "pink" | "coral" | "mint" | "blue" | "lavender" | "lemon" | "red" | "sakura";

type RawThemePayload = {
  site?: {
    title?: unknown;
  };
  theme?: {
    params?: RawThemeParams;
  };
};

type RawThemeParams = {
  appearance?: {
    accent?: unknown;
  };
  layout?: {
    album_grid_columns?: unknown;
  };
  features?: {
    show_header?: unknown;
    show_footer?: unknown;
    show_tags?: unknown;
    show_album_count?: unknown;
  };
  content?: {
    brand?: unknown;
    home_title?: unknown;
    home_eyebrow?: unknown;
    home_description?: unknown;
    copyright?: unknown;
    footer_text?: unknown;
  };
};

const accents: Record<AccentName, { color: string; soft: string }> = {
  pink: { color: "#ff6fae", soft: "#ffe3ef" },
  coral: { color: "#ff6b5f", soft: "#ffe6e2" },
  mint: { color: "#35c99b", soft: "#dff8ef" },
  blue: { color: "#4da3ff", soft: "#e3f1ff" },
  lavender: { color: "#9b7cff", soft: "#eee8ff" },
  lemon: { color: "#f4c430", soft: "#fff5c7" },
  red: { color: "#f04438", soft: "#ffe4e0" },
  sakura: { color: "#ff6fae", soft: "#ffe3ef" },
};

const defaults: ThemeConfig = {
  siteTitle: "albam",
  brand: "albam",
  accent: {
    name: "coral",
    ...accents.coral,
  },
  homeTitle: "Your Albums",
  homeEyebrow: "SELF-HOSTED PHOTO ALBUM",
  homeDescription: "写真をディレクトリごとに，シンプルで可愛いグリッドとして眺められるアルバムです。",
  layout: {
    albumGridColumns: 5,
    albumGridColumnsTablet: 3,
    albumGridColumnsMobile: 2,
  },
  features: {
    showHeader: true,
    showFooter: true,
    showTags: true,
    showAlbumCount: true,
  },
  footer: {
    copyright: "",
    text: "",
  },
};

export function getThemeConfig(): ThemeConfig {
  const raw = loadRawThemePayload();
  const params = raw.theme?.params ?? {};
  const albumGridColumns = numberValue(params.layout?.album_grid_columns, defaults.layout.albumGridColumns, 1, 8);
  const accentName = accentValue(params.appearance?.accent, defaults.accent.name);
  const accent = accents[accentName];

  return {
    siteTitle: stringValue(raw.site?.title, defaults.siteTitle),
    brand: stringValue(params.content?.brand, defaults.brand),
    accent: {
      name: accentName,
      ...accent,
    },
    homeTitle: stringValue(params.content?.home_title, defaults.homeTitle),
    homeEyebrow: stringValue(params.content?.home_eyebrow, defaults.homeEyebrow),
    homeDescription: stringValue(params.content?.home_description, defaults.homeDescription),
    layout: {
      albumGridColumns,
      albumGridColumnsTablet: Math.min(albumGridColumns, defaults.layout.albumGridColumnsTablet),
      albumGridColumnsMobile: Math.min(albumGridColumns, defaults.layout.albumGridColumnsMobile),
    },
    features: {
      showHeader: booleanValue(params.features?.show_header, defaults.features.showHeader),
      showFooter: booleanValue(params.features?.show_footer, defaults.features.showFooter),
      showTags: booleanValue(params.features?.show_tags, defaults.features.showTags),
      showAlbumCount: booleanValue(params.features?.show_album_count, defaults.features.showAlbumCount),
    },
    footer: {
      copyright: stringValue(params.content?.copyright, defaults.footer.copyright),
      text: stringValue(params.content?.footer_text, defaults.footer.text),
    },
  };
}

export function formatPageTitle(...parts: string[]) {
  return parts.filter((part) => part !== "").join(" | ");
}

export function siteTitlePrefix(theme: ThemeConfig) {
  return theme.brand === "" ? "" : theme.siteTitle;
}

function loadRawThemePayload(): RawThemePayload {
  const configFile = import.meta.env.ALBAM_THEME_CONFIG_FILE as string | undefined;
  if (!configFile) {
    return {};
  }

  try {
    return JSON.parse(readFileSync(configFile, "utf8")) as RawThemePayload;
  } catch (error) {
    console.error(error);
    return {};
  }
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
