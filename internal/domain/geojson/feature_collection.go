package geojson

import (
	"encoding/json"

	"github.com/twpayne/go-geom"

	"github.com/maxsnegir/zones_service/internal/dto"
)

type FeatureCollection struct {
	Type     string
	Features []*Feature
}

type Feature struct {
	Type       string
	Geometry   PostgisGeometry
	Properties map[string]interface{}
}

func (fc *FeatureCollection) FromFeatureCollectionJSON(geojson dto.FeatureCollectionJSON) error {
	if geojson.Type != "FeatureCollection" {
		return NotValidFeatureCollectionType{geojson.Type}
	}

	if geojson.Features == nil || len(geojson.Features) == 0 {
		return FeaturesIsRequiredErr
	}

	features := make([]*Feature, 0, len(geojson.Features))
	for _, feature := range geojson.Features {
		geometry, err := decodeGeometryJson(feature.Geometry)
		if err != nil {
			return err
		}
		features = append(features, &Feature{
			Type:       feature.Type,
			Geometry:   geometry,
			Properties: feature.Properties,
		})
	}
	fc.Type = geojson.Type
	fc.Features = features
	return nil
}

func decodeGeometryJson(fg dto.FeatureGeometryJSON) (PostgisGeometry, error) {
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
	return nil, UnsupportedGeometryTypeErr{T: fg.Type}
}
