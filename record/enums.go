package record

const IDVersionActual = "1.0"

// TipoHuella is the hashing algorithm used for the record fingerprint.
type TipoHuella string

const (
	TipoHuellaSHA256 TipoHuella = "01"
)

// SiNo is a yes/no flag. Several fields in the record use it.
type SiNo string

const (
	SiNoSi SiNo = "S"
	SiNoNo SiNo = "N"
)

// PrimerRegistroCadena marks the first record of a chain. Only "S" is valid:
// later records use the previous-record block instead.
type PrimerRegistroCadena string

const (
	PrimerRegistroCadenaSi PrimerRegistroCadena = "S"
)

// TipoFactura is the kind of invoice being registered.
type TipoFactura string

const (
	TipoFacturaCompleta                 TipoFactura = "F1"
	TipoFacturaSimplificada             TipoFactura = "F2"
	TipoFacturaSustitutivaSimplificadas TipoFactura = "F3"

	// R1 to R4 only differ in the article of the VAT law they correct under:
	// 80.1 and 80.2, 80.3, 80.4, and anything else. R5 corrects a simplified
	// invoice.
	TipoFacturaRectificativaR1 TipoFactura = "R1"
	TipoFacturaRectificativaR2 TipoFactura = "R2"
	TipoFacturaRectificativaR3 TipoFactura = "R3"
	TipoFacturaRectificativaR4 TipoFactura = "R4"
	TipoFacturaRectificativaR5 TipoFactura = "R5"
)

// ClaveTipoRectificativa says whether a corrective invoice replaces the
// original or only states the difference.
type ClaveTipoRectificativa string

const (
	ClaveTipoRectificativaSustitutiva ClaveTipoRectificativa = "S"
	ClaveTipoRectificativaIncremental ClaveTipoRectificativa = "I"
)

// Impuesto is the tax applied to a breakdown line.
type Impuesto string

const (
	ImpuestoIVA  Impuesto = "01"
	ImpuestoIPSI Impuesto = "02"
	ImpuestoIGIC Impuesto = "03"
	// "04" does not exist in the schema.
	ImpuestoOtros Impuesto = "05"
)

// CalificacionOperacion is the VAT treatment of the operation. A breakdown
// line carries either this or OperacionExenta, never both.
type CalificacionOperacion string

const (
	CalificacionOperacionSujetaNoExentaSinISP       CalificacionOperacion = "S1"
	CalificacionOperacionSujetaNoExentaConISP       CalificacionOperacion = "S2"
	CalificacionOperacionoSujetaArt7_14_Otros       CalificacionOperacion = "N1"
	CalificacionOperacionNoSujetaReglasLocalizacion CalificacionOperacion = "N2"
)

// OperacionExenta is the reason an operation is exempt from VAT.
type OperacionExenta string

const (
	OperacionExentaE1 OperacionExenta = "E1"
	OperacionExentaE2 OperacionExenta = "E2"
	OperacionExentaE3 OperacionExenta = "E3"
	OperacionExentaE4 OperacionExenta = "E4"
	OperacionExentaE5 OperacionExenta = "E5"
	OperacionExentaE6 OperacionExenta = "E6"
	OperacionExentaE7 OperacionExenta = "E7"
	OperacionExentaE8 OperacionExenta = "E8"
)

// ClaveRegimen is the VAT or IGIC regime of the operation. The schema calls it
// IdOperacionesTrascendenciaTributariaType and gives no descriptions see the
// AEAT record design document for what each number means.
type ClaveRegimen string

const (
	ClaveRegimenGeneral ClaveRegimen = "01"
	ClaveRegimen02      ClaveRegimen = "02"
	ClaveRegimen03      ClaveRegimen = "03"
	ClaveRegimen04      ClaveRegimen = "04"
	ClaveRegimen05      ClaveRegimen = "05"
	ClaveRegimen06      ClaveRegimen = "06"
	ClaveRegimen07      ClaveRegimen = "07"
	ClaveRegimen08      ClaveRegimen = "08"
	ClaveRegimen09      ClaveRegimen = "09"
	ClaveRegimen10      ClaveRegimen = "10"
	ClaveRegimen11      ClaveRegimen = "11"
	// "12", "13" and "16" do not exist in the schema.
	ClaveRegimen14 ClaveRegimen = "14"
	ClaveRegimen15 ClaveRegimen = "15"
	ClaveRegimen17 ClaveRegimen = "17"
	ClaveRegimen18 ClaveRegimen = "18"
	ClaveRegimen19 ClaveRegimen = "19"
	ClaveRegimen20 ClaveRegimen = "20"
	ClaveRegimen21 ClaveRegimen = "21"
)

// RechazoPrevio says whether the AEAT rejected this record before. Used in
// registration records.
type RechazoPrevio string

const (
	RechazoPrevioNo RechazoPrevio = "N"
	RechazoPrevioSi RechazoPrevio = "S"
	// The record does not exist at the AEAT at all. Not allowed on registration.
	RechazoPrevioNoExiste RechazoPrevio = "X"
)

// RechazoPrevioAnulacion is the same flag for cancellation records, where "X"
// is not allowed.
type RechazoPrevioAnulacion string

const (
	RechazoPrevioAnulacionSi RechazoPrevioAnulacion = "S"
	RechazoPrevioAnulacionNo RechazoPrevioAnulacion = "N"
)

// TercerosODestinatario says who issued the invoice when it was not the
// obligor.
type TercerosODestinatario string

const (
	TercerosODestinatarioDestinatario TercerosODestinatario = "D"
	TercerosODestinatarioTercero      TercerosODestinatario = "T"
)

// GeneradoPor says who generated a cancellation record.
type GeneradoPor string

const (
	GeneradoPorExpedidor    GeneradoPor = "E"
	GeneradoPorDestinatario GeneradoPor = "D"
	GeneradoPorTercero      GeneradoPor = "T"
)

// PersonaFisicaJuridicaIDType is the kind of fiscal ID used when the party has
// no Spanish NIF. There is no "01": for a Spanish NIF, use the NIF field.
type PersonaFisicaJuridicaIDType string

const (
	PersonaFisicaJuridicaIDTypeNIFIVA            PersonaFisicaJuridicaIDType = "02"
	PersonaFisicaJuridicaIDTypePasaporte         PersonaFisicaJuridicaIDType = "03"
	PersonaFisicaJuridicaIDTypeIDPaisResidencia  PersonaFisicaJuridicaIDType = "04"
	PersonaFisicaJuridicaIDTypeCertificadoResid  PersonaFisicaJuridicaIDType = "05"
	PersonaFisicaJuridicaIDTypeOtroDocProbatorio PersonaFisicaJuridicaIDType = "06"
	PersonaFisicaJuridicaIDTypeNoCensado         PersonaFisicaJuridicaIDType = "07"
)
