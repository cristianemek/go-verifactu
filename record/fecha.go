package record

import (
	"encoding/json"
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

func (f Fecha) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Format())
}

func (f *Fecha) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	t, err := time.Parse(fechaFormat, s)
	if err != nil {
		return err
	}

	*f = Fecha(t)
	return nil
}

func (f FechaHora) Format() string {
	t := time.Time(f).Truncate(time.Second)
	return t.Format(time.RFC3339)
}

func (f FechaHora) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Format())
}

func (f *FechaHora) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*f = FechaHora(t)
	return nil
}
