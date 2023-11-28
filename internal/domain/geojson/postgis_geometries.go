package geojson

import (
	"database/sql"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

type PostgisGeometry interface {
	ToEwkb() sql.Scanner
}

type PostgisMultiPolygon struct {
	geom.MultiPolygon
}

type PostgisPolygon struct {
	geom.Polygon
}

func (p *PostgisPolygon) ToEwkb() sql.Scanner {
	return &ewkb.Polygon{Polygon: &p.Polygon}
}

func (mp *PostgisMultiPolygon) ToEwkb() sql.Scanner {
	return &ewkb.MultiPolygon{MultiPolygon: &mp.MultiPolygon}
}
