package verifactu_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
	"github.com/cristianemek/go-verifactu/store/memory"
)

type storeConflictivo struct {
	llamadasUltimo int
	anexados       []*verifactu.Entry
	anterior       *verifactu.Entry
}

// Anexar implements [verifactu.Store].
func (s *storeConflictivo) Anexar(ctx context.Context, t verifactu.Tenant, e *verifactu.Entry) error {
	s.anexados = append(s.anexados, e)

	if len(s.anexados) == 1 {
		return verifactu.ErrConflictoDeSecuencia
	}

	return nil
}

// AnexarEnvio implements [verifactu.Store].
func (s *storeConflictivo) AnexarEnvio(ctx context.Context, t verifactu.Tenant, envio *verifactu.Envio) error {
	panic("unimplemented")
}

// Buscar implements [verifactu.Store].
func (s *storeConflictivo) Buscar(ctx context.Context, t verifactu.Tenant, idFactura verifactu.IDFactura, op verifactu.Operacion) (*verifactu.Entry, error) {
	return nil, verifactu.ErrNoEncontrado
}

func (s *storeConflictivo) Cadena(ctx context.Context, t verifactu.Tenant) ([]*verifactu.Entry, error) {
	return nil, nil
}

// Pendientes implements [verifactu.Store].
func (s *storeConflictivo) Pendientes(ctx context.Context, t verifactu.Tenant, limite int) ([]*verifactu.Entry, error) {
	panic("unimplemented")
}

// Ultimo implements [verifactu.Store].
func (s *storeConflictivo) Ultimo(ctx context.Context, t verifactu.Tenant) (*verifactu.Entry, error) {
	s.llamadasUltimo++

	if s.llamadasUltimo == 1 {
		return nil, verifactu.ErrNoEncontrado
	}
	return s.anterior, nil
}

// UltimoEnvio implements [verifactu.Store].
func (s *storeConflictivo) UltimoEnvio(ctx context.Context, t verifactu.Tenant) (*verifactu.Envio, error) {
	panic("unimplemented")
}

func (s *storeConflictivo) EnvioDe(ctx context.Context, t verifactu.Tenant, secuencia uint64) (*verifactu.Envio, error) {
	return nil, verifactu.ErrNoEncontrado
}

var _ verifactu.Store = (*storeConflictivo)(nil)

func TestAltaReintentaTrasConflicto(t *testing.T) {
	anterior := &verifactu.Entry{
		Secuencia: 1,
		Huella:    "huella",
		IDFactura: verifactu.IDFactura{
			NIF:      "12345678A",
			NumSerie: "1",
			Fecha:    record.Fecha(time.Now())},
	}

	store := &storeConflictivo{
		anterior: anterior,
	}

	engine, err := verifactu.New(verifactu.Config{Store: store, Now: fixedTime})
	if err != nil {
		t.Fatal(err)
	}

	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	_, err = engine.Alta(context.Background(), tenant, validRegistroAlta("001"))

	if err != nil {
		t.Fatalf("Alta() = %v, want nil", err)
	}

	if len(store.anexados) != 2 {
		t.Fatalf("Anexar() called %d times, want 2", len(store.anexados))
	}

	if store.anexados[0].Secuencia != 1 {
		t.Errorf("Anexar() first call Secuencia = %d, want 1", store.anexados[0].Secuencia)
	}

	if store.anexados[1].Secuencia != 2 {
		t.Errorf("Anexar() second call Secuencia = %d, want 2", store.anexados[1].Secuencia)
	}

	if store.anexados[0].Huella == store.anexados[1].Huella {
		t.Errorf("Anexar() first and second call Huella = %s, want different", store.anexados[0].Huella)
	}

	if store.anexados[1].Alta.Encadenamiento.RegistroAnterior.Huella != anterior.Huella {
		t.Errorf("Anexar() second call RegistroAnterior.Huella = %s, want %s", store.anexados[1].Alta.Encadenamiento.RegistroAnterior.Huella, anterior.Huella)
	}
}

type storeQueFalla struct {
	*memory.Store
	metodo string
	err    error
}

var _ verifactu.Store = (*storeQueFalla)(nil)

func (s *storeQueFalla) fallo(m string) error {
	if m == s.metodo {
		return s.err
	}
	return nil
}

func (s *storeQueFalla) Ultimo(ctx context.Context, t verifactu.Tenant) (*verifactu.Entry, error) {
	if err := s.fallo("Ultimo"); err != nil {
		return nil, err
	}
	return s.Store.Ultimo(ctx, t)
}

func (s *storeQueFalla) Buscar(ctx context.Context, t verifactu.Tenant, id verifactu.IDFactura, op verifactu.Operacion) (*verifactu.Entry, error) {
	if err := s.fallo("Buscar"); err != nil {
		return nil, err
	}
	return s.Store.Buscar(ctx, t, id, op)
}

func (s *storeQueFalla) Cadena(ctx context.Context, t verifactu.Tenant) ([]*verifactu.Entry, error) {
	if err := s.fallo("Cadena"); err != nil {
		return nil, err
	}
	return s.Store.Cadena(ctx, t)
}

func (s *storeQueFalla) Anexar(ctx context.Context, t verifactu.Tenant, e *verifactu.Entry) error {
	if err := s.fallo("Anexar"); err != nil {
		return err
	}
	return s.Store.Anexar(ctx, t, e)
}

