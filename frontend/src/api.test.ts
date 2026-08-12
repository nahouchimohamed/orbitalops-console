import { describe, expect, it } from "vitest";
import { jakartaAOI } from "./api";

describe("jakartaAOI", () => {
  it("returns a closed polygon around Jakarta", () => {
    const aoi = jakartaAOI();
    expect(aoi.type).toBe("Polygon");
    const ring = aoi.coordinates[0];
    expect(ring[0]).toEqual(ring[ring.length - 1]);
    expect(ring[0][0]).toBeCloseTo(106.7);
  });
});
