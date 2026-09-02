package verifactu

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cristianemek/go-verifactu/record"
)

const (
	// maxRegistrosPorEnvio is the maximum number of records that can be sent in a single submission, limitation of the AEAT.
	maxRegistrosPorEnvio   = 1000
	tiempoEsperaPorDefecto = 60 * time.Second
)

func resolverObligado(pendientes []*Entry, opts *opcionesEnvio) (string, error) {
	if opts != nil && opts.obligado != "" {
		return opts.obligado, nil
	}

	for _, e := range pendientes {
		if e.Alta != nil {
			return e.Alta.NombreRazonEmisor, nil
		}
	}

	return "", ErrObligadoDesconocido
}

func construirLote(t Tenant, pendientes []*Entry, nombre string) record.RegFactuSistemaFacturacion {
	cabecera := record.Cabecera{
		ObligadoEmision: record.PersonaFisicaJuridicaES{
			NombreRazon: nombre,
			NIF:         t.NIF,
		},
	}

	registros := make([]record.RegistroFactura, 0, len(pendientes))

	for _, e := range pendientes {
		registros = append(registros, record.RegistroFactura{
			RegistroAlta:      e.Alta,
			RegistroAnulacion: e.Anulacion,
		})
	}

	return record.RegFactuSistemaFacturacion{
		Cabecera:        cabecera,
		RegistroFactura: registros,
	}
}

// Remitir sends one batch of pending records to the AEAT and stores the answer.
// It holds the tenant lock during the call, so Alta and Anular wait for it.
// If it fails after the call, the records may already be filed and will be resent.
func (e *Engine) Remitir(ctx context.Context, t Tenant, opciones ...OpcionEnvio) (*Envio, error) {
	if e.transport == nil {
		return nil, ErrTransportRequerido
	}

	opts := aplicarOpcionesEnvio(opciones...)

	release := e.lock(t)

	defer release()

	pendientes, err := e.store.Pendientes(ctx, t, maxRegistrosPorEnvio)
	if err != nil {
		return nil, err
	}

	if len(pendientes) == 0 {
		return nil, ErrSinPendientes
	}

	if len(pendientes) < maxRegistrosPorEnvio {
		ultimoEnvio, err := e.store.UltimoEnvio(ctx, t)

		switch {
		case errors.Is(err, ErrNoEncontrado):
			// No previous send, we can send the pending entries.
		case err != nil:
			return nil, err
		default:
			ultimoEnvioTiempoEspera := ultimoEnvio.Instante.Add(ultimoEnvio.TiempoEspera)
			if ultimoEnvioTiempoEspera.After(e.now()) {
				return nil, fmt.Errorf("%w: last send was at %s, must wait until %s", ErrEsperaActiva, ultimoEnvio.Instante, ultimoEnvioTiempoEspera)
			}
		}
	}

	nombre, err := resolverObligado(pendientes, opts)
	if err != nil {
		return nil, err
	}

	lote := construirLote(t, pendientes, nombre)

	respuesta, err := e.transport.Remitir(ctx, t, lote)
	if err != nil {
		return nil, err
	}

	lineas, err := casarLineas(pendientes, respuesta.RespuestaLinea)
	if err != nil {
		return nil, err
	}

	envio := Envio{
		Instante:     e.now(),
		CSV:          respuesta.CSV,
		EstadoEnvio:  respuesta.EstadoEnvio,
		TiempoEspera: parseTiempoEspera(respuesta.TiempoEsperaEnvio),
		Lineas:       lineas,
	}

	if respuesta.DatosPresentacion != nil {
		envio.NIFPresentador = respuesta.DatosPresentacion.NIFPresentador
		envio.TimestampPresentacion = respuesta.DatosPresentacion.TimestampPresentacion
	}

	err = e.store.AnexarEnvio(ctx, t, &envio)
	if err != nil {
		return nil, err
	}

	return &envio, nil
}

func tipoOperacion(op Operacion) record.TipoOperacion {
	switch op {
	case OperacionAlta:
		return record.TipoOperacionAlta
	case OperacionAnulacion:
		return record.TipoOperacionAnulacion
	default:
		return ""
	}
}

func idFacturaDesde(r record.IDFacturaExpedida) IDFactura {
	return IDFactura{
		NIF:      r.IDEmisorFactura,
		NumSerie: r.NumSerieFactura,
		Fecha:    r.FechaExpedicionFactura,
	}
}

func casarLineas(pendientes []*Entry, lineas []record.RespuestaLinea) ([]LineaEnvio, error) {
	if len(lineas) != len(pendientes) {
		return nil, fmt.Errorf("%w: expected %d lines, got %d", ErrRespuestaDescuadrada, len(pendientes), len(lineas))
	}

	resultado := make([]LineaEnvio, 0, len(lineas))

	for i, linea := range lineas {
		entry := pendientes[i]

		if !entry.IDFactura.Equal(idFacturaDesde(linea.IDFactura)) || linea.Operacion.TipoOperacion != tipoOperacion(entry.Operacion) {
			return nil, fmt.Errorf("%w: line %d does not match entry", ErrRespuestaDescuadrada, i)
		}
		resultado = append(resultado, LineaEnvio{
			IDFactura:   entry.IDFactura,
			Operacion:   entry.Operacion,
			Secuencia:   entry.Secuencia,
			Estado:      linea.EstadoRegistro,
			CodigoError: linea.CodigoErrorRegistro,
			Descripcion: linea.DescripcionErrorRegistro,
			Duplicado:   linea.RegistroDuplicado,
		})
	}

	return resultado, nil
}

func parseTiempoEspera(s string) time.Duration {
	i, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || i <= 0 {
		return tiempoEsperaPorDefecto
	}
	return time.Duration(i) * time.Second
}
