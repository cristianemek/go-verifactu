package aeat

import (
	"encoding/xml"

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

func serializarSobre(r record.RegFactuSistemaFacturacion) ([]byte, error) {
	s := sobre{}
	s.Body.Peticion = r
	x, err := xml.Marshal(s)

	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), x...), nil
}
