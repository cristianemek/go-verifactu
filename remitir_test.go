package verifactu

import (
	"errors"
	"testing"
	"time"

	"github.com/cristianemek/go-verifactu/record"
)

func parEntradaLinea(secuencia uint64, numSerie string, op Operacion, tipo record.TipoOperacion) (*Entry, record.RespuestaLinea) {
	return &Entry{
			Secuencia: secuencia,
			Operacion: op,
			IDFactura: IDFactura{
				NIF:      "123123123A",
				NumSerie: numSerie,
				Fecha:    record.Fecha(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
		}, record.RespuestaLinea{
			IDFactura: record.IDFacturaExpedida{
				IDEmisorFactura:        "123123123A",
				NumSerieFactura:        numSerie,
				FechaExpedicionFactura: record.Fecha(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			Operacion: record.OperacionRespuesta{
				TipoOperacion: tipo,
			},
			EstadoRegistro: record.EstadoRegistroCorrecto,
		}
}

func TestParseTiempoEspera(t *testing.T) {
	testCases := []struct {
		input string
		want  time.Duration
	}{
		{
			input: "60",
			want:  60 * time.Second,
		},
		{
			input: "invalid",
			want:  60 * time.Second,
		},
		{
			input: "",
			want:  60 * time.Second,
		},
		{
			input: "120",
			want:  120 * time.Second,
		},
		{
			input: " 90 ",
			want:  90 * time.Second,
		},
		{
			input: "0",
			want:  60 * time.Second,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := parseTiempoEspera(tc.input)
			if got != tc.want {
				t.Errorf("parseTiempoEspera(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestResolverObligadoConOpcion(t *testing.T) {
	pendientes := []*Entry{
		{Alta: &record.RegistroAlta{NombreRazonEmisor: "DEL LOTE"}},
	}
	opts := aplicarOpcionesEnvio(ConObligado("EXPLICITO"))

	nombre, err := resolverObligado(pendientes, opts)
	if err != nil {
		t.Fatalf("resolverObligado returned an error: %v", err)
	}

	if nombre != "EXPLICITO" {
		t.Errorf("resolverObligado returned %q, want %q", nombre, "EXPLICITO")
	}

}

func TestResolverObligadoDelLote(t *testing.T) {
	pendientes := []*Entry{
		{Alta: &record.RegistroAlta{NombreRazonEmisor: "DEL LOTE"}},
	}
	opts := aplicarOpcionesEnvio()

	nombre, err := resolverObligado(pendientes, opts)
	if err != nil {
		t.Fatalf("resolverObligado returned an error: %v", err)
	}

	if nombre != "DEL LOTE" {
		t.Errorf("resolverObligado returned %q, want %q", nombre, "DEL LOTE")
	}
}

func TestResolverObligadoSinAlta(t *testing.T) {
	pendientes := []*Entry{
		{Anulacion: &record.RegistroAnulacion{}},
	}
	opts := aplicarOpcionesEnvio()

	nombre, err := resolverObligado(pendientes, opts)

	if !errors.Is(err, ErrObligadoDesconocido) {
		t.Fatalf("resolverObligado returned an unexpected error: %v, expected %v", err, ErrObligadoDesconocido)
	}
	if nombre != "" {
		t.Errorf("resolverObligado returned %q, want %q", nombre, "")
	}
}

func TestCasarLineasCorrecto(t *testing.T) {
	entryAlta, lineaAlta := parEntradaLinea(1, "001", OperacionAlta, record.TipoOperacionAlta)

	entryAnulacion, lineaAnulacion := parEntradaLinea(2, "002", OperacionAnulacion, record.TipoOperacionAnulacion)

	lineaAnulacion.EstadoRegistro = record.EstadoRegistroAceptadoConErrores
	lineaAnulacion.CodigoErrorRegistro = "2000"
	lineaAnulacion.DescripcionErrorRegistro = "El cálculo de la huella suministrada es incorrecta."

	pendientes := []*Entry{entryAlta, entryAnulacion}
	lineas := []record.RespuestaLinea{lineaAlta, lineaAnulacion}

	resultado, err := casarLineas(pendientes, lineas)

	if err != nil {
		t.Fatalf("casarLineas returned an error: %v", err)
	}

	if len(resultado) != 2 {
		t.Fatalf("casarLineas returned %d lines, want 2", len(resultado))
	}

	if resultado[0].Secuencia != 1 {
		t.Fatalf("casarLineas returned line 0 with Secuencia %d, want 1", resultado[0].Secuencia)
	}

	if resultado[0].Operacion != OperacionAlta {
		t.Fatalf("casarLineas returned line 0 with Operacion %v, want %v", resultado[0].Operacion, OperacionAlta)
	}

	if resultado[0].IDFactura.NumSerie != "001" {
		t.Fatalf("casarLineas returned line 0 with NumSerie %q, want %q", resultado[0].IDFactura.NumSerie, "001")
	}

	if resultado[1].Secuencia != 2 {
		t.Fatalf("casarLineas returned line 1 with Secuencia %d, want 2", resultado[1].Secuencia)
	}

	if resultado[1].Estado != record.EstadoRegistroAceptadoConErrores {
		t.Fatalf("casarLineas returned line 1 with Estado %v, want %v", resultado[1].Estado, record.EstadoRegistroAceptadoConErrores)
	}

	if resultado[1].CodigoError != "2000" {
		t.Fatalf("casarLineas returned line 1 with CodigoError %q, want %q", resultado[1].CodigoError, "2000")
	}

	if resultado[1].Descripcion != "El cálculo de la huella suministrada es incorrecta." {
		t.Fatalf("casarLineas returned line 1 with Descripcion %q, want %q", resultado[1].Descripcion, "El cálculo de la huella suministrada es incorrecta.")
	}

}

func TestCasarLineasLongitudDistinta(t *testing.T) {
	entryAlta, lineaAlta := parEntradaLinea(1, "001", OperacionAlta, record.TipoOperacionAlta)
	entryAnulacion, _ := parEntradaLinea(2, "002", OperacionAnulacion, record.TipoOperacionAnulacion)

	pendientes := []*Entry{entryAlta, entryAnulacion}
	lineas := []record.RespuestaLinea{lineaAlta}

	resultado, err := casarLineas(pendientes, lineas)

	if !errors.Is(err, ErrRespuestaDescuadrada) {
		t.Fatalf("casarLineas returned an unexpected error: %v, expected %v", err, ErrRespuestaDescuadrada)
	}

	if resultado != nil {
		t.Errorf("casarLineas returned a non-nil result: %v, want nil", resultado)
	}

}

func TestCasarLineasIDFacturaDistinta(t *testing.T) {
	entry, linea := parEntradaLinea(1, "001", OperacionAlta, record.TipoOperacionAlta)
	linea.IDFactura.NumSerieFactura = "999"

	pendientes := []*Entry{entry}
	lineas := []record.RespuestaLinea{linea}

	resultado, err := casarLineas(pendientes, lineas)

	if !errors.Is(err, ErrRespuestaDescuadrada) {
		t.Fatalf("casarLineas returned an unexpected error: %v, expected %v", err, ErrRespuestaDescuadrada)
	}

	if resultado != nil {
		t.Errorf("casarLineas returned a non-nil result: %v, want nil", resultado)
	}

}

func TestCasarLineasOperacionCruzada(t *testing.T) {
	entry, linea := parEntradaLinea(1, "001", OperacionAlta, record.TipoOperacionAnulacion)

	pendientes := []*Entry{entry}
	lineas := []record.RespuestaLinea{linea}

	resultado, err := casarLineas(pendientes, lineas)

	if !errors.Is(err, ErrRespuestaDescuadrada) {
		t.Fatalf("casarLineas returned an unexpected error: %v, expected %v", err, ErrRespuestaDescuadrada)
	}

	if resultado != nil {
		t.Errorf("casarLineas returned a non-nil result: %v, want nil", resultado)
	}
}
