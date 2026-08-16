package record

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func registrationFingerprintInput(r RegistroAlta) string {
	var b strings.Builder

	writeField(&b, "IDEmisorFactura", r.IDFactura.IDEmisorFactura)
	writeField(&b, "NumSerieFactura", r.IDFactura.NumSerieFactura)
	writeField(&b, "FechaExpedicionFactura", r.IDFactura.FechaExpedicionFactura.Format())
	writeField(&b, "TipoFactura", string(r.TipoFactura))
	writeField(&b, "CuotaTotal", r.CuotaTotal.Format())
	writeField(&b, "ImporteTotal", r.ImporteTotal.Format())
	writeField(&b, "Huella", previousFingerprint(r.Encadenamiento))
	writeField(&b, "FechaHoraHusoGenRegistro", r.FechaHoraHusoGenRegistro.Format())

	return b.String()
}

func hashFingerprintInput(s string) string {
	hash := sha256.Sum256([]byte(s))
	hashString := hex.EncodeToString(hash[:])

	return strings.ToUpper(hashString)
}

func cancellationFingerprintInput(c RegistroAnulacion) string {
	var b strings.Builder

	writeField(&b, "IDEmisorFacturaAnulada", c.IDFactura.IDEmisorFacturaAnulada)
	writeField(&b, "NumSerieFacturaAnulada", c.IDFactura.NumSerieFacturaAnulada)
	writeField(&b, "FechaExpedicionFacturaAnulada", c.IDFactura.FechaExpedicionFacturaAnulada.Format())
	writeField(&b, "Huella", previousFingerprint(c.Encadenamiento))
	writeField(&b, "FechaHoraHusoGenRegistro", c.FechaHoraHusoGenRegistro.Format())

	return b.String()
}

func writeField(b *strings.Builder, name string, value string) {
	if b.Len() > 0 {
		b.WriteString("&")
	}
	b.WriteString(name)
	b.WriteString("=")
	b.WriteString(strings.TrimSpace(value))
}

// Fingerprint returns 64-character uppercase hexadecimal SHA-256 hash of the fingerprint input string, RD 1007/2023.
func (r RegistroAlta) Fingerprint() string {
	return hashFingerprintInput(registrationFingerprintInput(r))
}

// FingerprintInput returns the fingerprint input string for the RegistrationRecord.
func (r RegistroAlta) FingerprintInput() string {
	return registrationFingerprintInput(r)
}

// Fingerprint returns 64-character uppercase hexadecimal SHA-256 hash of the fingerprint input string, RD 1007/2023.
func (c RegistroAnulacion) Fingerprint() string {
	return hashFingerprintInput(cancellationFingerprintInput(c))
}

// FingerprintInput returns the fingerprint input string for the CancellationRecord.
func (c RegistroAnulacion) FingerprintInput() string {
	return cancellationFingerprintInput(c)
}

func previousFingerprint(e Encadenamiento) string {
	if e.RegistroAnterior == nil {
		return ""
	}

	return e.RegistroAnterior.Huella
}
