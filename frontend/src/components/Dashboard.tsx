import type {
  GroundStation,
  ImagingPlan,
  PassWindow,
  Satellite,
  SystemStatus,
} from "../api";
import {
  ContactGantt,
  ContactLoadHistogram,
  ElevationBars,
  FleetOrbitSchematic,
  PlanStatusDonut,
  StationUtilization,
} from "./DashboardCharts";

type Props = {
  status: SystemStatus | null;
  sats: Satellite[];
  stations: GroundStation[];
  passes: PassWindow[];
  plans: ImagingPlan[];
  onOpenOps: () => void;
};

function formatDuration(aos: string, los: string): string {
  const ms = new Date(los).getTime() - new Date(aos).getTime();
  if (!Number.isFinite(ms) || ms <= 0) return "—";
  const mins = Math.round(ms / 60000);
  return `${mins} min`;
}

function statusLabel(status: SystemStatus | null): { label: string; ok: boolean } {
  if (!status) return { label: "Offline", ok: false };
  const ok = Boolean(status.planning);
  return { label: ok ? "Nominal" : "Degraded", ok };
}

export function Dashboard({ status, sats, stations, passes, plans, onOpenOps }: Props) {
  const health = statusLabel(status);
  const allPasses = passes ?? [];
  const upcoming = allPasses.slice(0, 5);
  const recentPlans = (plans ?? []).slice(0, 5);
  const plannedCount = (plans ?? []).filter(
    (p) => p.status === "planned" || p.status === "scheduled",
  ).length;
  const avgMaxEl =
    upcoming.length > 0
      ? upcoming.reduce((acc, p) => acc + p.max_elevation_deg, 0) / upcoming.length
      : 0;

  return (
    <div className="dashboard" data-testid="ops-dashboard">
      <div className="dash-hero">
        <div>
          <h2>Mission overview</h2>
          <p>
            Constellation health, contact analytics, and imaging queue — charts update from live
            planning and orbit services.
          </p>
        </div>
        <button type="button" className="primary" onClick={onOpenOps}>
          Open operations map
        </button>
      </div>

      <div className="kpi-grid">
        <div className="kpi-card">
          <span className="kpi-label">Satellites</span>
          <strong className="kpi-value">{sats.length}</strong>
          <span className="kpi-meta">active in fleet catalog</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-label">Ground stations</span>
          <strong className="kpi-value">{stations.length}</strong>
          <span className="kpi-meta">contact sites online</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-label">Passes (24h)</span>
          <strong className="kpi-value">{allPasses.length}</strong>
          <span className="kpi-meta">
            avg max el {upcoming.length ? `${avgMaxEl.toFixed(1)}°` : "—"}
          </span>
        </div>
        <div className="kpi-card">
          <span className="kpi-label">Imaging plans</span>
          <strong className="kpi-value">{(plans ?? []).length}</strong>
          <span className="kpi-meta">{plannedCount} queued / scheduled</span>
        </div>
      </div>

      <div className="chart-grid">
        <section className="dash-section chart-panel">
          <h3>Pass elevation</h3>
          <ElevationBars passes={allPasses} />
        </section>
        <section className="dash-section chart-panel">
          <h3>Contact load (UTC)</h3>
          <ContactLoadHistogram passes={allPasses} />
        </section>
        <section className="dash-section chart-panel">
          <h3>Contact timeline</h3>
          <ContactGantt passes={allPasses} />
        </section>
        <section className="dash-section chart-panel">
          <h3>Plan status mix</h3>
          <PlanStatusDonut plans={plans ?? []} />
        </section>
        <section className="dash-section chart-panel">
          <h3>Station utilization</h3>
          <StationUtilization passes={allPasses} stations={stations} />
        </section>
        <section className="dash-section chart-panel">
          <h3>Fleet schematic</h3>
          <FleetOrbitSchematic sats={sats} />
        </section>
      </div>

      <div className="dash-columns">
        <section className="dash-section">
          <h3>Service health</h3>
          <div className="health-list">
            <div className="health-row">
              <span>Overall</span>
              <span className={`health-badge ${health.ok ? "ok" : ""}`}>{health.label}</span>
            </div>
            <div className="health-row">
              <span>API gateway</span>
              <span className="health-badge ok">{status?.gateway ?? "unknown"}</span>
            </div>
            <div className="health-row">
              <span>Planning service</span>
              <span className={`health-badge ${status?.planning ? "ok" : ""}`}>
                {status?.planning ? "reachable" : "unreachable"}
              </span>
            </div>
            <div className="health-row">
              <span>Orbit service</span>
              <span className="health-badge ok">{status?.orbit ?? "unknown"}</span>
            </div>
            <div className="health-row">
              <span>Last status sync</span>
              <span className="kpi-meta">
                {status?.time ? new Date(status.time).toUTCString() : "—"}
              </span>
            </div>
          </div>
        </section>

        <section className="dash-section">
          <h3>Next AOS / LOS</h3>
          <div className="timeline">
            {upcoming.length === 0 && (
              <div className="timeline-item">No contact windows in the next 24h.</div>
            )}
            {upcoming.map((p, idx) => (
              <div className="timeline-item" key={`${p.satellite_id}-${p.aos}-${idx}`}>
                <strong>
                  {p.satellite_name} → {p.station_name}
                </strong>
                <div className="meta">
                  AOS {new Date(p.aos).toUTCString()}
                  <br />
                  Duration {formatDuration(p.aos, p.los)} · max el {p.max_elevation_deg.toFixed(1)}°
                </div>
              </div>
            ))}
          </div>
        </section>

        <section className="dash-section">
          <h3>Imaging queue</h3>
          <div className="timeline">
            {recentPlans.length === 0 && (
              <div className="timeline-item">No imaging plans yet. Create one from Operations.</div>
            )}
            {recentPlans.map((p) => (
              <div className="timeline-item" key={p.id}>
                <strong>
                  {p.satellite_name ?? p.satellite_id.slice(0, 8)} · {p.status}
                </strong>
                <div className="meta">
                  {new Date(p.start_time).toISOString()} → {new Date(p.end_time).toISOString()}
                  <br />
                  look {p.look_side} · {p.command_sequence?.length ?? 0} commands
                </div>
              </div>
            ))}
          </div>
        </section>
      </div>
    </div>
  );
}
