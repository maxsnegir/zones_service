package http

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseZoneIds(t *testing.T) {
	tests := []struct {
		name        string
		zoneIdsStr  string
		wantErr     bool
		expectedErr error
		expectedIds []int
	}{
		{
			name:        "empty ids",
			zoneIdsStr:  "",
			wantErr:     true,
			expectedErr: ErrEmptyZoneIds,
		},
		{
			name:        "wrong ids",
			zoneIdsStr:  "1,2,a,x,4",
			wantErr:     true,
			expectedErr: ErrInvalidZoneId,
		},
		{
			name:        "valid ids",
			zoneIdsStr:  "1,2,3,4",
			wantErr:     false,
			expectedIds: []int{1, 2, 3, 4},
		},
		{
			name:        "negative ids",
			zoneIdsStr:  "-1,2,3,4",
			wantErr:     true,
			expectedErr: ErrInvalidZoneId,
		},
		{
			name:        "zero ids",
			zoneIdsStr:  "0,2,3,4",
			wantErr:     true,
			expectedErr: ErrInvalidZoneId,
		},
		{
			name:        "duplicate ids",
			zoneIdsStr:  "4,4,3,3,1,1,1,2,2,10",
			wantErr:     false,
			expectedIds: []int{1, 2, 3, 4, 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIds, err := parseZoneIds(tt.zoneIdsStr, true)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tt.expectedIds), len(gotIds))
				require.ElementsMatch(t, tt.expectedIds, gotIds)
			}
		})
	}
}
