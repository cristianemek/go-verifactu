package aeat

import "errors"

var (
	ErrEntornoDesconocido   = errors.New("aeat: unknown environment")
	ErrRespuestaInesperada  = errors.New("aeat: unexpected response from AEAT")
	ErrCertificadoRequerido = errors.New("aeat: a client certificate is required to talk to the AEAT")
	ErrAccesoDenegado       = errors.New("aeat: the request was redirected to an error page: the client certificate is missing or was not accepted")
)
