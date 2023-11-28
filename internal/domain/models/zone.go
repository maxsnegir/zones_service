package models

import (
	"encoding/json"

	"github.com/twpayne/go-geom"
)

type Zone struct {
	Id      int64         `json:"id"`
	Title   string        `json:"title"`
	Polygon *geom.Polygon `json:"polygon"`
}

func (z *Zone) MarshalJSON() ([]byte, error) {
	type Alias Zone
	alias := struct {
		*Alias
		Id      string         `json:"id"`
		Polygon [][]geom.Coord `json:"polygon"`
	}{
		Alias:   (*Alias)(z),
		Polygon: z.Polygon.Coords(),
	}
	return json.Marshal(alias)
}
