package aeat

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func generarPar(t *testing.T, dir, nombre string) (rutaCert, rutaClave string) {
	t.Helper()

	clave, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		t.Fatalf("Error al generar la clave: %v", err)
	}

	plantilla := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "nombre"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(-time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	der, err := x509.CreateCertificate(rand.Reader, &plantilla, &plantilla, &clave.PublicKey, clave)
	if err != nil {
		t.Fatalf("Error al crear el certificado: %v", err)
	}

	rutaCert = filepath.Join(dir, nombre+".crt.pem")
	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	if err := os.WriteFile(rutaCert, certPem, 0o600); err != nil {
		t.Fatalf("Error al escribir el certificado: %v", err)
	}

	claveDER, err := x509.MarshalECPrivateKey(clave)
	if err != nil {
		t.Fatalf("Error al serializar la clave: %v", err)
	}

	rutaClave = filepath.Join(dir, nombre+".key.pem")

	if err := os.WriteFile(rutaClave, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: claveDER}), 0o600); err != nil {
		t.Fatalf("Error al escribir la clave: %v", err)
	}

	return rutaCert, rutaClave

}
