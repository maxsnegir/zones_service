package geojson

import (
	"encoding/json"
	"io"

	"github.com/twpayne/go-geom"
)

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

func (fg *FeatureGeometryJSON) Decode() (PostgisGeometry, error) {
	if fg.Type == "" {
		return nil, GeometryTypeIsRequiredErr
	}
	if fg.Coordinates == nil {
		return nil, CoordinatesIsRequiredErr
	}
	switch fg.Type {
	case "Polygon":
		var coords [][]geom.Coord
		if err := json.Unmarshal(*fg.Coordinates, &coords); err != nil {
			return nil, NotValidPolygonCoordinatesErr
		}
		polygon, err := geom.NewPolygon(geom.XY).SetCoords(coords)
		if err != nil || polygon.Empty() {
			return nil, NotValidPolygonCoordinatesErr
		}
		return &PostgisPolygon{Polygon: *polygon}, nil

	case "MultiPolygon":
		var coords [][][]geom.Coord
		if err := json.Unmarshal(*fg.Coordinates, &coords); err != nil {
			return nil, NotValidMultiPolygonCoordinatesErr
		}
		multipolygon, err := geom.NewMultiPolygon(geom.XY).SetCoords(coords)
		if err != nil || multipolygon.Empty() {
			return nil, NotValidMultiPolygonCoordinatesErr
		}
		return &PostgisMultiPolygon{MultiPolygon: *multipolygon}, nil

	}
	return nil, UnsupportedGeometryTypeErr{fg.Type}
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
