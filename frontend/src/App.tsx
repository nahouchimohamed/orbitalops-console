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
import { Dashboard } from "./components/Dashboard";
import { GroundTrackMap } from "./components/GroundTrackMap";
import { filterPassesByMinEl, sortPassesByAOS, sortPassesByElevation } from "./passFilters";
import { passQuality } from "./passQuality";
import { nextPass, relativeFromNow } from "./nextPass";
import { downloadText, passesToCSV } from "./exportCsv";

type View = "dashboard" | "ops";

export default function App() {
  const [view, setView] = useState<View>("dashboard");
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
  const [minElevation, setMinElevation] = useState(5);
  const [sortMode, setSortMode] = useState<"aos" | "elevation">("aos");
  const [fleetQuery, setFleetQuery] = useState("");
  const [passHours, setPassHours] = useState(24);
  const [trackHours, setTrackHours] = useState(1.5);
  const [priority, setPriority] = useState(2);

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
        setTrack(await api.groundTrack(selectedSat, trackHours));
      } catch (e) {
        setError(e instanceof Error ? e.message : String(e));
      }
    })();
  }, [selectedSat, trackHours]);

  useEffect(() => {
    if (!selectedStation) return;
    void (async () => {
      try {
        setError(null);
        setPasses((await api.passes(selectedStation, passHours)) ?? []);
      } catch (e) {
        setError(e instanceof Error ? e.message : String(e));
      }
    })();
  }, [selectedStation, passHours]);

  const healthy = Boolean(status?.planning);
  const filteredSats = useMemo(() => {
    const q = fleetQuery.trim().toLowerCase();
    if (!q) return sats;
    return sats.filter((s) => s.name.toLowerCase().includes(q) || String(s.norad_id ?? "").includes(q));
  }, [sats, fleetQuery]);

  const visiblePasses = useMemo(() => {
    const filtered = filterPassesByMinEl(passes, minElevation);
    return sortMode === "elevation" ? sortPassesByElevation(filtered) : sortPassesByAOS(filtered);
  }, [passes, minElevation, sortMode]);

  const upcoming = useMemo(() => nextPass(visiblePasses), [visiblePasses]);

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
        priority,
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
        <nav className="view-nav" aria-label="Primary">
          <button
            type="button"
            className={view === "dashboard" ? "nav-active" : ""}
            onClick={() => setView("dashboard")}
          >
            Dashboard
          </button>
          <button
            type="button"
            className={view === "ops" ? "nav-active" : ""}
            onClick={() => setView("ops")}
          >
            Operations
          </button>
        </nav>
        <div className="header-right">
          {upcoming && (
            <div className="next-pass-banner">
              Next {upcoming.satellite_name} → {upcoming.station_name} · {relativeFromNow(upcoming.aos)}
            </div>
          )}
          <div className="status-pill" data-testid="system-status">
            <span className={`status-dot ${healthy ? "ok" : ""}`} />
            {healthy ? "Planning + Gateway online" : "Waiting for backend…"}
          </div>
        </div>
      </header>

      {!healthy && (
        <div className="alert-strip" role="status">
          Planning service unreachable — pass windows and AOI planning may be stale.
        </div>
      )}

      {view === "dashboard" ? (
        <Dashboard
          status={status}
          sats={sats}
          stations={stations}
          passes={passes}
          plans={plans}
          onOpenOps={() => setView("ops")}
        />
      ) : (
        <div className="layout">
          <aside className="panel">
            <h2>Fleet</h2>
            <label className="filter-row">
              Track horizon {trackHours.toFixed(1)}h
              <input
                type="range"
                min={0.5}
                max={6}
                step={0.5}
                value={trackHours}
                onChange={(e) => setTrackHours(Number(e.target.value))}
              />
            </label>
            <input
              className="text-input"
              placeholder="Search satellites…"
              value={fleetQuery}
              onChange={(e) => setFleetQuery(e.target.value)}
            />
            <div className="list">
              {filteredSats.map((s) => (
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
            {error && (
              <div className="error" role="alert">
                {error}
              </div>
            )}
          </aside>

          <GroundTrackMap track={track} stations={stations} selectedStationId={selectedStation} />

          <aside className="panel">
            <h2>AOS / LOS passes</h2>
            <div className="btn-row">
              {[12, 24, 48].map((h) => (
                <button
                  key={h}
                  type="button"
                  className={passHours === h ? "ghost nav-active" : "ghost"}
                  onClick={() => setPassHours(h)}
                >
                  {h}h
                </button>
              ))}
            </div>
            <label className="filter-row">
              Min elevation {minElevation}°
              <input
                type="range"
                min={0}
                max={40}
                value={minElevation}
                onChange={(e) => setMinElevation(Number(e.target.value))}
              />
            </label>
            <div className="btn-row">
              <button
                type="button"
                className={sortMode === "aos" ? "ghost nav-active" : "ghost"}
                onClick={() => setSortMode("aos")}
              >
                Sort AOS
              </button>
              <button
                type="button"
                className={sortMode === "elevation" ? "ghost nav-active" : "ghost"}
                onClick={() => setSortMode("elevation")}
              >
                Sort elevation
              </button>
              <button
                type="button"
                className="ghost"
                onClick={() =>
                  downloadText(`passes-${passHours}h.csv`, passesToCSV(visiblePasses))
                }
              >
                Export CSV
              </button>
            </div>
            <div className="timeline">
              {visiblePasses.length === 0 && (
                <div className="timeline-item">No passes in the next 24h window.</div>
              )}
              {visiblePasses.slice(0, 8).map((p, idx) => (
                <div className="timeline-item" key={`${p.satellite_id}-${p.aos}-${idx}`}>
                  <strong>{p.satellite_name}</strong>
                  <div className="meta">
                    AOS {new Date(p.aos).toUTCString()}
                    <br />
                    LOS {new Date(p.los).toUTCString()}
                    <br />
                    max el {p.max_elevation_deg.toFixed(1)}°
                    {" · "}
                    <span className={`health-badge ${passQuality(p.max_elevation_deg).className}`}>
                      {passQuality(p.max_elevation_deg).label}
                    </span>
                  </div>
                </div>
              ))}
            </div>

            <h2 style={{ marginTop: "1.25rem" }}>Observation planning</h2>
            <label className="filter-row">
              Priority
              <select
                className="text-input"
                value={priority}
                onChange={(e) => setPriority(Number(e.target.value))}
              >
                <option value={1}>1 — critical</option>
                <option value={2}>2 — high</option>
                <option value={3}>3 — normal</option>
                <option value={4}>4 — background</option>
              </select>
            </label>
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
            <div className="btn-row">
              <button
                type="button"
                className="ghost"
                onClick={() => downloadText("command-sequence.json", exportJSON, "application/json")}
              >
                Download JSON
              </button>
            </div>
            <pre className="cmd">{exportJSON}</pre>
          </aside>
        </div>
      )}
    </div>
  );
}
