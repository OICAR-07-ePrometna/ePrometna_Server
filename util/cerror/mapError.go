package cerror

import (
	"ePrometna_Server/util/format"
	"errors"
	"fmt"
)

var (
	ErrBadDateFormat = fmt.Errorf("bad date format, should be %s", format.DateFormat)
	ErrBadUuid       = errors.New("failed to parse uuid")
)
