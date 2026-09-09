package verifactu_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/aeat"
	"github.com/cristianemek/go-verifactu/record"
	"github.com/cristianemek/go-verifactu/store/ledger"
	"github.com/cristianemek/go-verifactu/store/memory"
)

func ExampleEngine_Alta() {
	store := memory.New()

	engine, err := verifactu.New(verifactu.Config{
		Store: store,
		Now:   func() time.Time { return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC) },
		SistemaInformatico: &record.SistemaInformatico{
			NombreRazon:                 "Empresa",
			NIF:                         record.Ptr("A12345678"),
			NombreSistemaInformatico:    "go-verifactu",
			IdSistemaInformatico:        "1",
			Version:                     "0.1",
			NumeroInstalacion:           "1",
			TipoUsoPosibleSoloVerifactu: record.SiNoNo,
			TipoUsoPosibleMultiOT:       record.SiNoNo,
			IndicadorMultiplesOT:        record.SiNoNo,
		},
	})
	if err != nil {
		panic(err)
	}

	tenant := verifactu.Tenant{NIF: "A12345678", IDSistemaInformatico: "1"}

	factura := record.RegistroAlta{
		IDFactura:            record.IDFacturaExpedida{IDEmisorFactura: "A12345678", NumSerieFactura: "1", FechaExpedicionFactura: record.Fecha(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC))},
		NombreRazonEmisor:    "Nombre",
		TipoFactura:          record.TipoFacturaCompleta,
		DescripcionOperacion: "Factura de prueba",
		Desglose: record.Desglose{
			DetalleDesglose: []record.DetalleDesglose{{
				Impuesto:                      record.Ptr(record.ImpuestoIVA),
				ClaveRegimen:                  record.Ptr(record.ClaveRegimenGeneral),
				CalificacionOperacion:         record.Ptr(record.CalificacionOperacionSujetaNoExentaSinISP),
				CuotaRepercutida:              record.Ptr(record.Amount(2100)),
				TipoImpositivo:                record.Ptr(record.Porcentaje(2100)),
				BaseImponibleOimporteNoSujeto: 10000,
			}},
		},
		CuotaTotal:   2100,
		ImporteTotal: 12100,
	}

	entry, err := engine.Alta(context.Background(), tenant, factura)
	if err != nil {
		panic(err)
	}

	fmt.Printf("secuencia: %d\nhuella: %s\n", entry.Secuencia, entry.Huella)
	// Output:
	// secuencia: 1
	// huella: 6887A241A37169B6E22DC022356A2567937F56973CFF6C9F29065DA7648FCDAD
}

func ExampleEngine_Remitir() {
	// El certificado cualificado de la FNMT viene como .p12; convertirlo una vez
	// con: openssl pkcs12 -in certificado.p12 -out certificado.pem -nodes
	// El fichero resultante lleva certificado y clave, por eso va dos veces.
	cert, err := aeat.CargarPEM("certificado.pem", "certificado.pem")
	if err != nil {
		panic(err)
	}

	cliente, err := aeat.NewClient(aeat.Config{
		Entorno:         aeat.EntornoPruebas,
		TipoCertificado: aeat.CertificadoRepresentante,
		Certificado:     cert,
	})
	if err != nil {
		panic(err)
	}

	// ledger persiste la cadena en disco: memory la perderia al reiniciar.
	store, err := ledger.New("./datos-verifactu")
	if err != nil {
		panic(err)
	}

	engine, err := verifactu.New(verifactu.Config{
		Store:     store,
		Transport: cliente,
		SistemaInformatico: &record.SistemaInformatico{
			NombreRazon:                 "Empresa",
			NIF:                         record.Ptr("A12345678"),
			NombreSistemaInformatico:    "go-verifactu",
			IdSistemaInformatico:        "1",
			Version:                     "0.1",
			NumeroInstalacion:           "1",
			TipoUsoPosibleSoloVerifactu: record.SiNoNo,
			TipoUsoPosibleMultiOT:       record.SiNoNo,
			IndicadorMultiplesOT:        record.SiNoNo,
		},
	})

	if err != nil {
		panic(err)
	}

	tenant := verifactu.Tenant{NIF: "A12345678", IDSistemaInformatico: "01"}

	// Rellenar como en el ejemplo de Alta.
	var factura record.RegistroAlta

	if avisos := factura.Avisos(); len(avisos) > 0 {
		fmt.Println("La AEAT acepta el registro, pero hay avisos que conviene revisar:")
		for _, aviso := range avisos {
			fmt.Println("Aviso:", aviso)
		}
	}

	if _, err := engine.Alta(context.Background(), tenant, factura); err != nil {
		panic(err)
	}

	// Remitir envia un lote de lo pendiente. Los dos primeros casos son estados
	// normales, no fallos: no conviene registrarlos como errores.
	_, err = engine.Remitir(context.Background(), tenant)
	switch {
	case err == nil:
		fmt.Println("lote presentado")
	case errors.Is(err, verifactu.ErrSinPendientes):
		fmt.Println("nada que enviar")
	case errors.Is(err, verifactu.ErrEsperaActiva):
		fmt.Println("aun no toca: el temporizador de la AEAT no ha vencido")
	case errors.Is(err, verifactu.ErrFaultServidor):
		fmt.Println("fallo del servidor: hay que reenviar")
	case errors.Is(err, verifactu.ErrFaultCliente):
		fmt.Println("mensaje rechazado: corregir antes de reenviar, NO reintentar")
	case errors.Is(err, aeat.ErrAccesoDenegado):
		fmt.Println("el certificado falta o no fue aceptado")
	default:
		fmt.Println("error inesperado:", err)
	}
}
