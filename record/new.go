package record

func NewRegistroAlta(r RegistroAlta) (RegistroAlta, error) {
	r.IDVersion = IDVersionActual
	r.TipoHuella = TipoHuellaSHA256

	r.Huella = r.Fingerprint()

	if err := r.Validate(); err != nil {
		return RegistroAlta{}, err
	}
	return r, nil
}

func NewRegistroAnulacion(r RegistroAnulacion) (RegistroAnulacion, error) {
	r.IDVersion = IDVersionActual
	r.TipoHuella = TipoHuellaSHA256
	r.Huella = r.Fingerprint()

	if err := r.Validate(); err != nil {
		return RegistroAnulacion{}, err
	}
	return r, nil
}
