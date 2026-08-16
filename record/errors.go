package record

import (
	"errors"
)

var (
	ErrInvalidAmount     = errors.New("record: invalid amount")
	ErrInvalidPorcentaje = errors.New("record: invalid porcentaje")
	ErrValidation        = errors.New("record: validation error")
)
