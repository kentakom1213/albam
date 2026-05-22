import type { Album, Asset, Tag } from "./types";

export type AlbamClientOptions = {
  baseUrl?: string;
};

export class AlbamClient {
  private readonly baseUrl: string;

  constructor(options: AlbamClientOptions = {}) {
    this.baseUrl = options.baseUrl ?? "";
  }

  async getAlbums(): Promise<Album[]> {
    return this.get("/api/albums");
  }

  async getAlbum(slug: string): Promise<Album> {
    return this.get(`/api/albums/${encodeURIComponent(slug)}`);
  }

  async getAlbumAssets(slug: string): Promise<Asset[]> {
    return this.get(`/api/albums/${encodeURIComponent(slug)}/assets`);
  }

  async getTags(): Promise<Tag[]> {
    return this.get("/api/tags");
  }

  async searchAssets(query: string): Promise<Asset[]> {
    return this.get(`/api/search?q=${encodeURIComponent(query)}`);
  }

  private async get<T>(path: string): Promise<T> {
    const res = await fetch(`${this.baseUrl}${path}`);

    if (!res.ok) {
      throw new Error(`Albam API request failed: ${res.status} ${res.statusText}`);
    }

    return res.json() as Promise<T>;
  }
}
