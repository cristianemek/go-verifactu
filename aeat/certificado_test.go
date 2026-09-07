package aeat

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io/fs"
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
		Subject:      pkix.Name{CommonName: nombre},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
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

func TestCargarPEM(t *testing.T) {
	dir := t.TempDir()

	rutaCert, rutaClave := generarPar(t, dir, "cliente")

	cert, err := CargarPEM(rutaCert, rutaClave)
	if err != nil {
		t.Fatalf("Error al cargar el certificado y la clave: %v", err)
	}

	if len(cert.Certificate) != 1 {
		t.Fatalf("Se esperaba un certificado, pero se obtuvieron %d", len(cert.Certificate))
	}

}

func TestCargarPEMNoExiste(t *testing.T) {
	dir := t.TempDir()

	rutaCert := filepath.Join(dir, "no_existe.crt.pem")
	rutaClave := filepath.Join(dir, "no_existe.key.pem")

	_, err := CargarPEM(rutaCert, rutaClave)
	if err == nil {
		t.Fatalf("Se esperaba un error al cargar un certificado que no existe")
	}

	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Se esperaba un error de tipo fs.ErrNotExist, pero se obtuvo: %v", err)
	}
}

func TestCargarPEMClaveQueNoCorresponde(t *testing.T) {
	dir := t.TempDir()

	rutaCert, _ := generarPar(t, dir, "uno")
	_, rutaClave := generarPar(t, dir, "dos")

	_, err := CargarPEM(rutaCert, rutaClave)

	if err == nil {
		t.Errorf("Se esperaba un error al cargar un certificado y una clave que no corresponden")
	}

}
