export function formatDurationMin(aos: string, los: string): string {
  const ms = new Date(los).getTime() - new Date(aos).getTime();
  if (!Number.isFinite(ms) || ms <= 0) return "—";
  return `${Math.round(ms / 60000)} min`;
}

export function formatElevation(deg: number): string {
  if (!Number.isFinite(deg)) return "—";
  return `${deg.toFixed(1)}°`;
}
