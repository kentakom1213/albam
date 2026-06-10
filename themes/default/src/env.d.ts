/// <reference path="../.astro/types.d.ts" />

declare module "node:fs" {
  export function readFileSync(path: string, encoding: string): string;
}
