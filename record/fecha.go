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
