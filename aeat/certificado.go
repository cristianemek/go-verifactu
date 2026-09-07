package aeat

import (
	"crypto/tls"
	"fmt"
)

func CargarPEM(rutaCert, rutaClave string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(rutaCert, rutaClave)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("aeat: cannot load PEM certificate %q with key %q: %w", rutaCert, rutaClave, err)
	}

	return cert, nil

}
