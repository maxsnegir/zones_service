package geojson

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-geom"
)

func TestFeatureCollection_FromFeatureCollectionJSON_Ok(t *testing.T) {

	polygonCoords := json.RawMessage(`[[[37.829325, 55.696803], [37.830308, 55.687199], [37.85839, 55.67523], [37.829325, 55.696803]]]`)
	multiPolygonCoords := json.RawMessage(`[
        [[[37.829325,55.696803],[37.830308,55.687199],[37.85839,55.67523],[37.829325,55.696803]]],
        [[[37.829325,55.696803],[37.830308,55.687199],[37.85839,55.67523],[37.829325,55.696803]]]
    ]`)

	featureCollectionJSON := FeatureCollectionJSON{
		Type: "FeatureCollection",
		Features: []FeatureJSON{
			{
				Type: "Feature",
				Geometry: FeatureGeometryJSON{
					Type:        "Polygon",
					Coordinates: &polygonCoords,
				},
				Properties: map[string]interface{}{
					"key": "value",
				},
			},
			{
				Type: "Feature",
				Geometry: FeatureGeometryJSON{
					Type:        "MultiPolygon",
					Coordinates: &multiPolygonCoords,
				},
				Properties: map[string]interface{}{
					"color": "white",
				},
			},
		},
	}
	expectedFeatureCollection := FeatureCollection{
		Type: "FeatureCollection",
		Features: []*Feature{
			{
				Type: "Feature",
				Geometry: &PostgisPolygon{Polygon: *geom.NewPolygon(geom.XY).MustSetCoords([][]geom.Coord{
					{
						{37.829325, 55.696803},
						{37.830308, 55.687199},
						{37.85839, 55.67523},
						{37.829325, 55.696803},
					},
				})},
				Properties: map[string]interface{}{
					"key": "value",
				},
			},
			{
				Type: "Feature",
				Geometry: &PostgisMultiPolygon{MultiPolygon: *geom.NewMultiPolygon(geom.XY).MustSetCoords([][][]geom.Coord{
					{
						{
							{37.829325, 55.696803},
							{37.830308, 55.687199},
							{37.85839, 55.67523},
							{37.829325, 55.696803},
						},
					},
					{
						{
							{37.829325, 55.696803},
							{37.830308, 55.687199},
							{37.85839, 55.67523},
							{37.829325, 55.696803},
						},
					},
				})},
				Properties: map[string]interface{}{
					"color": "white",
				},
			},
		}}

	var featureCollection FeatureCollection
	err := featureCollection.FromFeatureCollectionJSON(featureCollectionJSON)

	require.NoError(t, err)
	require.EqualValues(t, expectedFeatureCollection, featureCollection)
}

func TestFeatureCollection_FromFeatureCollectionJSON_Err(t *testing.T) {
	wrongCoords := json.RawMessage(`[]`)
	tests := []struct {
		name        string
		featureCol  FeatureCollectionJSON
		expectedErr error
	}{
		{
			name:        "wrong feature collection type",
			featureCol:  FeatureCollectionJSON{Type: "NotFeatureCollection"},
			expectedErr: NotValidFeatureCollectionType{"NotFeatureCollection"},
		},
		{
			name:        "features not passed",
			featureCol:  FeatureCollectionJSON{Type: "FeatureCollection"},
			expectedErr: FeaturesIsRequiredErr,
		},
		{
			name:        "empty feature",
			featureCol:  FeatureCollectionJSON{Type: "FeatureCollection", Features: []FeatureJSON{}},
			expectedErr: FeaturesIsRequiredErr,
		},
		{
			name: "empty geometry type",
			featureCol: FeatureCollectionJSON{
				Type: "FeatureCollection",
				Features: []FeatureJSON{
					{
						Type:     "Feature",
						Geometry: FeatureGeometryJSON{Type: ""},
					},
				},
			},
			expectedErr: GeometryTypeIsRequiredErr,
		},
		{
			name: "coordinates not passed",
			featureCol: FeatureCollectionJSON{
				Type: "FeatureCollection",
				Features: []FeatureJSON{
					{
						Type:     "Feature",
						Geometry: FeatureGeometryJSON{Type: "Polygon"},
					},
				},
			},
			expectedErr: CoordinatesIsRequiredErr,
		},
		{
			name: "not valid geometry type",
			featureCol: FeatureCollectionJSON{
				Type: "FeatureCollection",
				Features: []FeatureJSON{
					{
						Type: "Feature",
						Geometry: FeatureGeometryJSON{
							Type:        "NotPolygon",
							Coordinates: &json.RawMessage{},
						},
					},
				},
			},
			expectedErr: UnsupportedGeometryTypeErr{"NotPolygon"},
		},
		{
			name: "empty polygon coordinates",
			featureCol: FeatureCollectionJSON{
				Type: "FeatureCollection",
				Features: []FeatureJSON{
					{
						Type: "Feature",
						Geometry: FeatureGeometryJSON{
							Type:        "Polygon",
							Coordinates: &json.RawMessage{},
						},
					},
				},
			},
			expectedErr: NotValidPolygonCoordinatesErr,
		},
		{
			name: "not valid polygon coordinates",
			featureCol: FeatureCollectionJSON{
				Type: "FeatureCollection",
				Features: []FeatureJSON{
					{
						Type: "Feature",
						Geometry: FeatureGeometryJSON{
							Type:        "Polygon",
							Coordinates: &wrongCoords,
						},
					},
				},
			},
			expectedErr: NotValidPolygonCoordinatesErr,
		},
		{
			name: "empty multipolygon coordinates",
			featureCol: FeatureCollectionJSON{
				Type: "FeatureCollection",
				Features: []FeatureJSON{
					{
						Type: "Feature",
						Geometry: FeatureGeometryJSON{
							Type:        "MultiPolygon",
							Coordinates: &json.RawMessage{},
						},
					},
				},
			},
			expectedErr: NotValidMultiPolygonCoordinatesErr,
		},
		{
			name: "not valid multipolygon coordinates",
			featureCol: FeatureCollectionJSON{
				Type: "FeatureCollection",
				Features: []FeatureJSON{
					{
						Type: "Feature",
						Geometry: FeatureGeometryJSON{
							Type:        "MultiPolygon",
							Coordinates: &wrongCoords,
						},
					},
				},
			},
			expectedErr: NotValidMultiPolygonCoordinatesErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fc FeatureCollection
			err := fc.FromFeatureCollectionJSON(tt.featureCol)

			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}
