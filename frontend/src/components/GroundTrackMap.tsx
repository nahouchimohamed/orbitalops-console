import { useMemo } from "react";
import DeckGL from "@deck.gl/react";
import { PathLayer, ScatterplotLayer } from "@deck.gl/layers";
// station horizon visualized as translucent scatter radius
import type { Layer } from "@deck.gl/core";
import { Map } from "react-map-gl/maplibre";
import maplibregl from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import type { GroundStation, GroundTrack } from "../api";

const INITIAL_VIEW = {
  longitude: 80.0,
  latitude: 15.0,
  zoom: 1.85,
  pitch: 20,
  bearing: -10,
};

type Props = {
  track: GroundTrack | null;
  stations: GroundStation[];
  selectedStationId: string | null;
};

export function GroundTrackMap({ track, stations, selectedStationId }: Props) {
  const layers = useMemo(() => {
    const path =
      track && track.samples.length > 1
        ? new PathLayer({
            id: "ground-track",
            data: [
              {
                path: track.samples.map((s) => [s.longitude, s.latitude] as [number, number]),
              },
            ],
            getPath: (d: { path: [number, number][] }) => d.path,
            getColor: [61, 184, 255, 230],
            widthMinPixels: 2,
            getWidth: 3.5,
            capRounded: true,
            jointRounded: true,
          })
        : null;

    const satPoint =
      track && track.samples.length
        ? new ScatterplotLayer({
            id: "sat-pos",
            data: [track.samples[track.samples.length - 1]],
            getPosition: (d: { longitude: number; latitude: number }) =>
              [d.longitude, d.latitude] as [number, number],
            getFillColor: [92, 224, 178, 255],
            getLineColor: [11, 18, 32, 255],
            lineWidthMinPixels: 1,
            stroked: true,
            getRadius: 70000,
            radiusMinPixels: 6,
          })
        : null;

    const selected = stations.find((s) => s.id === selectedStationId);
    const horizon = selected
      ? new ScatterplotLayer({
          id: "station-horizon",
          data: [selected],
          getPosition: (d: GroundStation) => [d.longitude, d.latitude] as [number, number],
          getFillColor: [240, 195, 90, 40],
          getLineColor: [240, 195, 90, 180],
          lineWidthMinPixels: 1,
          stroked: true,
          filled: true,
          getRadius: 450000,
          radiusMinPixels: 18,
        })
      : null;

    const gs = new ScatterplotLayer({
      id: "ground-stations",
      data: stations,
      getPosition: (d: GroundStation) => [d.longitude, d.latitude] as [number, number],
      getFillColor: (d: GroundStation) =>
        d.id === selectedStationId ? [240, 195, 90, 255] : [154, 173, 200, 220],
      getRadius: 80000,
      radiusMinPixels: 6,
    });

    return [horizon, path, satPoint, gs].filter(Boolean) as Layer[];
  }, [track, stations, selectedStationId]);

  return (
    <div className="map-wrap">
      <DeckGL
        initialViewState={INITIAL_VIEW}
        controller={true}
        layers={layers}
        style={{ position: "absolute", inset: "0", width: "100%", height: "100%" }}
      >
        <Map
          reuseMaps
          mapLib={maplibregl as unknown as never}
          mapStyle="https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json"
          attributionControl={false}
        />
      </DeckGL>
      <div className="map-overlay">
        {track && track.samples.length > 1
          ? `${track.satellite_name}: ${track.samples.length} samples · ${new Date(track.samples[0].time).toUTCString().slice(17, 25)}–${new Date(track.samples[track.samples.length - 1].time).toUTCString().slice(17, 25)} UTC`
          : track
            ? `${track.satellite_name}: awaiting samples`
            : "Select a satellite to render ground track"}
      </div>
    </div>
  );
}
