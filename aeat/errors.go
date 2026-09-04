package aeat

import "errors"

var (
	ErrEntornoDesconocido  = errors.New("aeat: unknown environment")
	ErrRespuestaInesperada = errors.New("aeat: unexpected response from AEAT")
)
