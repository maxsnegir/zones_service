package dto

type ZoneGeoJSON struct {
	ZoneId  int                   `json:"id"`
	GeoJSON FeatureCollectionJSON `json:"geojson"`
}
