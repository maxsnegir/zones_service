package http

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrInvalidZoneId = errors.New("invalid zone id")
	ErrEmptyZoneIds  = errors.New("ids is required")
)

func parseZoneIds(ids string, isRequired bool) ([]int, error) {
	zoneIdsStr := strings.Split(ids, ",")
	zoneIds := make([]int, 0, len(zoneIdsStr))

	if len(zoneIdsStr) == 1 && zoneIdsStr[0] == "" && isRequired {
		return nil, ErrEmptyZoneIds

	}

	for _, zoneIdStr := range zoneIdsStr {
		zoneId, err := strconv.Atoi(zoneIdStr)
		if err != nil {
			return nil, ErrInvalidZoneId
		}
		zoneIds = append(zoneIds, zoneId)
	}
	return zoneIds, nil
}
