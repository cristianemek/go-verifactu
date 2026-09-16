package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/aeat"
	"github.com/cristianemek/go-verifactu/record"
)

type servidor struct {
	engine   *verifactu.Engine
	clientes map[string]*aeat.Client
	tenants  map[string]TenantConfig
	sistema  string
	avisar   chan struct{}
}

type respuestaRegistro struct {
	Entry  *verifactu.Entry `json:"entry"`
	Avisos []string         `json:"avisos"`
	AEAT   *resultadoAEAT   `json:"aeat,omitempty"`
}

type resultadoAEAT struct {
	Estado      string `json:"estado"`
	Codigo      string `json:"codigo,omitempty"`
	Descripcion string `json:"descripcion,omitempty"`
	CSV         string `json:"csv,omitempty"`
}

func (s *servidor) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func responderError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (s *servidor) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nif := nifDeRuta(r)
		t, ok := s.tenants[nif]
		token, haveBearer := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")

		valid := subtle.ConstantTimeCompare([]byte(token), []byte(t.Token)) == 1

		if !ok || !haveBearer || !valid {
			responderError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		next(w, r)
	}
}

func (s *servidor) tenant(r *http.Request) verifactu.Tenant {
	return verifactu.Tenant{
		NIF:                  nifDeRuta(r),
		IDSistemaInformatico: s.sistema,
	}
}

func responderJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func mapearError(err error) (int, string) {
	var espera *verifactu.ErrorEspera

	switch {
	case errors.As(err, &espera):
		return http.StatusTooManyRequests, espera.Error()
	case errors.Is(err, verifactu.ErrSinPendientes):
		return http.StatusOK, err.Error()
	case errors.Is(err, record.ErrValidation):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, verifactu.ErrOpcionNoAplicable):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, verifactu.ErrNoEncontrado):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, verifactu.ErrFaultCliente):
		return http.StatusUnprocessableEntity, err.Error()
	case errors.Is(err, verifactu.ErrFaultServidor):
		return http.StatusBadGateway, err.Error()
	case errors.Is(err, aeat.ErrAccesoDenegado):
		return http.StatusServiceUnavailable, err.Error()
	case errors.Is(err, verifactu.ErrConflictoDeSecuencia):
		return http.StatusConflict, err.Error()
	default:
		return http.StatusInternalServerError, "Internal Server Error"
	}

}

func (s *servidor) alta(w http.ResponseWriter, r *http.Request) {
	var reg record.RegistroAlta

	err := json.NewDecoder(r.Body).Decode(&reg)
	if err != nil {
		responderError(w, http.StatusBadRequest, "Invalid json: "+err.Error())
		return
	}

	var opts []verifactu.OpcionRegistro

	if r.URL.Query().Has("subsanacion") {
		opts = append(opts, verifactu.ComoSubsanacion())
	}

	if r.URL.Query().Has("tras_rechazo") {
		opts = append(opts, verifactu.TrasRechazo())
	}

	avisos := []string{}

	for _, a := range reg.Avisos() {
		avisos = append(avisos, a.Error())
	}

	entry, err := s.engine.Alta(r.Context(), s.tenant(r), reg, opts...)

	if err != nil {
		status, msg := mapearError(err)
		if status == http.StatusInternalServerError {
			slog.Error("alta", "error", err)
		}

		responderError(w, status, msg)
		return
	}

	s.tocarTimbre()

	responderJSON(w, http.StatusCreated, respuestaRegistro{
		Entry:  entry,
		Avisos: avisos,
	})
}

func (s *servidor) anulacion(w http.ResponseWriter, r *http.Request) {
	var reg record.RegistroAnulacion

	err := json.NewDecoder(r.Body).Decode(&reg)
	if err != nil {
		responderError(w, http.StatusBadRequest, "Invalid json: "+err.Error())
		return
	}

	var opts []verifactu.OpcionRegistro

	if r.URL.Query().Has("subsanacion") {
		opts = append(opts, verifactu.ComoSubsanacion())
	}

	if r.URL.Query().Has("tras_rechazo") {
		opts = append(opts, verifactu.TrasRechazo())
	}

	entry, err := s.engine.Anular(r.Context(), s.tenant(r), reg, opts...)

	if err != nil {
		status, msg := mapearError(err)
		if status == http.StatusInternalServerError {
			slog.Error("anulacion", "error", err)
		}

		responderError(w, status, msg)
		return
	}

	s.tocarTimbre()

	responderJSON(w, http.StatusCreated, respuestaRegistro{
		Entry:  entry,
		Avisos: []string{},
	})
}

