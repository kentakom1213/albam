export type Album = {
  id: string;
  slug: string;
  path: string;
  title: string;
  description?: string;
  parentId?: string;
  coverAsset?: AssetSummary;
};

export type Asset = {
  id: string;
  albumId: string;
  filename: string;
  title?: string;
  description?: string;
  width: number;
  height: number;
  takenAt?: string;
  tags: Tag[];
  variants: AssetVariants;
};

export type AssetSummary = {
  id: string;
  title?: string;
  variants: AssetVariants;
};

export type AssetVariants = {
  thumb: string;
  medium: string;
  large?: string;
  original?: string;
};

export type Tag = {
  id: string;
  name: string;
  slug: string;
};
