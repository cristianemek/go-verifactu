package record

import (
	"encoding/json"
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

func TestFechaJson(t *testing.T) {
	fecha := Fecha(time.Date(2024, 1, 1, 0, 30, 0, 0, time.FixedZone("CET", 3600)))

	jsonData, err := json.Marshal(fecha)
	if err != nil {
		t.Fatalf("Fecha.MarshalJSON returned error: %v", err)
	}

	if string(jsonData) != `"01-01-2024"` {
		t.Errorf("Fecha.MarshalJSON() = %q, want %q", string(jsonData), `"01-01-2024"`)
	}

	fecha2 := Fecha{}
	validJSON := `"01-01-2024"`

	err = json.Unmarshal([]byte(validJSON), &fecha2)

	if err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}

	if fecha2.Format() != "01-01-2024" {
		t.Errorf("UnmarshalJSON() = %q, want %q", fecha2.Format(), "01-01-2024")
	}

	invalidJSON := `"not-a-date"`
	err = json.Unmarshal([]byte(invalidJSON), &fecha2)
	if err == nil {
		t.Fatalf("Expected error for invalid JSON, got nil")
	}

}

func TestFechaHoraJson(t *testing.T) {
	fechaHora := FechaHora(time.Date(2024, 1, 1, 0, 30, 0, 0, time.FixedZone("CET", 3600)))
	jsonData, err := json.Marshal(fechaHora)
	if err != nil {
		t.Fatalf("FechaHora.MarshalJSON returned error: %v", err)
	}

	if string(jsonData) != `"2024-01-01T00:30:00+01:00"` {
		t.Errorf("FechaHora.MarshalJSON() = %q, want %q", string(jsonData), `"2024-01-01T00:30:00+01:00"`)
	}

	fechaHora2 := FechaHora{}
	validJSON := `"2024-01-01T00:30:00+01:00"`

	err = json.Unmarshal([]byte(validJSON), &fechaHora2)

	if err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}

	if fechaHora2.Format() != "2024-01-01T00:30:00+01:00" {
		t.Errorf("UnmarshalJSON() = %q, want %q", fechaHora2.Format(), "2024-01-01T00:30:00+01:00")
	}

	invalidJSON := `"not-a-date"`
	err = json.Unmarshal([]byte(invalidJSON), &fechaHora2)
	if err == nil {
		t.Fatalf("Expected error for invalid JSON, got nil")
	}
}

func TestFechaXML(t *testing.T) {
	fecha := Fecha(time.Date(2024, 1, 1, 0, 30, 0, 0, time.FixedZone("CET", 3600)))

	xmlData, err := xml.Marshal(fecha)
	if err != nil {
		t.Fatalf("Fecha.MarshalXML returned error: %v", err)
	}

	if string(xmlData) != `<Fecha>01-01-2024</Fecha>` {
		t.Errorf("Fecha.MarshalXML() = %q, want %q", string(xmlData), `<Fecha>01-01-2024</Fecha>`)
	}

	fecha2 := Fecha{}
	validXML := `<Fecha>01-01-2024</Fecha>`

	err = xml.Unmarshal([]byte(validXML), &fecha2)

	if err != nil {
		t.Fatalf("UnmarshalXML returned error: %v", err)
	}

	if fecha2.Format() != "01-01-2024" {
		t.Errorf("UnmarshalXML() = %q, want %q", fecha2.Format(), "01-01-2024")
	}

	invalidXML := `<Fecha>not-a-date</Fecha>`
	err = xml.Unmarshal([]byte(invalidXML), &fecha2)
	if err == nil {
		t.Fatalf("Expected error for invalid XML, got nil")
	}

}

func TestFechaHoraXML(t *testing.T) {
	fechaHora := FechaHora(time.Date(2024, 1, 1, 0, 30, 0, 0, time.FixedZone("CET", 3600)))

	xmlData, err := xml.Marshal(fechaHora)
	if err != nil {
		t.Fatalf("FechaHora.MarshalXML returned error: %v", err)
	}

	if string(xmlData) != `<FechaHora>2024-01-01T00:30:00+01:00</FechaHora>` {
		t.Errorf("FechaHora.MarshalXML() = %q, want %q", string(xmlData), `<FechaHora>2024-01-01T00:30:00+01:00</FechaHora>`)
	}

	fechaHora2 := FechaHora{}
	validXML := `<FechaHora>2024-01-01T00:30:00+01:00</FechaHora>`

	err = xml.Unmarshal([]byte(validXML), &fechaHora2)

	if err != nil {
		t.Fatalf("UnmarshalXML returned error: %v", err)
	}

	if fechaHora2.Format() != "2024-01-01T00:30:00+01:00" {
		t.Errorf("UnmarshalXML() = %q, want %q", fechaHora2.Format(), "2024-01-01T00:30:00+01:00")
	}

	invalidXML := `<FechaHora>not-a-date</FechaHora>`
	err = xml.Unmarshal([]byte(invalidXML), &fechaHora2)
	if err == nil {
		t.Fatalf("Expected error for invalid XML, got nil")
	}

}
