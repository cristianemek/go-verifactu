package record

import (
	"encoding/xml"
	"time"
)

// RespuestaRegFactuSistemaFacturacion is the AEAT answer to a submission.
type RespuestaRegFactuSistemaFacturacion struct {
	XMLName           xml.Name           `xml:"RespuestaRegFactuSistemaFacturacion"`
	CSV               string             `xml:"CSV"`
	DatosPresentacion *DatosPresentacion `xml:"DatosPresentacion,omitempty"`
	Cabecera          CabeceraRespuesta  `xml:"Cabecera"`
	TiempoEsperaEnvio string             `xml:"TiempoEsperaEnvio"`
	EstadoEnvio       EstadoEnvio        `xml:"EstadoEnvio"`
	RespuestaLinea    []RespuestaLinea   `xml:"RespuestaLinea"`
}

// DatosPresentacion says who filed the submission and when, on the AEAT clock.
type DatosPresentacion struct {
	NIFPresentador        string    `xml:"NIFPresentador"`
	TimestampPresentacion time.Time `xml:"TimestampPresentacion"`
}

// CabeceraRespuesta echoes the header sent. Apart from Cabecera because that one
// pins the SuministroLR namespace and would fail to unmarshal here.
type CabeceraRespuesta struct {
	ObligadoEmision PersonaFisicaJuridicaES `xml:"ObligadoEmision"`
}

// RespuestaLinea is the outcome of one record, up to 1000 per answer.
type RespuestaLinea struct {
	IDFactura                IDFacturaExpedida  `xml:"IDFactura"`
	Operacion                OperacionRespuesta `xml:"Operacion"`
	RefExterna               string             `xml:"RefExterna"`
	EstadoRegistro           EstadoRegistro     `xml:"EstadoRegistro"`
	CodigoErrorRegistro      string             `xml:"CodigoErrorRegistro"`
	DescripcionErrorRegistro string             `xml:"DescripcionErrorRegistro"`
	RegistroDuplicado        *RegistroDuplicado `xml:"RegistroDuplicado"`
}

// OperacionRespuesta echoes the operation performed and the flags sent with it.
// Named apart from verifactu.Operacion, which means something else.
type OperacionRespuesta struct {
	TipoOperacion     TipoOperacion  `xml:"TipoOperacion"`
	Subsanacion       *SiNo          `xml:"Subsanacion,omitempty"`
	RechazoPrevio     *RechazoPrevio `xml:"RechazoPrevio,omitempty"`
	SinRegistroPrevio *SiNo          `xml:"SinRegistroPrevio,omitempty"`
}

// RegistroDuplicado is what the AEAT already had on file, sent only when a record
// is rejected as duplicated. Only place EstadoRegistroAlmacenado is used.
type RegistroDuplicado struct {
	IdPeticionRegistroDuplicado string                   `xml:"IdPeticionRegistroDuplicado"`
	EstadoRegistroDuplicado     EstadoRegistroAlmacenado `xml:"EstadoRegistroDuplicado"`
	CodigoErrorRegistro         string                   `xml:"CodigoErrorRegistro"`
	DescripcionErrorRegistro    string                   `xml:"DescripcionErrorRegistro"`
}
