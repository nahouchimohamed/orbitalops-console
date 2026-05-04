import type { PassWindow } from "./api";

export function passesToCSV(passes: PassWindow[]): string {
  const header = "satellite,station,aos,los,max_elevation_deg";
  const rows = (passes ?? []).map((p) => {
    const sat = csvEsc(p.satellite_name);
    const st = csvEsc(p.station_name);
    return `${sat},${st},${p.aos},${p.los},${p.max_elevation_deg.toFixed(2)}`;
  });
  return [header, ...rows].join("\n");
}

function csvEsc(v: string): string {
  if (v.includes(",") || v.includes('"')) return `"${v.replace(/"/g, '""')}"`;
  return v;
}

export function downloadText(filename: string, contents: string, mime = "text/csv"): void {
  const blob = new Blob([contents], { type: mime });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}
