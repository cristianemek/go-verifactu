package record

import (
	"errors"
	"fmt"
)

func (r RegistroAlta) Validate() error {
	var errs []error

	if r.Encadenamiento.PrimerRegistro == nil && r.Encadenamiento.RegistroAnterior == nil || (r.Encadenamiento.PrimerRegistro != nil && r.Encadenamiento.RegistroAnterior != nil) {
		errs = append(errs, fmt.Errorf("%w: RegistroAlta must have either PrimerRegistro or RegistroAnterior", ErrValidation))
	}

	if len(r.Desglose.DetalleDesglose) == 0 {
		errs = append(errs, fmt.Errorf("%w: Desglose len must be greater than 0", ErrValidation))
	}

	if len(r.Desglose.DetalleDesglose) > 12 {
		errs = append(errs, fmt.Errorf("%w: Desglose len must be less than or equal to 12", ErrValidation))
	}

	for i, d := range r.Desglose.DetalleDesglose {
		if d.CalificacionOperacion == nil && d.OperacionExenta == nil || (d.CalificacionOperacion != nil && d.OperacionExenta != nil) {
			errs = append(errs, fmt.Errorf("%w: DetalleDesglose[%d] must have either CalificacionOperacion or OperacionExenta", ErrValidation, i))
		}

	}

	if r.NombreRazonEmisor == "" {
		errs = append(errs, fmt.Errorf("%w: NombreRazonEmisor is required", ErrValidation))
	}

	if r.IDVersion == "" {
		errs = append(errs, fmt.Errorf("%w: IDVersion is required", ErrValidation))
	}

	if r.DescripcionOperacion == "" {
		errs = append(errs, fmt.Errorf("%w: DescripcionOperacion is required", ErrValidation))
	}

	if r.TipoHuella == "" {
		errs = append(errs, fmt.Errorf("%w: TipoHuella is required", ErrValidation))
	}

	if len(r.Huella) != 64 {
		errs = append(errs, fmt.Errorf("%w: Huella must be 64 characters long", ErrValidation))
	}

	if len(r.SistemaInformatico.IdSistemaInformatico) > 2 {
		errs = append(errs, fmt.Errorf("%w: IdSistemaInformatico must be 2 characters long", ErrValidation))
	}

	if len(r.DescripcionOperacion) > 500 {
		errs = append(errs, fmt.Errorf("%w: DescripcionOperacion must be 500 characters long", ErrValidation))
	}

	if r.SistemaInformatico.NIF == nil && r.SistemaInformatico.IDOtro == nil || (r.SistemaInformatico.NIF != nil && r.SistemaInformatico.IDOtro != nil) {
		errs = append(errs, fmt.Errorf("%w: SistemaInformatico must have either NIF or IDOtro", ErrValidation))
	}

	if r.SistemaInformatico.NIF != nil && !validNif(*r.SistemaInformatico.NIF) {
		errs = append(errs, fmt.Errorf("%w: SistemaInformatico NIF is invalid", ErrValidation))
	}

	if len(r.IDFactura.NumSerieFactura) == 0 || len(r.IDFactura.NumSerieFactura) > 60 {
		errs = append(errs, fmt.Errorf("%w: NumSerieFactura must be between 1 and 60 characters long", ErrValidation))
	}

	if len(r.NombreRazonEmisor) > 120 {
		errs = append(errs, fmt.Errorf("%w: NombreRazonEmisor must be 120 characters long", ErrValidation))
	}

	if !validNif(r.IDFactura.IDEmisorFactura) {
		errs = append(errs, fmt.Errorf("%w: IDEmisorFactura is invalid", ErrValidation))
	}

	if r.Destinatarios != nil {
		for i, d := range r.Destinatarios.IDDestinatario {
			if d.NIF == nil && d.IDOtro == nil || (d.NIF != nil && d.IDOtro != nil) {
				errs = append(errs, fmt.Errorf("%w: Destinatarios.IDDestinatario[%d] must have either NIF or IDOtro", ErrValidation, i))
			}

			if d.NIF != nil && !validNif(*d.NIF) {
				errs = append(errs, fmt.Errorf("%w: Destinatarios.IDDestinatario[%d].NIF is invalid", ErrValidation, i))
			}
		}
	}

	return errors.Join(errs...)
}

func (r RegistroAnulacion) Validate() error {
	var errs []error

	if r.Encadenamiento.PrimerRegistro == nil && r.Encadenamiento.RegistroAnterior == nil || (r.Encadenamiento.PrimerRegistro != nil && r.Encadenamiento.RegistroAnterior != nil) {
		errs = append(errs, fmt.Errorf("%w: RegistroAnulacion must have either PrimerRegistro or RegistroAnterior", ErrValidation))
	}

	if r.IDVersion == "" {
		errs = append(errs, fmt.Errorf("%w: IDVersion is required", ErrValidation))
	}

	if r.TipoHuella == "" {
		errs = append(errs, fmt.Errorf("%w: TipoHuella is required", ErrValidation))
	}

	if len(r.Huella) != 64 {
		errs = append(errs, fmt.Errorf("%w: Huella must be 64 characters long", ErrValidation))
	}

	if !validNif(r.IDFactura.IDEmisorFacturaAnulada) {
		errs = append(errs, fmt.Errorf("%w: IDEmisorFacturaAnulada is invalid", ErrValidation))
	}

	if r.IDFactura.NumSerieFacturaAnulada == "" || len(r.IDFactura.NumSerieFacturaAnulada) > 60 {
		errs = append(errs, fmt.Errorf("%w: NumSerieFacturaAnulada must be between 1 and 60 characters long", ErrValidation))
	}

	if r.SistemaInformatico.NIF == nil && r.SistemaInformatico.IDOtro == nil || (r.SistemaInformatico.NIF != nil && r.SistemaInformatico.IDOtro != nil) {
		errs = append(errs, fmt.Errorf("%w: SistemaInformatico must have either NIF or IDOtro", ErrValidation))
	}

	if r.SistemaInformatico.NIF != nil && !validNif(*r.SistemaInformatico.NIF) {
		errs = append(errs, fmt.Errorf("%w: SistemaInformatico NIF is invalid", ErrValidation))
	}

	if len(r.SistemaInformatico.IdSistemaInformatico) > 2 {
		errs = append(errs, fmt.Errorf("%w: IdSistemaInformatico must be 2 characters long", ErrValidation))
	}

	return errors.Join(errs...)
}

func validNif(nif string) bool {
	return len(nif) == 9
}
