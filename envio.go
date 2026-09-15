package verifactu

import (
	"time"

	"github.com/cristianemek/go-verifactu/record"
)

type Envio struct {
	Instante              time.Time
	CSV                   string
	NIFPresentador        string
	TimestampPresentacion time.Time
	EstadoEnvio           record.EstadoEnvio
	Lineas                []LineaEnvio
	TiempoEspera          time.Duration
}

type LineaEnvio struct {
	IDFactura   IDFactura
	Operacion   Operacion
	Estado      record.EstadoRegistro
	CodigoError string
	Descripcion string
	Duplicado   *record.RegistroDuplicado `json:",omitempty"`
	Secuencia   uint64
}

// Procesada reports whether the AEAT gave a final answer.
// A rejection counts: fix it with TrasRechazo, never by resending the same record.
func (e LineaEnvio) Procesada() bool {
	return e.Estado == record.EstadoRegistroCorrecto || e.Estado == record.EstadoRegistroAceptadoConErrores || e.Estado == record.EstadoRegistroIncorrecto
}
