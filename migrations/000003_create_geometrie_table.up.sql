CREATE TABLE IF NOT EXISTS zone_geometry
(
    id   SERIAL PRIMARY KEY,
    zone_id int REFERENCES zone (id),
    geom GEOMETRY,
    properties json
);
CREATE INDEX IF NOT EXISTS zone_geometry_geom_idx ON zone_geometry USING GIST (geom);
