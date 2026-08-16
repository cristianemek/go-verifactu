package record

import (
	"encoding/xml"
	"errors"
	"testing"
)

func TestParseAmount(t *testing.T) {
	tests := []struct {
		input string
		want  Amount
	}{
		{"12.10", 1210},
		{"12", 1200},
		{"1234.5", 123450},
		{"-0.50", -50},
		{"   21.99   ", 2199},
		{".07", 7},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseAmount(tt.input)
			if err != nil {
				t.Fatalf("ParseAmount %q returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseAmount %q, expected %d and returned: %d", tt.input, tt.want, got)
			}
		})

	}
}

func TestParseAmountInvalid(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"abc"},
		{"12.10.5"},
		{"12,15"},
		{""},
		{"1.234"},
		{"12.+5"},
		{"12.-5"},
		{"-"},
		{"+-"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseAmount(tt.input)
			if err == nil {
				t.Fatalf("ParseAmount %q, expected error and returned nil", tt.input)
			}
			if !errors.Is(err, ErrInvalidAmount) {
				t.Errorf("ParseAmount %q, expected ErrInvalidAmount and returned: %v", tt.input, err)
			}
		})
	}

}

func TestAmountFormat(t *testing.T) {
	tests := []struct {
		input Amount
		want  string
	}{
		{1210, "12.10"},
		{1200, "12.00"},
		{123450, "1234.50"},
		{-50, "-0.50"},
		{2199, "21.99"},
		{7, "0.07"},
		{0, "0.00"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.input.Format()
			if got != tt.want {
				t.Errorf("Amount.Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAmountMarshalXML(t *testing.T) {
	type wrapper struct {
		Amount Amount
	}

	testCases := []struct {
		input Amount
		want  string
	}{
		{Amount(1234), "<wrapper><Amount>12.34</Amount></wrapper>"},
		{Amount(-5678), "<wrapper><Amount>-56.78</Amount></wrapper>"},
		{Amount(0), "<wrapper><Amount>0.00</Amount></wrapper>"},
		{Amount(50), "<wrapper><Amount>0.50</Amount></wrapper>"},
	}

	for _, tt := range testCases {
		t.Run(tt.want, func(t *testing.T) {
			input := wrapper{Amount: tt.input}
			xmlEnc, err := xml.Marshal(input)
			if err != nil {
				t.Fatalf("Amount.MarshalXML returned error: %v", err)
			}

			if string(xmlEnc) != tt.want {
				t.Errorf("Amount.MarshalXML() = %q, want %q", string(xmlEnc), tt.want)
			}
		})
	}

}

func TestPorcentajeFormat(t *testing.T) {
	tests := []struct {
		input Porcentaje
		want  string
	}{
		{1210, "12.10"},
		{1200, "12.00"},
		{2199, "21.99"},
		{7, "0.07"},
		{0, "0.00"},
		{99999, "999.99"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.input.Format()
			if got != tt.want {
				t.Errorf("Porcentaje.Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParsePorcentaje(t *testing.T) {
	tests := []struct {
		input string
		want  Porcentaje
	}{
		{"12", 1200},
		{"12.00", 1200},
		{"21.99", 2199},
		{"5.0", 500},
		{"0", 0},
		{"   21   ", 2100},
		{"999.99", 99999},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParsePorcentaje(tt.input)
			if err != nil {
				t.Fatalf("ParsePorcentaje %q returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParsePorcentaje %q, expected %d and returned: %d", tt.input, tt.want, got)
			}
		})
	}
}

func TestParsePorcentajeInvalid(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"abc"},
		{"12.10.5"},
		{"12,15"},
		{""},
		{"1000"},
		{"+34"},
		{"-12"},
		{"12.-5"},
		{"-"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParsePorcentaje(tt.input)
			if err == nil {
				t.Fatalf("ParsePorcentaje %q, expected error and returned nil", tt.input)
			}
			if !errors.Is(err, ErrInvalidPorcentaje) {
				t.Errorf("ParsePorcentaje %q, expected ErrInvalidPorcentaje and returned: %v", tt.input, err)
			}
		})
	}
}

func TestPorcentajeMarshalXML(t *testing.T) {
	type wrapper struct {
		Porcentaje Porcentaje
	}

	testCases := []struct {
		input Porcentaje
		want  string
	}{
		{Porcentaje(1234), "<wrapper><Porcentaje>12.34</Porcentaje></wrapper>"},
		{Porcentaje(0), "<wrapper><Porcentaje>0.00</Porcentaje></wrapper>"},
		{Porcentaje(50), "<wrapper><Porcentaje>0.50</Porcentaje></wrapper>"},
		{Porcentaje(99999), "<wrapper><Porcentaje>999.99</Porcentaje></wrapper>"},
	}

	for _, tt := range testCases {
		t.Run(tt.want, func(t *testing.T) {
			input := wrapper{Porcentaje: tt.input}
			xmlEnc, err := xml.Marshal(input)
			if err != nil {
				t.Fatalf("Porcentaje.MarshalXML returned error: %v", err)
			}

			if string(xmlEnc) != tt.want {
				t.Errorf("Porcentaje.MarshalXML() = %q, want %q", string(xmlEnc), tt.want)
			}
		})
	}

}
