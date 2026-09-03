package aeat

import (
	"fmt"
)

const (
	rutaVerifactu = "/wlpl/TIKE-CONT/ws/SistemaFacturacion/VerifactuSOAP"
)

type Entorno string

const (
	EntornoProduccion Entorno = "produccion"
	EntornoPruebas    Entorno = "pruebas"
)

type TipoCertificado string

const (
	CertificadoRepresentante TipoCertificado = "representante"
	CertificadoSello         TipoCertificado = "sello"
)

var hosts = map[Entorno]map[TipoCertificado]string{
	EntornoProduccion: {
		CertificadoRepresentante: "www1.agenciatributaria.gob.es",
		CertificadoSello:         "www10.agenciatributaria.gob.es",
	},
	EntornoPruebas: {
		CertificadoRepresentante: "prewww1.aeat.es",
		CertificadoSello:         "prewww10.aeat.es",
	},
}

func endpoint(e Entorno, t TipoCertificado) (string, error) {
	host, ok := hosts[e][t]
	if !ok {
		return "", fmt.Errorf("%w: %s/%s", ErrEntornoDesconocido, e, t)
	}
	return "https://" + host + rutaVerifactu, nil
}
