const API_BASE = import.meta.env.VITE_API_BASE?.replace(/\/$/, "") || "";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json", ...(init?.headers || {}) },
    ...init,
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || res.statusText);
  }
  return res.json() as Promise<T>;
}

export type Satellite = {
  id: string;
  name: string;
  norad_id?: number;
  tle_line1: string;
  tle_line2: string;
};

export type GroundStation = {
  id: string;
  name: string;
  latitude: number;
  longitude: number;
  altitude_m: number;
  min_elevation_deg: number;
};

export type TrackSample = {
  time: string;
  latitude: number;
  longitude: number;
  altitude_km: number;
};

export type GroundTrack = {
  satellite_id: string;
  satellite_name: string;
  samples: TrackSample[];
};

export type PassWindow = {
  satellite_id: string;
  satellite_name: string;
  station_id: string;
  station_name: string;
  aos: string;
  los: string;
  max_elevation_deg: number;
  max_elevation_at: string;
};

export type ImagingPlan = {
  id: string;
  request_id: string;
  satellite_id: string;
  satellite_name?: string;
  start_time: string;
  end_time: string;
  look_side: string;
  status: string;
  command_sequence: Array<{ time: string; command: string; args?: Record<string, unknown> }>;
};

export type SystemStatus = {
  gateway: string;
  planning: boolean;
  orbit: string;
  time: string;
};

export const api = {
  status: () => request<SystemStatus>("/api/v1/system/status"),
  satellites: () => request<Satellite[]>("/api/v1/satellites"),
  stations: () => request<GroundStation[]>("/api/v1/ground-stations"),
  groundTrack: (id: string, hours = 1.5) =>
    request<GroundTrack>(`/api/v1/satellites/${id}/ground-track?hours=${hours}&step_seconds=60`),
  passes: (stationId: string, hours = 24) =>
    request<PassWindow[]>(`/api/v1/ground-stations/${stationId}/passes?hours=${hours}`),
  plans: () => request<ImagingPlan[]>("/api/v1/plans"),
  createRequest: (body: { name: string; aoi_geojson: object; priority?: number }) =>
    request<{ request: unknown; plans: ImagingPlan[] }>("/api/v1/requests", {
      method: "POST",
      body: JSON.stringify(body),
    }),
};

export function tokyoAOI() {
  return {
    type: "Polygon",
    coordinates: [
      [
        [139.6, 35.6],
        [139.85, 35.6],
        [139.85, 35.85],
        [139.6, 35.85],
        [139.6, 35.6],
      ],
    ],
  };
}
