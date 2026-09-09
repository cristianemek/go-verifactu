package record

import "fmt"

// Regimenes donde el desglose es parcial por diseño, el total no tiene porque coincidir con la suma.
var regimenDesgloseParcial = map[ClaveRegimen]bool{
	ClaveRegimen03: true,
	ClaveRegimen05: true,
	ClaveRegimen06: true,
	ClaveRegimen08: true,
	ClaveRegimen09: true,
}

const (
	// ±10,00 €, del apartado 16 del PDF de validaciones de la aeat
	toleranciaTotales = Amount(1000)
)

func diferencia(a, b Amount) Amount {
	if a > b {
		return a - b
	}
	return b - a
}

// Avisos devuelve una lista de avisos si hay discrepancias en los totales del registro.
// Validate no incluye esto a proposito, porque la AEAT lo considera un aviso, no un error. Se puede enviar el registro aunque haya avisos.
func (r RegistroAlta) Avisos() []error {
	if desgloseParcial(r.Desglose) {
		return nil
	}

	bases, cuotas, recargos := sumasDesglose(r.Desglose)

	var avisos []error

	if diferencia(r.CuotaTotal, cuotas+recargos) > toleranciaTotales {
		avisos = append(avisos, fmt.Errorf("%w: cuota total es %v, cuotas + recargos es %v, diferencia: %v - error AEAT 2006", ErrAviso, r.CuotaTotal.Format(), (cuotas+recargos).Format(), diferencia(r.CuotaTotal, cuotas+recargos).Format()))
	}

	if diferencia(r.ImporteTotal, bases+cuotas+recargos) > toleranciaTotales {
		avisos = append(avisos, fmt.Errorf("%w: importe total es %v, bases + cuotas + recargos es %v, diferencia: %v - error AEAT 2005", ErrAviso, r.ImporteTotal.Format(), (bases+cuotas+recargos).Format(), diferencia(r.ImporteTotal, bases+cuotas+recargos).Format()))
	}

	return avisos

}

func desgloseParcial(d Desglose) bool {
	for _, det := range d.DetalleDesglose {

		if det.ClaveRegimen != nil && regimenDesgloseParcial[*det.ClaveRegimen] {
			return true
		}

	}
	return false
}

func sumasDesglose(d Desglose) (bases, cuotas, recargos Amount) {
	for _, det := range d.DetalleDesglose {
		bases += det.BaseImponibleOimporteNoSujeto

		if det.CuotaRepercutida != nil {
			cuotas += *det.CuotaRepercutida
		}

		if det.CuotaRecargoEquivalencia != nil {
			recargos += *det.CuotaRecargoEquivalencia
		}
	}
	return
}
