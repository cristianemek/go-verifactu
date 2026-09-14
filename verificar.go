package verifactu

import (
	"fmt"

	"github.com/cristianemek/go-verifactu/record"
)

func VerificarCadena(entradas []*Entry) error {
	for i, e := range entradas {
		if e.Secuencia != uint64(i+1) {
			return fmt.Errorf("%w: secuencia %d in position %d, expected %d", ErrCadenaBifurcada, e.Secuencia, i, i+1)
		}

		var calculada string
		var enc record.Encadenamiento

		switch e.Operacion {
		case OperacionAlta:
			if e.Alta == nil {
				return fmt.Errorf("%w: corrupt entry: missing Alta for secuencia %d", ErrCadenaBifurcada, e.Secuencia)
			}

			calculada = e.Alta.Fingerprint()
			enc = e.Alta.Encadenamiento

		case OperacionAnulacion:
			if e.Anulacion == nil {
				return fmt.Errorf("%w: corrupt entry: missing Anulacion for secuencia %d", ErrCadenaBifurcada, e.Secuencia)
			}

			calculada = e.Anulacion.Fingerprint()
			enc = e.Anulacion.Encadenamiento
		default:
			return fmt.Errorf("%w: unknown operation type %s for secuencia %d", ErrCadenaBifurcada, e.Operacion, e.Secuencia)
		}

		if calculada != e.Huella {
			return fmt.Errorf("%w: corrupt entry: mismatched fingerprint for secuencia %d", ErrCadenaBifurcada, e.Secuencia)
		}

		if i == 0 {
			if enc.PrimerRegistro == nil {
				return fmt.Errorf("%w: the first entry (secuencia %d) must have a non-nil PrimerRegistro", ErrCadenaBifurcada, e.Secuencia)
			}

			if enc.RegistroAnterior != nil {
				return fmt.Errorf("%w: the first entry (secuencia %d) must have a nil RegistroAnterior", ErrCadenaBifurcada, e.Secuencia)
			}
		} else {
			if enc.RegistroAnterior == nil {
				return fmt.Errorf("%w: entry at secuencia %d must have a non-nil RegistroAnterior", ErrCadenaBifurcada, e.Secuencia)
			}

			if enc.RegistroAnterior.Huella != entradas[i-1].Huella {
				return fmt.Errorf("%w: entry at secuencia %d has mismatched RegistroAnterior Huella", ErrCadenaBifurcada, e.Secuencia)
			}
		}
	}
	return nil
}
