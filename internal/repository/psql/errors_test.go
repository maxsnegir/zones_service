package psql

import (
	"testing"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/require"
)

type someErr string

func (e someErr) Error() string {
	return string(e)
}

func TestParsePostgisError(t *testing.T) {

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "any error",
			err:  someErr("some error"),
			want: someErr("some error"),
		},
		{
			name: "postgres any error",
			err:  &pgconn.PgError{Code: "23505", Message: "unique_violation"},
			want: &pgconn.PgError{Code: "23505", Message: "unique_violation"},
		},
		{
			name: "postgres internal error",
			err:  &pgconn.PgError{Code: internalPostgresErrorCode, Message: "self-intersection"},
			want: PostgisValidationErr{Message: "self-intersection"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parsePostgisError(tt.err)
			require.Equal(t, err, tt.want)
		})
	}
}
