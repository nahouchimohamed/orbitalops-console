import type { ImagingPlan, PassWindow, Satellite, GroundStation } from "../api";
import { passQuality } from "../passQuality";
import { planStatusColor } from "../planColors";

function durationMinutes(aos: string, los: string): number {
  const ms = new Date(los).getTime() - new Date(aos).getTime();
  if (!Number.isFinite(ms) || ms <= 0) return 0;
  return ms / 60000;
}

function hourBucket(iso: string): number {
  return new Date(iso).getUTCHours();
}

/** Horizontal bar chart: max elevation of upcoming passes */
export function ElevationBars({ passes }: { passes: PassWindow[] }) {
  const rows = (passes ?? []).slice(0, 8);
  const maxEl = Math.max(90, ...rows.map((p) => p.max_elevation_deg));
  const W = 420;
  const rowH = 28;
  const H = Math.max(rowH * Math.max(rows.length, 1) + 28, 80);
  const left = 118;
  const right = 36;
  const barW = W - left - right;

  return (
    <svg viewBox={`0 0 ${W} ${H}`} className="chart-svg" role="img" aria-label="Pass max elevation">
      <text x={0} y={14} className="chart-axis-label">
        Max elevation (°)
      </text>
      {rows.length === 0 && (
        <text x={left} y={48} className="chart-empty">
          No passes to plot
        </text>
      )}
      {rows.map((p, i) => {
        const y = 28 + i * rowH;
        const w = (p.max_elevation_deg / maxEl) * barW;
        const label = `${p.satellite_name.split(" ")[0]}→${p.station_name.split(" ")[0]}`;
        return (
          <g key={`${p.satellite_id}-${p.aos}-${i}`}>
            <text x={left - 8} y={y + 14} textAnchor="end" className="chart-tick">
              {label.length > 16 ? `${label.slice(0, 15)}…` : label}
            </text>
            <rect x={left} y={y + 2} width={barW} height={18} rx={4} className="chart-track" />
            <rect
              x={left}
              y={y + 2}
              width={Math.max(w, 2)}
              height={18}
              rx={4}
              className={
                passQuality(p.max_elevation_deg).label === "excellent" ||
                passQuality(p.max_elevation_deg).label === "good"
                  ? "chart-bar-teal"
                  : "chart-bar-accent"
              }
            />
            <text x={left + w + 6} y={y + 15} className="chart-value">
              {p.max_elevation_deg.toFixed(0)}°
            </text>
          </g>
        );
      })}
    </svg>
  );
}

/** 24h contact-load histogram by UTC hour */
export function ContactLoadHistogram({ passes }: { passes: PassWindow[] }) {
  const counts = Array.from({ length: 24 }, () => 0);
  for (const p of passes ?? []) {
    counts[hourBucket(p.aos)] += 1;
  }
  const peak = Math.max(1, ...counts);
  const W = 420;
  const H = 160;
  const padL = 28;
  const padR = 12;
  const padT = 20;
  const padB = 28;
  const plotW = W - padL - padR;
  const plotH = H - padT - padB;
  const barGap = 2;
  const barW = plotW / 24 - barGap;

  return (
    <svg viewBox={`0 0 ${W} ${H}`} className="chart-svg" role="img" aria-label="Contact load by UTC hour">
      <text x={padL} y={14} className="chart-axis-label">
        Contacts by UTC hour · peak {peak} at {String(counts.indexOf(peak)).padStart(2, "0")}z
      </text>
      {[0, 0.5, 1].map((t) => {
        const y = padT + plotH * (1 - t);
        return (
          <g key={t}>
            <line x1={padL} x2={W - padR} y1={y} y2={y} className="chart-gridline" />
            <text x={padL - 6} y={y + 3} textAnchor="end" className="chart-tick">
              {Math.round(peak * t)}
            </text>
          </g>
        );
      })}
      {counts.map((c, h) => {
        const x = padL + h * (barW + barGap);
        const bh = (c / peak) * plotH;
        const y = padT + plotH - bh;
        return (
          <g key={h}>
            <rect
              x={x}
              y={y}
              width={barW}
              height={Math.max(bh, c > 0 ? 2 : 0)}
              rx={2}
              className={c > 0 ? "chart-bar-teal" : "chart-bar-muted"}
            />
            {h % 3 === 0 && (
              <text x={x + barW / 2} y={H - 8} textAnchor="middle" className="chart-tick">
                {String(h).padStart(2, "0")}
              </text>
            )}
          </g>
        );
      })}
    </svg>
  );
}