func (s *servidor) estado(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	serie := q.Get("serie")
	fecha := q.Get("fecha")
	op := q.Get("op")

	if serie == "" || fecha == "" {
		responderError(w, http.StatusBadRequest, "Missing required query parameters: serie and fecha")
		return
	}

	f, err := record.ParseFecha(fecha)
	if err != nil {
		responderError(w, http.StatusBadRequest, "Invalid date format: "+err.Error())
		return
	}

	if op == "" {
		op = string(verifactu.OperacionAlta)
	}

	if op != string(verifactu.OperacionAlta) && op != string(verifactu.OperacionAnulacion) {
		responderError(w, http.StatusBadRequest, "Invalid operation type: "+op)
		return
	}

	id := verifactu.IDFactura{
		NIF:      nifDeRuta(r),
		NumSerie: serie,
		Fecha:    f,
	}

	entry, err := s.engine.Estado(r.Context(), s.tenant(r), id, verifactu.Operacion(op))

	if err != nil {
		status, msg := mapearError(err)
		if status == http.StatusInternalServerError {
			slog.Error("estado", "error", err)
		}
		responderError(w, status, msg)
		return
	}

	aeat := &resultadoAEAT{
		Estado: "Pendiente",
	}

	envio, err := s.engine.EnvioDe(r.Context(), s.tenant(r), entry.Secuencia)

	switch {
	case errors.Is(err, verifactu.ErrNoEncontrado):
	case err != nil:
		slog.Error("estado", "error", err)
		responderError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	default:
		for _, linea := range envio.Lineas {
			if linea.Secuencia == entry.Secuencia {
				aeat.Estado = string(linea.Estado)
				aeat.Codigo = linea.CodigoError
				aeat.Descripcion = linea.Descripcion

				if linea.Estado != record.EstadoRegistroIncorrecto {
					aeat.CSV = envio.CSV
				}
				break
			}
		}
	}

	responderJSON(w, http.StatusOK, respuestaRegistro{
		Entry:  entry,
		Avisos: []string{},
		AEAT:   aeat,
	})
}

func (s *servidor) remitir(ctx context.Context, nif string) error {
	t, ok := s.tenants[nif]
	if !ok {
		return fmt.Errorf("tenant not found: %s", nif)
	}

	tenant := verifactu.Tenant{
		NIF:                  nif,
		IDSistemaInformatico: s.sistema,
	}

	envio, err := s.engine.Remitir(ctx, tenant, verifactu.ConObligado(t.Nombre))

	if errors.Is(err, verifactu.ErrSinPendientes) {
		return nil
	}

	var espera *verifactu.ErrorEspera

	if errors.As(err, &espera) {
		slog.Info("en espera", "nif", nif, "hasta", espera.Hasta.Format("15:04:05"), "faltan", espera.Restante.Round(time.Second))

		time.AfterFunc(espera.Restante+time.Second, s.tocarTimbre)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error remitting: %w", err)
	}

	var rechazados int

	for _, linea := range envio.Lineas {
		if linea.Estado != record.EstadoRegistroCorrecto {
			rechazados++
			slog.Warn("registro no aceptado", "nif", nif, "serie", linea.IDFactura.NumSerie, "estado", linea.Estado, "codigo", linea.CodigoError, "descripcion", linea.Descripcion)
		}

	}

	nivel := slog.LevelInfo
	if envio.EstadoEnvio != record.EstadoEnvioCorrecto {
		nivel = slog.LevelWarn
	}

	slog.Log(ctx, nivel, "remitido", "nif", nif, "csv", envio.CSV, "estado", envio.EstadoEnvio, "correctos", len(envio.Lineas)-rechazados, "rechazados", rechazados)

	return nil
}

func (s *servidor) conexion(w http.ResponseWriter, r *http.Request) {
	cliente, ok := s.clientes[nifDeRuta(r)]
	if !ok {
		responderError(w, http.StatusNotFound, "Tenant not found")
		return
	}

	err := cliente.ProbarConexion(r.Context())
	if err != nil {
		status, msg := mapearError(err)
		if status == http.StatusInternalServerError {
			slog.Error("conexion", "error", err)
		}

		responderError(w, status, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *servidor) bucleRemision(ctx context.Context, cada time.Duration) {
	ticker := time.NewTicker(cada)
	defer ticker.Stop()

	bloqueados := map[string]bool{}

	s.vuelta(bloqueados)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.vuelta(bloqueados)
		case <-s.avisar:
			s.vuelta(bloqueados)
		}
	}
}

func nifDeRuta(r *http.Request) string {
	return strings.ToUpper(r.PathValue("nif"))
}

func (s *servidor) vuelta(bloqueados map[string]bool) {
	ctxVuelta, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

	for nif := range s.tenants {
		if bloqueados[nif] {
			continue
		}

		if err := s.remitir(ctxVuelta, nif); err != nil {
			if errors.Is(err, verifactu.ErrFaultCliente) {
				bloqueados[nif] = true
				slog.Error("Fault del cliente: no se reintenta hasta reiniciar", "nif", nif, "error", err, "bloqueado", true)
			} else {
				slog.Error("remitir", "nif", nif, "error", err)
			}
		}
	}

	cancel()
}

func (s *servidor) tocarTimbre() {
	select {
	case s.avisar <- struct{}{}:
	default:
	}
}
