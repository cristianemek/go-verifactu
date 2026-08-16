package record

import (
	"encoding/xml"
	"time"
)

const (
	fechaFormat = "02-01-2006"
)

type Fecha time.Time

type FechaHora time.Time

func (f Fecha) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(f.Format(), start)
}

func (f Fecha) Format() string {
	t := time.Time(f)
	return t.Format(fechaFormat)
}

func (f FechaHora) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(f.Format(), start)
}

func (f FechaHora) Format() string {
	t := time.Time(f).Truncate(time.Second)
	return t.Format(time.RFC3339)
}