/** Gantt-style contact windows over the next 24h */
export function ContactGantt({ passes }: { passes: PassWindow[] }) {
  const rows = (passes ?? []).slice(0, 6);
  const now = Date.now();
  const horizon = 24 * 3600 * 1000;
  const W = 420;
  const rowH = 26;
  const H = Math.max(rowH * Math.max(rows.length, 1) + 36, 90);
  const left = 110;
  const right = 12;
  const plotW = W - left - right;
  const top = 22;

  return (
    <svg viewBox={`0 0 ${W} ${H}`} className="chart-svg" role="img" aria-label="Contact window timeline">
      <text x={left} y={14} className="chart-axis-label">
        Next 24h contact windows
      </text>
      {[0, 0.25, 0.5, 0.75, 1].map((t) => {
        const x = left + t * plotW;
        return (
          <g key={t}>
            <line x1={x} x2={x} y1={top} y2={H - 10} className="chart-gridline" />
            <text x={x} y={H - 2} textAnchor="middle" className="chart-tick">
              +{Math.round(t * 24)}h
            </text>
          </g>
        );
      })}
      {rows.length === 0 && (
        <text x={left} y={top + 28} className="chart-empty">
          No windows scheduled
        </text>
      )}
      {rows.map((p, i) => {
        const aos = new Date(p.aos).getTime();
        const los = new Date(p.los).getTime();
        const x0 = left + Math.max(0, Math.min(1, (aos - now) / horizon)) * plotW;
        const x1 = left + Math.max(0, Math.min(1, (los - now) / horizon)) * plotW;
        const y = top + i * rowH;
        const label = p.satellite_name.split(" ")[0];
        return (
          <g key={`${p.satellite_id}-${p.aos}-${i}`}>
            <text x={left - 8} y={y + 14} textAnchor="end" className="chart-tick">
              {label}
            </text>
            <rect x={left} y={y + 4} width={plotW} height={16} rx={4} className="chart-track" />
            <rect
              x={x0}
              y={y + 4}
              width={Math.max(x1 - x0, 3)}
              height={16}
              rx={4}
              className="chart-bar-warn"
            />
          </g>
        );
      })}
    </svg>
  );
}

/** Donut of imaging plan statuses */
export function PlanStatusDonut({ plans }: { plans: ImagingPlan[] }) {
  const list = plans ?? [];
  const buckets = new Map<string, number>();
  for (const p of list) {
    const key = (p.status || "unknown").toLowerCase();
    buckets.set(key, (buckets.get(key) ?? 0) + 1);
  }
  if (buckets.size === 0) buckets.set("empty", 1);

  const entries = [...buckets.entries()];
  const colors = entries.map(([label]) => planStatusColor(label));
  const total = entries.reduce((a, [, n]) => a + n, 0);
  const cx = 78;
  const cy = 78;
  const R = 54;
  const r = 34;
  let angle = -Math.PI / 2;

  const arcs = entries.map(([label, n], i) => {
    const sweep = (n / total) * Math.PI * 2;
    const a0 = angle;
    const a1 = angle + sweep;
    angle = a1;
    const large = sweep > Math.PI ? 1 : 0;
    const x0 = cx + R * Math.cos(a0);
    const y0 = cy + R * Math.sin(a0);
    const x1 = cx + R * Math.cos(a1);
    const y1 = cy + R * Math.sin(a1);
    const xi0 = cx + r * Math.cos(a0);
    const yi0 = cy + r * Math.sin(a0);
    const xi1 = cx + r * Math.cos(a1);
    const yi1 = cy + r * Math.sin(a1);
    const d = [
      `M ${x0} ${y0}`,
      `A ${R} ${R} 0 ${large} 1 ${x1} ${y1}`,
      `L ${xi1} ${yi1}`,
      `A ${r} ${r} 0 ${large} 0 ${xi0} ${yi0}`,
      "Z",
    ].join(" ");
    return { d, label, n, color: colors[i % colors.length] };
  });

  return (
    <svg viewBox="0 0 280 160" className="chart-svg" role="img" aria-label="Imaging plan status mix">
      {arcs.map((a) => (
        <path key={a.label} d={a.d} fill={a.color} opacity={list.length === 0 ? 0.35 : 0.92} />
      ))}
      <text x={cx} y={cy - 4} textAnchor="middle" className="chart-donut-value">
        {list.length}
      </text>
      <text x={cx} y={cy + 14} textAnchor="middle" className="chart-tick">
        plans
      </text>
      {entries.map(([label, n], i) => (
        <g key={label} transform={`translate(168 ${28 + i * 22})`}>
          <rect width={10} height={10} rx={2} fill={colors[i]} />
          <text x={16} y={9} className="chart-tick">
            {label} · {list.length === 0 ? 0 : n}
          </text>
        </g>
      ))}
    </svg>
  );
}

