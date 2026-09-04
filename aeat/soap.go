package aeat

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/record"
)

const (
	nsSOAP = "http://schemas.xmlsoap.org/soap/envelope/"
)

type sobre struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Header  struct{} `xml:"http://schemas.xmlsoap.org/soap/envelope/ Header"`
	Body    struct {
		Peticion record.RegFactuSistemaFacturacion
	} `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

type sobreRespuesta struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    struct {
		Respuesta *record.RespuestaRegFactuSistemaFacturacion
		Fault     *SoapFault `xml:"Fault"`
	} `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

type SoapFault struct {
	Code    string `xml:"faultcode"`
	Message string `xml:"faultstring"`
	Detail  string `xml:"detail"`
}

func serializarSobre(r record.RegFactuSistemaFacturacion) ([]byte, error) {
	s := sobre{}
	s.Body.Peticion = r
	x, err := xml.Marshal(s)

	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), x...), nil
}

func (f *SoapFault) Error() string {
	return fmt.Sprintf("%s: %s", f.Code, f.Message)
}

func clasificarFault(f *SoapFault) error {
	parts := strings.Split(f.Code, ":")

	if parts[len(parts)-1] == "Server" {
		return fmt.Errorf("%w: %w", verifactu.ErrFaultServidor, f)
	}

	return fmt.Errorf("%w: %w", verifactu.ErrFaultCliente, f)
}

func parsearRespuesta(body []byte) (record.RespuestaRegFactuSistemaFacturacion, error) {
	var s sobreRespuesta
	err := xml.Unmarshal(body, &s)
	if err != nil {
		return record.RespuestaRegFactuSistemaFacturacion{}, fmt.Errorf("Body is not a valid sobre SOAP: %w", err)
	}

	if s.Body.Fault != nil {
		return record.RespuestaRegFactuSistemaFacturacion{}, clasificarFault(s.Body.Fault)
	}

	if s.Body.Respuesta != nil {
		return *s.Body.Respuesta, nil
	}

	return record.RespuestaRegFactuSistemaFacturacion{}, ErrRespuestaInesperada

}
