package verifactu

type opcionesRegistro struct {
	esSubsanacion bool
	esTrasRechazo bool
}

// OpcionRegistro marks an Alta or Anular as a correction of a record that
// already exists. Passing one turns off idempotency for that call.
type OpcionRegistro func(*opcionesRegistro)

// ComoSubsanacion sends the record again with corrected data. The Engine fills
// in Subsanacion = "S".
func ComoSubsanacion() OpcionRegistro {
	return func(o *opcionesRegistro) {
		o.esSubsanacion = true
	}
}

// TrasRechazo sends the record again after the AEAT rejected the previous one.
// On an Alta it also implies ComoSubsanacion, and the Engine fills in
// RechazoPrevio = "X"; on an Anular the value is "S".
func TrasRechazo() OpcionRegistro {
	return func(o *opcionesRegistro) {
		o.esTrasRechazo = true
	}
}

func aplicarOpcionesRegistro(opciones ...OpcionRegistro) *opcionesRegistro {
	opts := &opcionesRegistro{}
	for _, opcion := range opciones {
		opcion(opts)
	}
	return opts
}
