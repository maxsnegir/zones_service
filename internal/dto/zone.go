package dto

import (
	"errors"
)

type ZoneIds []int

func (ids ZoneIds) Validate() error {
	for _, zoneId := range ids {
		if zoneId < 1 {
			return ErrInvalidId
		}
	}
	return nil
}

var (
	EmptyIdsErr     = errors.New("ids cannot be empty")
	ErrInvalidId    = errors.New("invalid id")
	ErrDuplicateKey = errors.New("duplicate key")
	ErrEmptyData    = errors.New("empty data")
)

type ZoneContainsPointIn struct {
	ZoneIds ZoneIds `json:"ids"`
	Point   Point   `json:"point"`
}

type BatchZoneContainsPointIn struct {
	Key     string  `json:"key"`
	ZoneIds ZoneIds `json:"ids"`
	Point   Point   `json:"point"`
}

type BatchZoneContainsPointInCollection []BatchZoneContainsPointIn

func (b BatchZoneContainsPointInCollection) Validate() error {
	if len(b) == 0 {
		return ErrEmptyData
	}
	cacheKeys := make(map[string]struct{})
	for _, v := range b {
		if _, ok := cacheKeys[v.Key]; ok {
			return ErrDuplicateKey
		} else {
			cacheKeys[v.Key] = struct{}{}
		}
		if err := v.ZoneIds.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type ZoneContainsPointOut struct {
	ZoneId   int  `json:"id"`
	Contains bool `json:"contains"`
}

type BatchZoneContainsPointOut struct {
	Key      string `json:"key"`
	Contains bool   `json:"contains"`
}

func (in ZoneContainsPointIn) Validate() error {
	if err := in.ZoneIds.Validate(); err != nil {
		return err
	}

	return in.Point.Validate()
}
