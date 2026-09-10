package record

import "encoding/xml"

func Ptr[T any](v T) *T {
	return &v
}

const (
	nsSuministroLR          = "https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroLR.xsd"
	nsSuministroInformacion = "https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroInformacion.xsd"
)

type IDFacturaExpedida struct {
	IDEmisorFactura        string `xml:"IDEmisorFactura"`
	NumSerieFactura        string `xml:"NumSerieFactura"`
	FechaExpedicionFactura Fecha  `xml:"FechaExpedicionFactura"`
}

type IDFacturaExpedidaBaja struct {
	IDEmisorFacturaAnulada        string `xml:"IDEmisorFacturaAnulada"`
	NumSerieFacturaAnulada        string `xml:"NumSerieFacturaAnulada"`
	FechaExpedicionFacturaAnulada Fecha  `xml:"FechaExpedicionFacturaAnulada"`
}

type PersonaFisicaJuridicaES struct {
	NombreRazon string `xml:"NombreRazon"`
	NIF         string `xml:"NIF"`
}

type EncadenamientoFacturaAnterior struct {
	IDEmisorFactura        string `xml:"IDEmisorFactura"`
	NumSerieFactura        string `xml:"NumSerieFactura"`
	FechaExpedicionFactura Fecha  `xml:"FechaExpedicionFactura"`
	Huella                 string `xml:"Huella"`
}

type Encadenamiento struct {
	PrimerRegistro   *PrimerRegistroCadena          `xml:"PrimerRegistro,omitempty" json:",omitempty"`
	RegistroAnterior *EncadenamientoFacturaAnterior `xml:"RegistroAnterior,omitempty" json:",omitempty"`
}

type SistemaInformatico struct {
	NombreRazon              string  `xml:"NombreRazon"`
	NIF                      *string `xml:"NIF,omitempty" json:",omitempty"`
	IDOtro                   *IDOtro `xml:"IDOtro,omitempty" json:",omitempty"`
	NombreSistemaInformatico string  `xml:"NombreSistemaInformatico"`
	// IdSistemaInformatico is limited to 2 characters by the schema.
	IdSistemaInformatico        string `xml:"IdSistemaInformatico"`
	Version                     string `xml:"Version"`
	NumeroInstalacion           string `xml:"NumeroInstalacion"`
	TipoUsoPosibleSoloVerifactu SiNo   `xml:"TipoUsoPosibleSoloVerifactu"`
	TipoUsoPosibleMultiOT       SiNo   `xml:"TipoUsoPosibleMultiOT"`
	IndicadorMultiplesOT        SiNo   `xml:"IndicadorMultiplesOT"`
}

// IDOtro identifies a party that has no Spanish NIF.
type IDOtro struct {
	CodigoPais *string                     `xml:"CodigoPais,omitempty" json:",omitempty"`
	IDType     PersonaFisicaJuridicaIDType `xml:"IDType"`
	ID         string                      `xml:"ID"`
}

func NewEncadenamientoPrimerRegistro() Encadenamiento {
	return Encadenamiento{
		PrimerRegistro: Ptr(PrimerRegistroCadenaSi),
	}
}

func NewEncadenamientoRegistroAnterior(idEmisor string, numSerie string, fechaExpedicion Fecha, huella string) Encadenamiento {
	return Encadenamiento{
		RegistroAnterior: Ptr(EncadenamientoFacturaAnterior{
			IDEmisorFactura:        idEmisor,
			NumSerieFactura:        numSerie,
			FechaExpedicionFactura: fechaExpedicion,
			Huella:                 huella,
		}),
	}
}

// DetalleDesglose is one breakdown line. It carries either
// CalificacionOperacion or OperacionExenta, never both.
type DetalleDesglose struct {
	Impuesto                      *Impuesto              `xml:"Impuesto,omitempty" json:",omitempty"`
	ClaveRegimen                  *ClaveRegimen          `xml:"ClaveRegimen,omitempty" json:",omitempty"`
	CalificacionOperacion         *CalificacionOperacion `xml:"CalificacionOperacion,omitempty" json:",omitempty"`
	OperacionExenta               *OperacionExenta       `xml:"OperacionExenta,omitempty" json:",omitempty"`
	TipoImpositivo                *Porcentaje            `xml:"TipoImpositivo,omitempty" json:",omitempty"`
	BaseImponibleOimporteNoSujeto Amount                 `xml:"BaseImponibleOimporteNoSujeto"`
	BaseImponibleACoste           *Amount                `xml:"BaseImponibleACoste,omitempty" json:",omitempty"`
	CuotaRepercutida              *Amount                `xml:"CuotaRepercutida,omitempty" json:",omitempty"`
	TipoRecargoEquivalencia       *Porcentaje            `xml:"TipoRecargoEquivalencia,omitempty" json:",omitempty"`
	CuotaRecargoEquivalencia      *Amount                `xml:"CuotaRecargoEquivalencia,omitempty" json:",omitempty"`
}

