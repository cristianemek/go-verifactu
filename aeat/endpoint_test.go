package aeat

import (
	"errors"
	"testing"
)

func TestEndpoint(t *testing.T) {
	testCases := []struct {
		name    string
		entorno Entorno
		tipo    TipoCertificado
		want    string
		wantErr error
	}{
		{
			name:    "pruebas representante",
			entorno: EntornoPruebas,
			tipo:    CertificadoRepresentante,
			want:    "https://prewww1.aeat.es/wlpl/TIKE-CONT/ws/SistemaFacturacion/VerifactuSOAP",
			wantErr: nil,
		},
		{
			name:    "pruebas sello",
			entorno: EntornoPruebas,
			tipo:    CertificadoSello,
			want:    "https://prewww10.aeat.es/wlpl/TIKE-CONT/ws/SistemaFacturacion/VerifactuSOAP",
			wantErr: nil,
		},
		{
			name:    "produccion representante",
			entorno: EntornoProduccion,
			tipo:    CertificadoRepresentante,
			want:    "https://www1.agenciatributaria.gob.es/wlpl/TIKE-CONT/ws/SistemaFacturacion/VerifactuSOAP",
			wantErr: nil,
		},
		{
			name:    "produccion sello",
			entorno: EntornoProduccion,
			tipo:    CertificadoSello,
			want:    "https://www10.agenciatributaria.gob.es/wlpl/TIKE-CONT/ws/SistemaFacturacion/VerifactuSOAP",
			wantErr: nil,
		},
		{
			name:    "entorno desconocido",
			entorno: Entorno("desconocido"),
			tipo:    CertificadoRepresentante,
			want:    "",
			wantErr: ErrEntornoDesconocido,
		},
		{
			name:    "tipo de certificado desconocido",
			entorno: EntornoPruebas,
			tipo:    TipoCertificado("desconocido"),
			want:    "",
			wantErr: ErrEntornoDesconocido,
		},
		{
			name:    "ambos desconocidos",
			entorno: Entorno("desconocido"),
			tipo:    TipoCertificado("desconocido"),
			want:    "",
			wantErr: ErrEntornoDesconocido,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := endpoint(tc.entorno, tc.tipo)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("endpoint() error = %v, wantErr %v", err, tc.wantErr)
				}

				if got != "" {
					t.Errorf("endpoint() got = %v, want empty string on error", got)
				}
			} else {

				if err != nil {
					t.Errorf("endpoint() unexpected error = %v", err)
				}

				if got != tc.want {
					t.Errorf("endpoint() got = %v, want %v", got, tc.want)
				}
			}

		})
	}
}
