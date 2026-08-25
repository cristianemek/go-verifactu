package record

// EstadoRegistroAlmacenado is the state of a record stored at the AEAT only inside of RegistroDuplicado.
type EstadoRegistroAlmacenado string

const (
	EstadoRegistroAlmacenadoCorrecta           EstadoRegistroAlmacenado = "Correcta"
	EstadoRegistroAlmacenadoAceptadaConErrores EstadoRegistroAlmacenado = "AceptadaConErrores"
	EstadoRegistroAlmacenadoAnulada            EstadoRegistroAlmacenado = "Anulada"
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

// EstadoEnvio it is the global state of the submission.
type EstadoEnvio string

const (
	EstadoEnvioCorrecto             EstadoEnvio = "Correcto"
	EstadoEnvioParcialmenteCorrecto EstadoEnvio = "ParcialmenteCorrecto"
	EstadoEnvioIncorrecto           EstadoEnvio = "Incorrecto"
)

// EstadoRegistro is the state of a record of the response.
type EstadoRegistro string

const (
	EstadoRegistroCorrecto           EstadoRegistro = "Correcto"
	EstadoRegistroAceptadoConErrores EstadoRegistro = "AceptadoConErrores"
	EstadoRegistroIncorrecto         EstadoRegistro = "Incorrecto"
)
