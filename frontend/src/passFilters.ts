import type { PassWindow } from "./api";

export function filterPassesByMinEl(passes: PassWindow[], minEl: number): PassWindow[] {
  return (passes ?? []).filter((p) => p.max_elevation_deg >= minEl);
}

export function sortPassesByAOS(passes: PassWindow[]): PassWindow[] {
  return [...(passes ?? [])].sort(
    (a, b) => new Date(a.aos).getTime() - new Date(b.aos).getTime(),
  );
}

export function sortPassesByElevation(passes: PassWindow[]): PassWindow[] {
  return [...(passes ?? [])].sort((a, b) => b.max_elevation_deg - a.max_elevation_deg);
}
