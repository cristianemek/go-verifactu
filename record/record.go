package record

func ptr[T any](v T) *T {
	return &v
}

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