// Desglose holds the breakdown lines. The schema allows up to 12.
type Desglose struct {
	DetalleDesglose []DetalleDesglose `xml:"DetalleDesglose" json:",omitempty"`
}

// RegistroAlta is a registration record: the invoice as reported to the AEAT.
// Field order matters — the schema uses a sequence and xmllint will reject a
// different order.
type RegistroAlta struct {
	XMLName xml.Name `xml:"https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroInformacion.xsd RegistroAlta" json:"-"`

	IDVersion         string            `xml:"IDVersion"`
	IDFactura         IDFacturaExpedida `xml:"IDFactura"`
	RefExterna        *string           `xml:"RefExterna,omitempty" json:",omitempty"`
	NombreRazonEmisor string            `xml:"NombreRazonEmisor"`
	Subsanacion       *SiNo             `xml:"Subsanacion,omitempty" json:",omitempty"`
	RechazoPrevio     *RechazoPrevio    `xml:"RechazoPrevio,omitempty" json:",omitempty"`
	TipoFactura       TipoFactura       `xml:"TipoFactura"`

	TipoRectificativa    *ClaveTipoRectificativa `xml:"TipoRectificativa,omitempty" json:",omitempty"`
	FacturasRectificadas *FacturasRectificadas   `xml:"FacturasRectificadas,omitempty" json:",omitempty"`
	FacturasSustituidas  *FacturasSustituidas    `xml:"FacturasSustituidas,omitempty" json:",omitempty"`
	ImporteRectificacion *DesgloseRectificacion  `xml:"ImporteRectificacion,omitempty" json:",omitempty"`

	FechaOperacion                      *Fecha                 `xml:"FechaOperacion,omitempty" json:",omitempty"`
	DescripcionOperacion                string                 `xml:"DescripcionOperacion"`
	FacturaSimplificadaArt7273          *SiNo                  `xml:"FacturaSimplificadaArt7273,omitempty" json:",omitempty"`
	FacturaSinIdentifDestinatarioArt61d *SiNo                  `xml:"FacturaSinIdentifDestinatarioArt61d,omitempty" json:",omitempty"`
	Macrodato                           *SiNo                  `xml:"Macrodato,omitempty" json:",omitempty"`
	EmitidaPorTerceroODestinatario      *TercerosODestinatario `xml:"EmitidaPorTerceroODestinatario,omitempty" json:",omitempty"`
	Tercero                             *PersonaFisicaJuridica `xml:"Tercero,omitempty" json:",omitempty"`
	Destinatarios                       *Destinatarios         `xml:"Destinatarios,omitempty" json:",omitempty"`
	Cupon                               *SiNo                  `xml:"Cupon,omitempty" json:",omitempty"`

	Desglose     Desglose `xml:"Desglose"`
	CuotaTotal   Amount   `xml:"CuotaTotal"`
	ImporteTotal Amount   `xml:"ImporteTotal"`

	Encadenamiento           Encadenamiento     `xml:"Encadenamiento"`
	SistemaInformatico       SistemaInformatico `xml:"SistemaInformatico"`
	FechaHoraHusoGenRegistro FechaHora          `xml:"FechaHoraHusoGenRegistro"`

	NumRegistroAcuerdoFacturacion *string `xml:"NumRegistroAcuerdoFacturacion,omitempty" json:",omitempty"`
	IdAcuerdoSistemaInformatico   *string `xml:"IdAcuerdoSistemaInformatico,omitempty" json:",omitempty"`

	TipoHuella TipoHuella `xml:"TipoHuella"`
	Huella     string     `xml:"Huella"`
}

