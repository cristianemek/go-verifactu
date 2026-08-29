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
	TiempoEspera          time.Duration
	Lineas                []LineaEnvio
}

type LineaEnvio struct {
	IDFactura   IDFactura
	Operacion   Operacion
	Estado      record.EstadoRegistro
	CodigoError string
	Descripcion string
	Duplicado   *record.RegistroDuplicado
}

// Liquidada returns true if processed. AEAT keeps this record, so to fix it
// you must send a "Subsanación" patch. Warning: it reuses the same invoice ID.
func (e *LineaEnvio) Liquidada() bool {
	return e.Estado == record.EstadoRegistroCorrecto || e.Estado == record.EstadoRegistroAceptadoConErrores
}
