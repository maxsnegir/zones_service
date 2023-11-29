package geojson

type ZoneGEOJSON struct {
	ZoneId  int                   `json:"id"`
	GeoJSON FeatureCollectionJSON `json:"geojson"`
}