// RegistroAnulacion cancels a previously registered invoice.
type RegistroAnulacion struct {
	XMLName xml.Name `xml:"https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroInformacion.xsd RegistroAnulacion" json:"-"`

	IDVersion         string                  `xml:"IDVersion"`
	IDFactura         IDFacturaExpedidaBaja   `xml:"IDFactura"`
	RefExterna        *string                 `xml:"RefExterna,omitempty" json:",omitempty"`
	SinRegistroPrevio *SiNo                   `xml:"SinRegistroPrevio,omitempty" json:",omitempty"`
	RechazoPrevio     *RechazoPrevioAnulacion `xml:"RechazoPrevio,omitempty" json:",omitempty"`
	GeneradoPor       *GeneradoPor            `xml:"GeneradoPor,omitempty" json:",omitempty"`
	Generador         *PersonaFisicaJuridica  `xml:"Generador,omitempty" json:",omitempty"`

	Encadenamiento           Encadenamiento     `xml:"Encadenamiento"`
	SistemaInformatico       SistemaInformatico `xml:"SistemaInformatico"`
	FechaHoraHusoGenRegistro FechaHora          `xml:"FechaHoraHusoGenRegistro"`

	TipoHuella TipoHuella `xml:"TipoHuella"`
	Huella     string     `xml:"Huella"`
}

// PersonaFisicaJuridica is a party that may be Spanish or foreign.
type PersonaFisicaJuridica struct {
	NombreRazon string  `xml:"NombreRazon"`
	NIF         *string `xml:"NIF,omitempty" json:",omitempty"`
	IDOtro      *IDOtro `xml:"IDOtro,omitempty" json:",omitempty"`
}

// Destinatarios holds the invoice recipients. The schema allows up to 1000.
type Destinatarios struct {
	IDDestinatario []PersonaFisicaJuridica `xml:"IDDestinatario" json:",omitempty"`
}

// FacturasRectificadas lists the invoices this one corrects.
type FacturasRectificadas struct {
	IDFacturaRectificada []IDFacturaExpedida `xml:"IDFacturaRectificada" json:",omitempty"`
}

// FacturasSustituidas lists the invoices this one replaces.
type FacturasSustituidas struct {
	IDFacturaSustituida []IDFacturaExpedida `xml:"IDFacturaSustituida" json:",omitempty"`
}

// DesgloseRectificacion holds the amounts being replaced in a substitutive
// corrective invoice.
type DesgloseRectificacion struct {
	BaseRectificada         Amount  `xml:"BaseRectificada"`
	CuotaRectificada        Amount  `xml:"CuotaRectificada"`
	CuotaRecargoRectificado *Amount `xml:"CuotaRecargoRectificado,omitempty" json:",omitempty"`
}

// Cabecera identifies the taxpayer the records belong to.
type Cabecera struct {
	XMLName               xml.Name                 `xml:"https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroLR.xsd Cabecera"`
	ObligadoEmision       PersonaFisicaJuridicaES  `xml:"https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroInformacion.xsd ObligadoEmision"`
	Representante         *PersonaFisicaJuridicaES `xml:"https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroInformacion.xsd Representante,omitempty" json:",omitempty"`
	RemisionVoluntaria    *RemisionVoluntaria      `xml:"https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroInformacion.xsd RemisionVoluntaria,omitempty" json:",omitempty"`
	RemisionRequerimiento *RemisionRequerimiento   `xml:"https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroInformacion.xsd RemisionRequerimiento,omitempty" json:",omitempty"`
}

type RemisionVoluntaria struct {
	FechaFinVeriFactu *Fecha `xml:"FechaFinVeriFactu,omitempty" json:",omitempty"`
	Incidencia        *SiNo  `xml:"Incidencia,omitempty" json:",omitempty"`
}

type RemisionRequerimiento struct {
	RefRequerimiento string `xml:"RefRequerimiento"`
	FinRequerimiento *SiNo  `xml:"FinRequerimiento,omitempty" json:",omitempty"`
}

// RegistroFactura wraps a single record. The schema requires exactly one of
// the two fields.
type RegistroFactura struct {
	RegistroAlta      *RegistroAlta      `xml:"RegistroAlta,omitempty" json:",omitempty"`
	RegistroAnulacion *RegistroAnulacion `xml:"RegistroAnulacion,omitempty" json:",omitempty"`
}

// RegFactuSistemaFacturacion is the root element sent to the AEAT. Up to 1000
// records per submission.
type RegFactuSistemaFacturacion struct {
	XMLName         xml.Name          `xml:"https://www2.agenciatributaria.gob.es/static_files/common/internet/dep/aplicaciones/es/aeat/tike/cont/ws/SuministroLR.xsd RegFactuSistemaFacturacion"`
	Cabecera        Cabecera          `xml:"Cabecera"`
	RegistroFactura []RegistroFactura `xml:"RegistroFactura" json:",omitempty"`
}
