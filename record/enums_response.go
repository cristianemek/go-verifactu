package record

// EstadoRegistroSF is the state of a record stored at the AEAT.
type EstadoRegistroSF string

const (
	EstadoRegistroSFCorrecta           EstadoRegistroSF = "Correcta"
	EstadoRegistroSFAceptadaConErrores EstadoRegistroSF = "AceptadaConErrores"
	EstadoRegistroSFAnulada            EstadoRegistroSF = "Anulada"
)

// TipoOperacion is the operation the AEAT reports having performed.
type TipoOperacion string

const (
	TipoOperacionAlta      TipoOperacion = "Alta"
	TipoOperacionAnulacion TipoOperacion = "Anulacion"
)

// TipoPeriodo is the month of a query period.
type TipoPeriodo string

const (
	TipoPeriodoEnero      TipoPeriodo = "01"
	TipoPeriodoFebrero    TipoPeriodo = "02"
	TipoPeriodoMarzo      TipoPeriodo = "03"
	TipoPeriodoAbril      TipoPeriodo = "04"
	TipoPeriodoMayo       TipoPeriodo = "05"
	TipoPeriodoJunio      TipoPeriodo = "06"
	TipoPeriodoJulio      TipoPeriodo = "07"
	TipoPeriodoAgosto     TipoPeriodo = "08"
	TipoPeriodoSeptiembre TipoPeriodo = "09"
	TipoPeriodoOctubre    TipoPeriodo = "10"
	TipoPeriodoNoviembre  TipoPeriodo = "11"
	TipoPeriodoDiciembre  TipoPeriodo = "12"
)

// IndicadorRepresentante marks a query made by the obligor's representative.
// Only "S" is valid.
type IndicadorRepresentante string

const (
	IndicadorRepresentanteSi IndicadorRepresentante = "S"
)