func (s *storeQueFalla) EnvioDe(ctx context.Context, t verifactu.Tenant, secuencia uint64) (*verifactu.Envio, error) {
	if err := s.fallo("EnvioDe"); err != nil {
		return nil, err
	}
	return s.Store.EnvioDe(ctx, t, secuencia)
}

func (s *storeQueFalla) Pendientes(ctx context.Context, t verifactu.Tenant, limite int) ([]*verifactu.Entry, error) {
	if err := s.fallo("Pendientes"); err != nil {
		return nil, err
	}
	return s.Store.Pendientes(ctx, t, limite)
}

func (s *storeQueFalla) UltimoEnvio(ctx context.Context, t verifactu.Tenant) (*verifactu.Envio, error) {
	if err := s.fallo("UltimoEnvio"); err != nil {
		return nil, err
	}
	return s.Store.UltimoEnvio(ctx, t)
}

func (s *storeQueFalla) AnexarEnvio(ctx context.Context, t verifactu.Tenant, envio *verifactu.Envio) error {
	if err := s.fallo("AnexarEnvio"); err != nil {
		return err
	}
	return s.Store.AnexarEnvio(ctx, t, envio)
}

func TestErroresDelStore(t *testing.T) {
	ctx := context.Background()
	tenant := verifactu.Tenant{NIF: "89890001K", IDSistemaInformatico: "01"}

	alta := func(e *verifactu.Engine, _ *transporteFalso, _ *verifactu.Entry) error {
		_, err := e.Alta(ctx, tenant, validRegistroAlta("002"))
		return err
	}

	altaTrasRechazo := func(e *verifactu.Engine, _ *transporteFalso, _ *verifactu.Entry) error {
		if _, err := e.Alta(ctx, tenant, validRegistroAlta("002")); err != nil {
			return err
		}
		_, err := e.Alta(ctx, tenant, validRegistroAlta("002"), verifactu.TrasRechazo())
		return err
	}

	anular := func(e *verifactu.Engine, _ *transporteFalso, _ *verifactu.Entry) error {
		_, err := e.Anular(ctx, tenant, validRegistroAnulacion("001"))
		return err
	}

	remitir := func(e *verifactu.Engine, _ *transporteFalso, _ *verifactu.Entry) error {
		_, err := e.Remitir(ctx, tenant)
		return err
	}

	testCases := []struct {
		name   string
		metodo string
		err    error
		accion func(*verifactu.Engine, *transporteFalso, *verifactu.Entry) error
	}{
		{name: "alta, idempotencia", metodo: "Buscar", accion: alta},
		{name: "alta, encadenar", metodo: "Ultimo", accion: alta},
		{name: "alta, guardar", metodo: "Anexar", accion: alta},
		{name: "alta, reintentos agotados", metodo: "Anexar", err: verifactu.ErrConflictoDeSecuencia, accion: alta},
		{name: "alta tras rechazo, historial", metodo: "Cadena", accion: altaTrasRechazo},
		{name: "alta tras rechazo, envio", metodo: "EnvioDe", accion: altaTrasRechazo},
		{name: "anular, idempotencia", metodo: "Buscar", accion: anular},
		{name: "anular, encadenar", metodo: "Ultimo", accion: anular},
		{name: "anular, guardar", metodo: "Anexar", accion: anular},
		{name: "anular, reintentos agotados", metodo: "Anexar", err: verifactu.ErrConflictoDeSecuencia, accion: anular},
		{name: "remitir, pendientes", metodo: "Pendientes", accion: remitir},
		{name: "remitir, ultimo envio", metodo: "UltimoEnvio", accion: remitir},
		{
			name:   "remitir, guardar envio",
			metodo: "AnexarEnvio",
			accion: func(e *verifactu.Engine, tf *transporteFalso, primera *verifactu.Entry) error {
				tf.respuesta.RespuestaLinea = []record.RespuestaLinea{
					respuestaLineaPara(primera, record.TipoOperacionAlta, record.EstadoRegistroCorrecto),
				}
				return remitir(e, tf, primera)
			},
		},
		{
			name: "remitir, falla el transporte",
			accion: func(e *verifactu.Engine, tf *transporteFalso, primera *verifactu.Entry) error {
				tf.err = context.DeadlineExceeded
				return remitir(e, tf, primera)
			},
		},
		{name: "remitir, respuesta sin lineas", err: verifactu.ErrRespuestaDescuadrada, accion: remitir},
		{
			name: "remitir, solo anulaciones sin obligado",
			err:  verifactu.ErrObligadoDesconocido,
			accion: func(e *verifactu.Engine, _ *transporteFalso, _ *verifactu.Entry) error {
				otro := verifactu.Tenant{NIF: tenant.NIF, IDSistemaInformatico: "02"}
				if _, err := e.Anular(ctx, otro, validRegistroAnulacion("001")); err != nil {
					return err
				}
				_, err := e.Remitir(ctx, otro)
				return err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			want := tc.err
			if want == nil {
				want = context.DeadlineExceeded
			}

			store := &storeQueFalla{Store: memory.New()}
			tf := &transporteFalso{}

			engine, err := verifactu.New(verifactu.Config{Store: store, Transport: tf, Now: fixedTime})
			if err != nil {
				t.Fatalf("New() = %v", err)
			}

			primera, err := engine.Alta(ctx, tenant, validRegistroAlta("001"))
			if err != nil {
				t.Fatalf("Alta() = %v", err)
			}

			store.metodo, store.err = tc.metodo, want

			if err := tc.accion(engine, tf, primera); !errors.Is(err, want) {
				t.Errorf("err = %v, want %v", err, want)
			}
		})
	}
}
