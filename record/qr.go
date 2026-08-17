package record

import (
	"fmt"
	"net/url"
)

// Entorno represents the environment for QR code validation, the value is the URL of the validation endpoint.
type Entorno string

const (
	EntornoProduccion Entorno = "https://www2.agenciatributaria.gob.es/wlpl/TIKE-CONT/ValidarQR"
	EntornoPruebas    Entorno = "https://prewww2.aeat.es/wlpl/TIKE-CONT/ValidarQR"
)

// ComparisonURL builds the AEAT verification URL that goes inside the invoice
// QR code. It returns just the URL: rendering the image is up to you.
//
// The printed QR must follow ISO/IEC 18004:2015, measure between 30x30 and
// 40x40 mm, use error correction level M, and keep at least 2 mm of blank space
// on all four sides (6 mm recommended). It goes at the top of the first page,
// with "QR tributario:" above it and "VERI*FACTU" below. See the official AEAT
// QR specification for layout details and edge cases.
//
// Query parameters come out alphabetically, unlike the AEAT examples. The
// service accepts either.
func (r RegistroAlta) ComparisonURL(entorno Entorno) (string, error) {

	if !validAscii(r.IDFactura.IDEmisorFactura) {
		return "", fmt.Errorf("%w: IDEmisorFactura contains non-ASCII characters", ErrInvalidQr)
	}
	if !validAscii(r.IDFactura.NumSerieFactura) {
		return "", fmt.Errorf("%w: NumSerieFactura contains non-ASCII characters", ErrInvalidQr)
	}

	if !validAscii(r.IDFactura.FechaExpedicionFactura.Format()) {
		return "", fmt.Errorf("%w: FechaExpedicionFactura contains non-ASCII characters", ErrInvalidQr)
	}

	if !validAscii(r.ImporteTotal.Format()) {
		return "", fmt.Errorf("%w: ImporteTotal contains non-ASCII characters", ErrInvalidQr)
	}

	params := url.Values{}

	params.Set("nif", r.IDFactura.IDEmisorFactura)
	params.Set("numserie", r.IDFactura.NumSerieFactura)
	params.Set("fecha", r.IDFactura.FechaExpedicionFactura.Format())
	params.Set("importe", r.ImporteTotal.Format())

	return string(entorno) + "?" + params.Encode(), nil
}

func validAscii(s string) bool {

	for _, r := range s {
		if r < 32 || r > 126 {
			return false
		}
	}

	return true
}
