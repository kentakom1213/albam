export type Album = {
  id: string;
  title: string;
  kind: string;
  description: string;
  photoCount: number;
  createdAt: string;
  updatedAt: string;
  size: string;
  visibility: "public" | "private";
  tags: string[];
  tone?: "peach" | "linen" | "mint" | "sky" | "lilac" | "lemon";
};

export type Photo = {
  id: string;
  title: string;
  albumId: string;
  src?: string;
  tone?: Album["tone"];
};

export const albums: Album[] = [
  {
    id: "spring-walk",
    title: "Spring walk",
    kind: "DAILY ALBUM",
    description: "春の散歩で見つけたやわらかな光と街角の記録です。",
    photoCount: 24,
    createdAt: "May 02, 2026",
    updatedAt: "May 04, 2026",
    size: "620 MB",
    visibility: "public",
    tags: ["daily", "walk"],
    tone: "peach",
  },
  {
    id: "coffee-time",
    title: "Coffee time",
    kind: "DAILY ALBUM",
    description: "お気に入りのカフェと日常の記録です。",
    photoCount: 12,
    createdAt: "May 12, 2026",
    updatedAt: "May 12, 2026",
    size: "320 MB",
    visibility: "private",
    tags: ["daily", "cafe"],
    tone: "linen",
  },
  {
    id: "tiny-flowers",
    title: "Tiny flowers",
    kind: "FLOWER ALBUM",
    description: "散歩中に見つけた小さな花のアルバムです。",
    photoCount: 36,
    createdAt: "Apr 29, 2026",
    updatedAt: "May 01, 2026",
    size: "860 MB",
    visibility: "public",
    tags: ["flower", "walk"],
    tone: "mint",
  },
  {
    id: "weekend-trip",
    title: "Weekend trip",
    kind: "TRAVEL ALBUM",
    description:
      "友人との週末旅行。海沿いの散歩，カフェ，夕方の空をまとめたアルバムです。",
    photoCount: 48,
    createdAt: "May 18, 2026",
    updatedAt: "May 23, 2026",
    size: "1.2 GB",
    visibility: "public",
    tags: ["travel", "sea", "cafe", "friends", "film"],
    tone: "sky",
  },
  {
    id: "room-light",
    title: "Room light",
    kind: "ROOM ALBUM",
    description: "部屋の光と小物のスナップです。",
    photoCount: 18,
    createdAt: "Apr 21, 2026",
    updatedAt: "Apr 25, 2026",
    size: "410 MB",
    visibility: "private",
    tags: ["room", "daily"],
    tone: "lilac",
  },
  {
    id: "blue-sky",
    title: "Blue sky",
    kind: "SKY ALBUM",
    description: "晴れた日の空を集めたアルバムです。",
    photoCount: 29,
    createdAt: "Apr 18, 2026",
    updatedAt: "Apr 20, 2026",
    size: "700 MB",
    visibility: "public",
    tags: ["sky", "walk"],
    tone: "lemon",
  },
  {
    id: "good-morning",
    title: "Good morning",
    kind: "DAILY ALBUM",
    description: "朝のテーブルと散歩道の記録です。",
    photoCount: 16,
    createdAt: "Apr 11, 2026",
    updatedAt: "Apr 11, 2026",
    size: "390 MB",
    visibility: "private",
    tags: ["daily"],
    tone: "linen",
  },
  {
    id: "park-day",
    title: "Park day",
    kind: "DAILY ALBUM",
    description: "公園で過ごした日の写真です。",
    photoCount: 21,
    createdAt: "Apr 02, 2026",
    updatedAt: "Apr 03, 2026",
    size: "520 MB",
    visibility: "public",
    tags: ["daily", "park"],
    tone: "mint",
  },
  {
    id: "sweet-home",
    title: "Sweet home",
    kind: "HOME ALBUM",
    description: "家で過ごす時間の小さな記録です。",
    photoCount: 32,
    createdAt: "Mar 29, 2026",
    updatedAt: "Apr 01, 2026",
    size: "920 MB",
    visibility: "private",
    tags: ["home", "daily"],
    tone: "peach",
  },
  {
    id: "film-notes",
    title: "Film notes",
    kind: "FILM ALBUM",
    description: "フィルム風に残した日々のメモです。",
    photoCount: 14,
    createdAt: "Mar 20, 2026",
    updatedAt: "Mar 21, 2026",
    size: "280 MB",
    visibility: "public",
    tags: ["film", "daily"],
    tone: "sky",
  },
];

const photoTitles = [
  "beach-001",
  "coffee-002",
  "flower-003",
  "sky-004",
  "room-005",
  "sun-006",
  "street-007",
  "walk-008",
  "green-009",
  "blue-010",
  "purple-011",
  "window-012",
];

const tones: Photo["tone"][] = [
  "peach",
  "linen",
  "mint",
  "sky",
  "lilac",
  "lemon",
  "sky",
  "peach",
  "mint",
  "sky",
  "lilac",
  "peach",
];

export const photos: Photo[] = photoTitles.map((title, index) => ({
  id: `photo-${String(index + 1).padStart(3, "0")}`,
  title,
  albumId: "weekend-trip",
  tone: tones[index],
}));
