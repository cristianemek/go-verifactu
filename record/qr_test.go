package record

import (
	"errors"
	"testing"
)

func TestComparisonURL(t *testing.T) {
	rec := validRegistroAlta()
	rec.IDFactura.NumSerieFactura = "12345678&G33"
	rec.ImporteTotal = 24140

	got, err := rec.ComparisonURL(EntornoPruebas)
	if err != nil {
		t.Fatalf("Error occurred: %v", err)
	}

	wantedURL := "https://prewww2.aeat.es/wlpl/TIKE-CONT/ValidarQR?fecha=01-01-2024&importe=241.40&nif=89890001K&numserie=12345678%26G33"

	if got != wantedURL {
		t.Errorf("Unexpected comparison URL: %s", got)
	}
}

func TestProductionComparisonURL(t *testing.T) {
	rec := validRegistroAlta()
	rec.IDFactura.NumSerieFactura = "12345678&G33"
	rec.ImporteTotal = 24140

	got, err := rec.ComparisonURL(EntornoProduccion)
	if err != nil {
		t.Fatalf("Error occurred: %v", err)
	}
	wantedURL := "https://www2.agenciatributaria.gob.es/wlpl/TIKE-CONT/ValidarQR?fecha=01-01-2024&importe=241.40&nif=89890001K&numserie=12345678%26G33"

	if got != wantedURL {
		t.Errorf("Unexpected comparison URL: %s", got)
	}
}

func TestComparisonURLWithInvalidCharacters(t *testing.T) {
	testCases := []struct {
		name         string
		modifyRecord func(*RegistroAlta)
	}{
		{
			name: "IDEmisorFactura with non-ASCII characters",
			modifyRecord: func(r *RegistroAlta) {
				r.IDFactura.IDEmisorFactura = "89890001K€"
			},
		},
		{
			name: "Invalid Ñ character",
			modifyRecord: func(r *RegistroAlta) {
				r.IDFactura.IDEmisorFactura = "12345678/Ñ33"
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := validRegistroAlta()
			tc.modifyRecord(&rec)
			_, err := rec.ComparisonURL(EntornoPruebas)

			if err == nil {
				t.Errorf("Expected error for %s, but got none", tc.name)
			}

			if !errors.Is(err, ErrInvalidQr) {
				t.Errorf("expected ErrInvalidQr")
			}

		})

	}

}
