package record

import (
	"encoding/xml"
	"time"
)

const (
	fechaFormat = "02-01-2006"
)

type Fecha time.Time

func (f Fecha) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	t := time.Time(f)
	return e.EncodeElement(t.Format(fechaFormat), start)
}

func (f Fecha) Format() string {
	t := time.Time(f)
	return t.Format(fechaFormat)
}
