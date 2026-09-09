package record

import (
	"errors"
	"testing"
)

func TestAvisos(t *testing.T) {

	testCases := []struct {
		name         string
		desglose     Desglose
		cuotaTotal   Amount
		importeTotal Amount
		want         int
	}{
		{
			name:         "cuadra exacto",
			cuotaTotal:   2100,
			importeTotal: 12100,
			desglose: Desglose{
				DetalleDesglose: []DetalleDesglose{
					{
						BaseImponibleOimporteNoSujeto: 10000,
						CuotaRepercutida:              Ptr(Amount(2100)),
					},
				},
			},
			want: 0,
		},
		{
			name:         "cuadra dentro de tolerancia",
			cuotaTotal:   3000,
			importeTotal: 12100,
			desglose: Desglose{
				DetalleDesglose: []DetalleDesglose{
					{
						BaseImponibleOimporteNoSujeto: 10000,
						CuotaRepercutida:              Ptr(Amount(2100)),
					},
				},
			},
			want: 0,
		},
		{
			name:         "justo en el limite de tolerancia",
			cuotaTotal:   3100,
			importeTotal: 12100,
			desglose: Desglose{
				DetalleDesglose: []DetalleDesglose{
					{
						BaseImponibleOimporteNoSujeto: 10000,
						CuotaRepercutida:              Ptr(Amount(2100)),
					},
				},
			},
			want: 0,
		}, {
			name:         "fuera de tolerancia",
			cuotaTotal:   3101,
			importeTotal: 12100,
			desglose: Desglose{
				DetalleDesglose: []DetalleDesglose{
					{
						BaseImponibleOimporteNoSujeto: 10000,
						CuotaRepercutida:              Ptr(Amount(2100)),
					},
				},
			},
			want: 1,
		},
		{
			name:         "importe descuadrado",
			cuotaTotal:   3100,
			importeTotal: 13101,
			desglose: Desglose{
				DetalleDesglose: []DetalleDesglose{
					{
						BaseImponibleOimporteNoSujeto: 10000,
						CuotaRepercutida:              Ptr(Amount(2100)),
					},
				},
			},
			want: 1,
		},
		{
			name:         "los dos mal",
			cuotaTotal:   3101,
			importeTotal: 13101,
			desglose: Desglose{
				DetalleDesglose: []DetalleDesglose{
					{
						BaseImponibleOimporteNoSujeto: 10000,
						CuotaRepercutida:              Ptr(Amount(2100)),
					},
				},
			},
			want: 2,
		},
		{
			name:         "regimen parcial",
			cuotaTotal:   3101,
			importeTotal: 13101,
			desglose: Desglose{
				DetalleDesglose: []DetalleDesglose{
					{
						BaseImponibleOimporteNoSujeto: 10000,
						CuotaRepercutida:              Ptr(Amount(2100)),
						ClaveRegimen:                  Ptr(ClaveRegimen03),
					},
				},
			},
			want: 0,
		},
		{
			name:         "sin regimen parcial",
			cuotaTotal:   3101,
			importeTotal: 12100,
			desglose: Desglose{
				DetalleDesglose: []DetalleDesglose{
					{
						BaseImponibleOimporteNoSujeto: 10000,
						CuotaRepercutida:              Ptr(Amount(2100)),
						ClaveRegimen:                  nil,
					},
				},
			},
			want: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := RegistroAlta{
				Desglose:     tc.desglose,
				CuotaTotal:   tc.cuotaTotal,
				ImporteTotal: tc.importeTotal,
			}

			got := r.Avisos()

			if len(got) != tc.want {
				t.Errorf("Avisos() = %d, want %d", len(got), tc.want)
			}

			for i, a := range got {
				if !errors.Is(a, ErrAviso) {
					t.Errorf("Avisos()[%d] = %v, want ErrAviso", i, a)
				}
			}
		})
	}
}
