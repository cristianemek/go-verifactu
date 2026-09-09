package record

import (
	"errors"
)

var (
	ErrInvalidAmount     = errors.New("record: invalid amount")
	ErrInvalidPorcentaje = errors.New("record: invalid porcentaje")
	ErrValidation        = errors.New("record: validation error")
	ErrAviso             = errors.New("record: warning, the AEAT accepts the record but flags it")
	ErrInvalidQr         = errors.New("record: invalid QR data")
)
