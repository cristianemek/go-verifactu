package record

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

type RegistrationRecord struct {
	IDFactura    IDFacturaExpedida
	TipoFactura  string
	CuotaTotal   Amount
	ImporteTotal Amount
	// PreviousHash is the fingerprint of the previous record in the chain, empty
	// for the first record of a SIF. Note the naming trap: in the fingerprint
	// input string this field is named "Huella", but RegistroAlta/Huella in the
	// AEAT XML schema is this record's own fingerprint, not the previous one.
	PreviousHash             string
	FechaHoraHusoGenRegistro time.Time
}

type CancellationRecord struct {
	IDFacturaAnulada IDFacturaExpedidaBaja
	// PreviousHash is the fingerprint of the previous record in the chain, empty
	// for the first record of a SIF. Note the naming trap: in the fingerprint
	// input string this field is named "Huella", but RegistroAnulacion/Huella in the
	// AEAT XML schema is this record's own fingerprint, not the previous one.
	PreviousHash             string
	FechaHoraHusoGenRegistro time.Time
}

func registrationFingerprintInput(r RegistrationRecord) string {
	var b strings.Builder

	writeField(&b, "IDEmisorFactura", r.IDFactura.IDEmisorFactura)
	writeField(&b, "NumSerieFactura", r.IDFactura.NumSerieFactura)
	writeField(&b, "FechaExpedicionFactura", r.IDFactura.FechaExpedicionFactura.Format())
	writeField(&b, "TipoFactura", r.TipoFactura)
	writeField(&b, "CuotaTotal", r.CuotaTotal.Format())
	writeField(&b, "ImporteTotal", r.ImporteTotal.Format())
	writeField(&b, "Huella", r.PreviousHash)
	writeField(&b, "FechaHoraHusoGenRegistro", r.FechaHoraHusoGenRegistro.Format(time.RFC3339))

	return b.String()
}

func hashFingerprintInput(s string) string {
	hash := sha256.Sum256([]byte(s))
	hashString := hex.EncodeToString(hash[:])

	return strings.ToUpper(hashString)
}

func cancellationFingerprintInput(c CancellationRecord) string {
	var b strings.Builder

	writeField(&b, "IDEmisorFacturaAnulada", c.IDFacturaAnulada.IDEmisorFacturaAnulada)
	writeField(&b, "NumSerieFacturaAnulada", c.IDFacturaAnulada.NumSerieFacturaAnulada)
	writeField(&b, "FechaExpedicionFacturaAnulada", c.IDFacturaAnulada.FechaExpedicionFacturaAnulada.Format())
	writeField(&b, "Huella", c.PreviousHash)
	writeField(&b, "FechaHoraHusoGenRegistro", c.FechaHoraHusoGenRegistro.Format(time.RFC3339))

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
func (r RegistrationRecord) Fingerprint() string {
	return hashFingerprintInput(registrationFingerprintInput(r))
}

// FingerprintInput returns the fingerprint input string for the RegistrationRecord.
func (r RegistrationRecord) FingerprintInput() string {
	return registrationFingerprintInput(r)
}

// Fingerprint returns 64-character uppercase hexadecimal SHA-256 hash of the fingerprint input string, RD 1007/2023.
func (c CancellationRecord) Fingerprint() string {
	return hashFingerprintInput(cancellationFingerprintInput(c))
}

// FingerprintInput returns the fingerprint input string for the CancellationRecord.
func (c CancellationRecord) FingerprintInput() string {
	return cancellationFingerprintInput(c)
}
