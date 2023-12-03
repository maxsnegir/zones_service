package dto

import (
	"errors"
)

var (
	InvalidLatitudeError  = errors.New("invalid latitude")
	InvalidLongitudeError = errors.New("invalid longitude")
)

type Point struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

func (p *Point) Validate() error {
	if p.Lat < -90 || p.Lat > 90 {
		return InvalidLatitudeError
	}
	if p.Lon < -180 || p.Lon > 180 {
		return InvalidLongitudeError
	}
	return nil
}
