package geojson

type FeatureCollection struct {
	Type     string     `json:"type"`
	Features []*Feature `json:"features"`
}

type Feature struct {
	Type       string                 `json:"type"`
	Geometry   PostgisGeometry        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

func (fc *FeatureCollection) FromFeatureCollectionJSON(geojson FeatureCollectionJSON) error {
	if geojson.Type != "FeatureCollection" {
		return NotValidFeatureCollectionType{geojson.Type}
	}

	if geojson.Features == nil || len(geojson.Features) == 0 {
		return FeaturesIsRequiredErr
	}

	features := make([]*Feature, 0, len(geojson.Features))
	for _, feature := range geojson.Features {
		geometry, err := feature.Geometry.Decode()
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
