package main

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/aeat"
	"github.com/cristianemek/go-verifactu/record"
)

type servidor struct {
	engine  *verifactu.Engine
	cliente *aeat.Client
	tenants map[string]TenantConfig
	sistema string
}

type respuestaRegistro struct {
	Entry  *verifactu.Entry `json:"entry"`
	Avisos []string         `json:"avisos"`
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
		nif := r.PathValue("nif")
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
		NIF:                  r.PathValue("nif"),
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
		NIF:      r.PathValue("nif"),
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
	responderJSON(w, http.StatusOK, respuestaRegistro{
		Entry:  entry,
		Avisos: []string{},
	})
}
