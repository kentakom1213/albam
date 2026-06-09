export type Album = {
  id: string;
  title: string;
  kind: string;
  description: string;
  photoCount: number;
  latestMonth?: string;
  createdAt: string;
  updatedAt: string;
  size: string;
  visibility: "public" | "private";
  tone?: "peach" | "linen" | "mint" | "sky" | "lilac" | "lemon";
};

export type Photo = {
  id: string;
  title: string;
  albumId: string;
  src?: string;
  width: number;
  height: number;
  aspectRatio: number;
  favorite?: boolean;
  tone?: Album["tone"];
};

export const albums: Album[] = [
  {
    id: "spring-walk",
    title: "Spring walk",
    kind: "DAILY ALBUM",
    description: "春の散歩で見つけたやわらかな光と街角の記録です。",
    photoCount: 24,
    latestMonth: "2026/05",
    createdAt: "May 02, 2026",
    updatedAt: "May 04, 2026",
    size: "620 MB",
    visibility: "public",
    tone: "peach",
  },
  {
    id: "coffee-time",
    title: "Coffee time",
    kind: "DAILY ALBUM",
    description: "お気に入りのカフェと日常の記録です。",
    photoCount: 12,
    latestMonth: "2026/05",
    createdAt: "May 12, 2026",
    updatedAt: "May 12, 2026",
    size: "320 MB",
    visibility: "private",
    tone: "linen",
  },
  {
    id: "tiny-flowers",
    title: "Tiny flowers",
    kind: "FLOWER ALBUM",
    description: "散歩中に見つけた小さな花のアルバムです。",
    photoCount: 36,
    latestMonth: "2026/04",
    createdAt: "Apr 29, 2026",
    updatedAt: "May 01, 2026",
    size: "860 MB",
    visibility: "public",
    tone: "mint",
  },
  {
    id: "weekend-trip",
    title: "Weekend trip",
    kind: "TRAVEL ALBUM",
    description:
      "友人との週末旅行。海沿いの散歩，カフェ，夕方の空をまとめたアルバムです。",
    photoCount: 48,
    latestMonth: "2026/05",
    createdAt: "May 18, 2026",
    updatedAt: "May 23, 2026",
    size: "1.2 GB",
    visibility: "public",
    tone: "sky",
  },
  {
    id: "room-light",
    title: "Room light",
    kind: "ROOM ALBUM",
    description: "部屋の光と小物のスナップです。",
    photoCount: 18,
    latestMonth: "2026/04",
    createdAt: "Apr 21, 2026",
    updatedAt: "Apr 25, 2026",
    size: "410 MB",
    visibility: "private",
    tone: "lilac",
  },
  {
    id: "blue-sky",
    title: "Blue sky",
    kind: "SKY ALBUM",
    description: "晴れた日の空を集めたアルバムです。",
    photoCount: 29,
    latestMonth: "2026/04",
    createdAt: "Apr 18, 2026",
    updatedAt: "Apr 20, 2026",
    size: "700 MB",
    visibility: "public",
    tone: "lemon",
  },
  {
    id: "good-morning",
    title: "Good morning",
    kind: "DAILY ALBUM",
    description: "朝のテーブルと散歩道の記録です。",
    photoCount: 16,
    latestMonth: "2026/04",
    createdAt: "Apr 11, 2026",
    updatedAt: "Apr 11, 2026",
    size: "390 MB",
    visibility: "private",
    tone: "linen",
  },
  {
    id: "park-day",
    title: "Park day",
    kind: "DAILY ALBUM",
    description: "公園で過ごした日の写真です。",
    photoCount: 21,
    latestMonth: "2026/04",
    createdAt: "Apr 02, 2026",
    updatedAt: "Apr 03, 2026",
    size: "520 MB",
    visibility: "public",
    tone: "mint",
  },
  {
    id: "sweet-home",
    title: "Sweet home",
    kind: "HOME ALBUM",
    description: "家で過ごす時間の小さな記録です。",
    photoCount: 32,
    latestMonth: "2026/03",
    createdAt: "Mar 29, 2026",
    updatedAt: "Apr 01, 2026",
    size: "920 MB",
    visibility: "private",
    tone: "peach",
  },
  {
    id: "film-notes",
    title: "Film notes",
    kind: "FILM ALBUM",
    description: "フィルム風に残した日々のメモです。",
    photoCount: 14,
    latestMonth: "2026/03",
    createdAt: "Mar 20, 2026",
    updatedAt: "Mar 21, 2026",
    size: "280 MB",
    visibility: "public",
    tone: "sky",
  },
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

const photoShapes = [
  { width: 1600, height: 900 },
  { width: 1400, height: 1000 },
  { width: 1200, height: 900 },
  { width: 900, height: 1200 },
  { width: 1000, height: 1000 },
  { width: 1800, height: 1200 },
  { width: 900, height: 1350 },
  { width: 2000, height: 1100 },
];

export const photos: Photo[] = Array.from({ length: 36 }, (_, index) => {
  const shape = photoShapes[index % photoShapes.length];
  const photoNumber = String(index + 1).padStart(3, "0");

  return {
    id: `photo-${photoNumber}`,
    title: `photo-${photoNumber}`,
    albumId: "weekend-trip",
    width: shape.width,
    height: shape.height,
    aspectRatio: shape.width / shape.height,
    favorite: index % 7 === 0,
    tone: tones[index % tones.length],
  };
});
