import type { PassWindow } from "./api";

export function nextPass(passes: PassWindow[], now = Date.now()): PassWindow | null {
  const upcoming = (passes ?? [])
    .filter((p) => new Date(p.aos).getTime() >= now)
    .sort((a, b) => new Date(a.aos).getTime() - new Date(b.aos).getTime());
  return upcoming[0] ?? null;
}

export function relativeFromNow(iso: string, now = Date.now()): string {
  const t = new Date(iso).getTime();
  if (!Number.isFinite(t)) return "—";
  const mins = Math.round((t - now) / 60000);
  if (Math.abs(mins) < 1) return "now";
  if (mins > 0) return `in ${mins} min`;
  return `${Math.abs(mins)} min ago`;
}
