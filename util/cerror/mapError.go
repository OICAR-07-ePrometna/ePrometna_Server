package cerror

import (
	"ePrometna_Server/util/format"
	"errors"
	"fmt"
)

var (
	ErrBadDateFormat      = fmt.Errorf("bad date format, should be %s", format.DateFormat)
	ErrBadDateTimeFormat  = fmt.Errorf("bad date and time format, should be %s", format.DateTimeFormat)
	ErrBadTimeFormat      = fmt.Errorf("bad time format, should be %s", format.TimeFormat)
	ErrBadUuid            = errors.New("failed to parse uuid")
	ErrUnknownRole        = errors.New("unknown role")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidTokenFormat = errors.New("invalid token format")
)
