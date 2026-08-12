import { useCallback, useEffect, useMemo, useState } from "react";
import {
  api,
  jakartaAOI,
  type GroundStation,
  type GroundTrack,
  type ImagingPlan,
  type PassWindow,
  type Satellite,
  type SystemStatus,
} from "./api";
import { GroundTrackMap } from "./components/GroundTrackMap";

export default function App() {
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [sats, setSats] = useState<Satellite[]>([]);
  const [stations, setStations] = useState<GroundStation[]>([]);
  const [selectedSat, setSelectedSat] = useState<string | null>(null);
  const [selectedStation, setSelectedStation] = useState<string | null>(null);
  const [track, setTrack] = useState<GroundTrack | null>(null);
  const [passes, setPasses] = useState<PassWindow[]>([]);
  const [plans, setPlans] = useState<ImagingPlan[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadBase = useCallback(async () => {
    try {
      setError(null);
      const [st, satellites, gs, existingPlans] = await Promise.all([
        api.status(),
        api.satellites(),
        api.stations(),
        api.plans(),
      ]);
      setStatus(st);
      setSats(satellites ?? []);
      setStations(gs ?? []);
      setPlans(existingPlans ?? []);
      if (!selectedSat && satellites?.[0]) setSelectedSat(satellites[0].id);
      if (!selectedStation && gs?.[0]) setSelectedStation(gs[0].id);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    }
  }, [selectedSat, selectedStation]);

  useEffect(() => {
    void loadBase();
  }, []);

  useEffect(() => {
    if (!selectedSat) return;
    void (async () => {
      try {
        setError(null);
        setTrack(await api.groundTrack(selectedSat));
      } catch (e) {
        setError(e instanceof Error ? e.message : String(e));
      }
    })();
  }, [selectedSat]);

  useEffect(() => {
    if (!selectedStation) return;
    void (async () => {
      try {
        setError(null);
        setPasses((await api.passes(selectedStation)) ?? []);
      } catch (e) {
        setError(e instanceof Error ? e.message : String(e));
      }
    })();
  }, [selectedStation]);

  const healthy = Boolean(status?.planning);

  const exportJSON = useMemo(() => {
    const selected = (plans ?? []).slice(0, 2);
    return JSON.stringify(
      {
        exported_at: new Date().toISOString(),
        plans: selected.map((p) => ({
          satellite: p.satellite_name,
          window: { start: p.start_time, end: p.end_time },
          commands: p.command_sequence,
        })),
      },
      null,
      2,
    );
  }, [plans]);

  async function createObservation() {
    setBusy(true);
    try {
      setError(null);
      const res = await api.createRequest({
        name: `AOI-Jakarta-${new Date().toISOString().slice(0, 16)}`,
        aoi_geojson: jakartaAOI(),
        priority: 2,
      });
      setPlans((prev) => [...(res.plans ?? []), ...prev]);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="app">
      <header className="header">
        <div className="brand">
          <h1>OrbitalOps Console</h1>
          <span>Spacecraft control lite — constellation mission planning</span>
        </div>
        <div className="status-pill" data-testid="system-status">
          <span className={`status-dot ${healthy ? "ok" : ""}`} />
          {healthy ? "Planning + Gateway online" : "Waiting for backend…"}
        </div>
      </header>

      <div className="layout">
        <aside className="panel">
          <h2>Fleet</h2>
          <div className="list">
            {sats.map((s) => (
              <button
                key={s.id}
                className={`list-item ${selectedSat === s.id ? "active" : ""}`}
                onClick={() => setSelectedSat(s.id)}
              >
                <strong>{s.name}</strong>
                <small>NORAD {s.norad_id ?? "—"}</small>
              </button>
            ))}
          </div>

          <h2 style={{ marginTop: "1.25rem" }}>Ground stations</h2>
          <div className="list">
            {stations.map((g) => (
              <button
                key={g.id}
                className={`list-item ${selectedStation === g.id ? "active" : ""}`}
                onClick={() => setSelectedStation(g.id)}
              >
                <strong>{g.name}</strong>
                <small>
                  {g.latitude.toFixed(2)}, {g.longitude.toFixed(2)} · el≥{g.min_elevation_deg}°
                </small>
              </button>
            ))}
          </div>
          {error && <div className="error">{error}</div>}
        </aside>

        <GroundTrackMap track={track} stations={stations} selectedStationId={selectedStation} />

        <aside className="panel">
          <h2>AOS / LOS passes</h2>
          <div className="timeline">
            {(passes ?? []).length === 0 && (
              <div className="timeline-item">No passes in the next 24h window.</div>
            )}
            {(passes ?? []).slice(0, 8).map((p, idx) => (
              <div className="timeline-item" key={`${p.satellite_id}-${p.aos}-${idx}`}>
                <strong>{p.satellite_name}</strong>
                <div className="meta">
                  AOS {new Date(p.aos).toUTCString()}
                  <br />
                  LOS {new Date(p.los).toUTCString()}
                  <br />
                  max el {p.max_elevation_deg.toFixed(1)}°
                </div>
              </div>
            ))}
          </div>

          <h2 style={{ marginTop: "1.25rem" }}>Observation planning</h2>
          <div className="btn-row">
            <button className="primary" disabled={busy} onClick={() => void createObservation()}>
              {busy ? "Planning…" : "Request Jakarta AOI"}
            </button>
            <button className="ghost" onClick={() => void loadBase()}>
              Refresh
            </button>
          </div>

          <div className="timeline">
            {(plans ?? []).slice(0, 6).map((p) => (
              <div className="timeline-item" key={p.id}>
                <strong>
                  {p.satellite_name ?? p.satellite_id.slice(0, 8)} · {p.status}
                </strong>
                <div className="meta">
                  {new Date(p.start_time).toISOString()} → {new Date(p.end_time).toISOString()}
                </div>
              </div>
            ))}
          </div>

          <h2 style={{ marginTop: "1.25rem" }}>Command sequence export</h2>
          <pre className="cmd">{exportJSON}</pre>
        </aside>
      </div>
    </div>
  );
}
