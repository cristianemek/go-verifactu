package aeat

import (
	"crypto/tls"
	"fmt"
)

// CargarPEM loads a client certificate and its private key from two PEM files.
// If a single file holds both blocks, pass the same path twice.
//
// Spanish qualified certificates are issued as .p12. Convert once with:
//
//	openssl pkcs12 -in certificado.p12 -out certificado.pem -nodes
//
// -nodes leaves the private key unencrypted on disk: restrict its permissions.
func CargarPEM(rutaCert, rutaClave string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(rutaCert, rutaClave)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("aeat: cannot load PEM certificate %q with key %q: %w", rutaCert, rutaClave, err)
	}

	return cert, nil

}
