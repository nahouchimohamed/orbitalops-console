import { useMemo } from "react";
import DeckGL from "@deck.gl/react";
import { PathLayer, ScatterplotLayer } from "@deck.gl/layers";
import type { Layer } from "@deck.gl/core";
import { Map } from "react-map-gl/maplibre";
import maplibregl from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import type { GroundStation, GroundTrack } from "../api";

const INITIAL_VIEW = {
  longitude: 139.7,
  latitude: 35.7,
  zoom: 2.2,
  pitch: 0,
  bearing: 0,
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
            getColor: [61, 184, 255, 220],
            widthMinPixels: 2,
            getWidth: 3,
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
            getRadius: 60000,
            radiusMinPixels: 5,
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

    return [path, satPoint, gs].filter(Boolean) as Layer[];
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
        {track
          ? `${track.satellite_name}: ${track.samples.length} track samples`
          : "Select a satellite to render ground track"}
      </div>
    </div>
  );
}
