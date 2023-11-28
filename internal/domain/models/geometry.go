package models

import (
	"github.com/twpayne/go-geom"
)

type Geometry struct {
	Id         int64                  `json:"id"`
	Geom       geom.T                 `json:"geom"`
	Properties map[string]interface{} `json:"properties"`
}
