package record

import (
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
