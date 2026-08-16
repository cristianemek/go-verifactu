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
