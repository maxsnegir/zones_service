package dto

import (
	"errors"
)

var (
	EmptyIdsErr  = errors.New("ids cannot be empty")
	ErrInvalidId = errors.New("invalid id")
)

type ZoneContainsPointIn struct {
	ZoneIds []int `json:"ids"`
	Point   Point `json:"point"`
}

type ZoneContainsPointOut struct {
	ZoneId   int  `json:"id"`
	Contains bool `json:"contains"`
}

func (in ZoneContainsPointIn) Validate() error {
	if len(in.ZoneIds) == 0 {
		return EmptyIdsErr
	}
	for _, zoneId := range in.ZoneIds {
		if zoneId < 1 {
			return ErrInvalidId
		}
	}

	return in.Point.Validate()
}
