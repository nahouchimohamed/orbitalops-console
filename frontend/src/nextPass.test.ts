import { describe, expect, it } from "vitest";
import { nextPass, relativeFromNow } from "./nextPass";
import type { PassWindow } from "./api";

describe("nextPass", () => {
  it("picks soonest future AOS", () => {
    const now = Date.parse("2026-06-01T00:00:00Z");
    const passes = [
      { aos: "2026-06-01T03:00:00Z", los: "2026-06-01T03:10:00Z", satellite_name: "Late" },
      { aos: "2026-06-01T01:00:00Z", los: "2026-06-01T01:10:00Z", satellite_name: "Soon" },
    ] as PassWindow[];
    expect(nextPass(passes, now)?.satellite_name).toBe("Soon");
  });
  it("formats relative time", () => {
    const now = Date.parse("2026-06-01T00:00:00Z");
    expect(relativeFromNow("2026-06-01T00:10:00Z", now)).toBe("in 10 min");
  });
});
