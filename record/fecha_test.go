package record

import (
	"encoding/xml"
	"testing"
	"time"
)

func TestFechaMarshalXML(t *testing.T) {
	type wrapper struct {
		Fecha Fecha
	}

	testCases := []struct {
		input Fecha
		want  string
	}{
		{Fecha(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)), "<wrapper><Fecha>02-01-2023</Fecha></wrapper>"},
		{Fecha(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)), "<wrapper><Fecha>25-12-2023</Fecha></wrapper>"},
		{Fecha(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)), "<wrapper><Fecha>15-06-2023</Fecha></wrapper>"},
	}

	for _, tt := range testCases {
		t.Run(tt.want, func(t *testing.T) {
			input := wrapper{Fecha: tt.input}
			xmlEnc, err := xml.Marshal(input)
			if err != nil {
				t.Fatalf("Fecha.MarshalXML returned error: %v", err)
			}

			if string(xmlEnc) != tt.want {
				t.Errorf("Fecha.MarshalXML() = %q, want %q", string(xmlEnc), tt.want)
			}
		})
	}

}

func TestFechaHoraMarshalXML(t *testing.T) {
	type wrapper struct {
		FechaHora FechaHora
	}

	testCases := []struct {
		input FechaHora
		want  string
	}{
		{FechaHora(time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC)), "<wrapper><FechaHora>2023-01-02T15:04:05Z</FechaHora></wrapper>"},
		{FechaHora(time.Date(2023, 12, 25, 23, 59, 59, 0, time.UTC)), "<wrapper><FechaHora>2023-12-25T23:59:59Z</FechaHora></wrapper>"},
		{FechaHora(time.Date(2023, 6, 15, 8, 30, 0, 0, time.UTC)), "<wrapper><FechaHora>2023-06-15T08:30:00Z</FechaHora></wrapper>"},
		{FechaHora(time.Date(2023, 6, 15, 8, 30, 0, 123456789, time.UTC)), "<wrapper><FechaHora>2023-06-15T08:30:00Z</FechaHora></wrapper>"},
		{FechaHora(time.Date(2023, 6, 15, 8, 30, 0, 999999999, time.FixedZone("CET", 3600))), "<wrapper><FechaHora>2023-06-15T08:30:00+01:00</FechaHora></wrapper>"},
	}

	for _, tt := range testCases {
		t.Run(tt.want, func(t *testing.T) {
			input := wrapper{FechaHora: tt.input}
			xmlEnc, err := xml.Marshal(input)
			if err != nil {
				t.Fatalf("FechaHora.MarshalXML returned error: %v", err)
			}

			if string(xmlEnc) != tt.want {
				t.Errorf("FechaHora.MarshalXML() = %q, want %q", string(xmlEnc), tt.want)
			}
		})

	}
}
