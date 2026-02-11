import { describe, expect, it } from "vitest";
import { tokyoAOI } from "./api";

describe("tokyoAOI", () => {
  it("returns a closed polygon around Tokyo", () => {
    const aoi = tokyoAOI();
    expect(aoi.type).toBe("Polygon");
    const ring = aoi.coordinates[0];
    expect(ring[0]).toEqual(ring[ring.length - 1]);
    expect(ring[0][0]).toBeCloseTo(139.6);
  });
});
