package record

// Entorno represents the environment for QR code validation, the value is the URL of the validation endpoint.
type Entorno string

const (
	EntornoProduccion Entorno = "https://www2.agenciatributaria.gob.es/wlpl/TIKE-CONT/ValidarQR"
	EntornoPruebas    Entorno = "https://prewww2.aeat.es/wlpl/TIKE-CONT/ValidarQR"
)

func (r RegistroAlta) ComparisonURL(entorno Entorno) (string, error) {
	return "", nil
}
