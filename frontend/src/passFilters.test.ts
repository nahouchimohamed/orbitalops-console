import { describe, expect, it } from "vitest";
import { filterPassesByMinEl, sortPassesByElevation } from "./passFilters";
import type { PassWindow } from "./api";

const sample = [
  { max_elevation_deg: 10, aos: "2026-05-01T02:00:00Z", los: "2026-05-01T02:10:00Z", satellite_name: "A", station_name: "S", satellite_id: "1", station_id: "2", max_elevation_at: "" },
  { max_elevation_deg: 40, aos: "2026-05-01T01:00:00Z", los: "2026-05-01T01:10:00Z", satellite_name: "B", station_name: "S", satellite_id: "1", station_id: "2", max_elevation_at: "" },
] as PassWindow[];

describe("pass filters", () => {
  it("filters by elevation", () => {
    expect(filterPassesByMinEl(sample, 20)).toHaveLength(1);
  });
  it("sorts by elevation desc", () => {
    expect(sortPassesByElevation(sample)[0].max_elevation_deg).toBe(40);
  });
});
