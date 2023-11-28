package psql

import (
	"errors"

	"github.com/jackc/pgconn"
)

const internalPostgresErrorCode = "XX000"

type PostgisValidationErr struct {
	Message string
}

func (e PostgisValidationErr) Error() string {
	return e.Message
}

func parsePostgisError(err error) error {
	var e *pgconn.PgError
	ok := errors.As(err, &e)
	if !ok || e.Code != internalPostgresErrorCode {
		return err
	}
	// ToDo find a way to define postgis validation errors
	return PostgisValidationErr{Message: e.Message}
}
