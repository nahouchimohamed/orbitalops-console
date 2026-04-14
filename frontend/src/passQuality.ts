export function passQuality(maxEl: number): { label: string; className: string } {
  if (maxEl >= 40) return { label: "excellent", className: "ok" };
  if (maxEl >= 20) return { label: "good", className: "ok" };
  if (maxEl >= 10) return { label: "fair", className: "" };
  return { label: "marginal", className: "" };
}
