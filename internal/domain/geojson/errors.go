package geojson

import (
	"errors"
	"fmt"
)

type UnsupportedGeometryTypeErr struct {
	T string
}

func (e UnsupportedGeometryTypeErr) Error() string {
	return fmt.Sprintf("unsupported geometry type: %s", e.T)
}

type NotValidFeatureCollectionType struct {
	T string
}

func (e NotValidFeatureCollectionType) Error() string {
	return fmt.Sprintf("not valid feature collection type: %s", e.T)
}

type NotValidFeatureType struct {
	T string
}

func (e NotValidFeatureType) Error() string {
	return fmt.Sprintf("not valid feature type: %s", e.T)
}

var (
	SerializationErr                   = errors.New("serialization error")
	FeaturesIsRequiredErr              = errors.New("features is required")
	GeometryTypeIsRequiredErr          = errors.New("geometry is required")
	CoordinatesIsRequiredErr           = errors.New("coordinates is required")
	NotValidPolygonCoordinatesErr      = errors.New("not valid polygon coordinates")
	NotValidMultiPolygonCoordinatesErr = errors.New("not valid multipolygon coordinates")
)