/** Station utilization: pass count + avg duration */
export function StationUtilization({
  passes,
  stations,
}: {
  passes: PassWindow[];
  stations: GroundStation[];
}) {
  const byStation = new Map<string, { name: string; count: number; mins: number }>();
  for (const s of stations ?? []) {
    byStation.set(s.id, { name: s.name, count: 0, mins: 0 });
  }
  for (const p of passes ?? []) {
    const cur = byStation.get(p.station_id) ?? { name: p.station_name, count: 0, mins: 0 };
    cur.count += 1;
    cur.mins += durationMinutes(p.aos, p.los);
    byStation.set(p.station_id, cur);
  }
  const rows = [...byStation.values()].sort((a, b) => b.count - a.count);
  const maxC = Math.max(1, ...rows.map((r) => r.count));
  const W = 420;
  const rowH = 32;
  const H = Math.max(rowH * Math.max(rows.length, 1) + 24, 80);
  const left = 100;
  const barW = W - left - 80;

  return (
    <svg viewBox={`0 0 ${W} ${H}`} className="chart-svg" role="img" aria-label="Station utilization">
      <text x={0} y={14} className="chart-axis-label">
        Passes per station (24h)
      </text>
      {rows.map((r, i) => {
        const y = 24 + i * rowH;
        const w = (r.count / maxC) * barW;
        const avg = r.count ? r.mins / r.count : 0;
        return (
          <g key={r.name}>
            <text x={left - 8} y={y + 16} textAnchor="end" className="chart-tick">
              {r.name.length > 12 ? `${r.name.slice(0, 11)}…` : r.name}
            </text>
            <rect x={left} y={y + 4} width={barW} height={18} rx={4} className="chart-track" />
            <rect x={left} y={y + 4} width={Math.max(w, r.count ? 4 : 0)} height={18} rx={4} className="chart-bar-accent" />
            <text x={left + barW + 8} y={y + 17} className="chart-value">
              {r.count} · {avg.toFixed(0)}m
            </text>
          </g>
        );
      })}
    </svg>
  );
}

/** Fleet overview: satellites as orbit rings (schematic) */
export function FleetOrbitSchematic({ sats }: { sats: Satellite[] }) {
  const list = (sats ?? []).slice(0, 6);
  const cx = 140;
  const cy = 90;
  return (
    <svg viewBox="0 0 280 180" className="chart-svg" role="img" aria-label="Fleet orbit schematic">
      <circle cx={cx} cy={cy} r={22} className="chart-earth" />
      <text x={cx} y={cy + 4} textAnchor="middle" className="chart-tick">
        Earth
      </text>
      {list.length === 0 && (
        <text x={cx} y={160} textAnchor="middle" className="chart-empty">
          No satellites in catalog
        </text>
      )}
      {list.map((s, i) => {
        const R = 38 + i * 14;
        const ang = (-Math.PI / 2) + (i * 0.9);
        const x = cx + R * Math.cos(ang);
        const y = cy + R * Math.sin(ang);
        return (
          <g key={s.id}>
            <circle cx={cx} cy={cy} r={R} fill="none" className="chart-orbit-ring" />
            <circle cx={x} cy={y} r={5} className="chart-sat-dot" />
            <text x={x + 8} y={y + 3} className="chart-tick">
              {s.name.length > 14 ? `${s.name.slice(0, 13)}…` : s.name}
            </text>
          </g>
        );
      })}
    </svg>
  );
}
