package dto

import (
	"encoding/json"
	"io"
)

type ZoneGeoJSON struct {
	ZoneId  int                   `json:"id"`
	GeoJSON FeatureCollectionJSON `json:"geojson"`
}

type FeatureCollectionJSON struct {
	Type     string        `json:"type"`
	Features []FeatureJSON `json:"features"`
}

type FeatureJSON struct {
	Type       string                 `json:"type"`
	Geometry   FeatureGeometryJSON    `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type FeatureGeometryJSON struct {
	Type        string           `json:"type"`
	Coordinates *json.RawMessage `json:"coordinates"`
}

func NewFeatureCollectionJSON(r io.ReadCloser) (*FeatureCollectionJSON, error) {
	var featureCollection FeatureCollectionJSON

	if err := json.NewDecoder(r).Decode(&featureCollection); err != nil {
		return nil, err
	}
	return &featureCollection, nil
}

func MustNewFeatureCollectionJSON(r io.ReadCloser) *FeatureCollectionJSON {
	featureCollection, err := NewFeatureCollectionJSON(r)
	if err != nil {
		panic(err)
	}
	return featureCollection
}
