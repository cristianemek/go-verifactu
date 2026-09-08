# go-verifactu

[![Go Reference](https://pkg.go.dev/badge/github.com/cristianemek/go-verifactu.svg)](https://pkg.go.dev/github.com/cristianemek/go-verifactu)
[![CI](https://github.com/cristianemek/go-verifactu/actions/workflows/ci.yml/badge.svg)](https://github.com/cristianemek/go-verifactu/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/cristianemek/go-verifactu/branch/main/graph/badge.svg)](https://codecov.io/gh/cristianemek/go-verifactu)
[![Go Report Card](https://goreportcard.com/badge/github.com/cristianemek/go-verifactu)](https://goreportcard.com/report/github.com/cristianemek/go-verifactu)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Librería en Go para VERI*FACTU, el sistema de facturación de la AEAT.

Calcula la huella encadenada, monta la URL del QR y remite los registros a
Hacienda.

Sin dependencias. Go 1.22+.

En desarrollo: la API puede cambiar hasta la v1.0.0.

## Instalación

```
go get github.com/cristianemek/go-verifactu
```

## Uso

```go
store, _ := ledger.New("./datos-verifactu")

cert, _ := aeat.CargarPEM("certificado.pem", "certificado.pem")
cliente, _ := aeat.NewClient(aeat.Config{
    Entorno:         aeat.EntornoPruebas,
    TipoCertificado: aeat.CertificadoRepresentante,
    Certificado:     cert,
})

engine, _ := verifactu.New(verifactu.Config{Store: store, Transport: cliente})

engine.Alta(ctx, tenant, factura)   // registra y encadena
engine.Remitir(ctx, tenant)         // envía lo pendiente a la AEAT
```

Los ejemplos completos están en el
[godoc](https://pkg.go.dev/github.com/cristianemek/go-verifactu#pkg-examples).

Dos cosas antes de empezar: los importes van en enteros —`record.Amount(2100)`
son 21,00 €— y `Alta` es idempotente, así que reintentar tras un timeout es
seguro.

## Qué cubre

Sólo la modalidad VERI*FACTU y territorio común. No hace firma XAdES, ni
TicketBAI, ni factura electrónica B2B.

El certificado se carga en PEM. Si tienes un `.p12` de la FNMT, se convierte una
vez:

```
openssl pkcs12 -in certificado.p12 -out certificado.pem -nodes
```

## Aviso

Esto no es un SIF, es una herramienta para construir uno. No lleva declaración
responsable: si la usas en tu software de facturación, esa parte te toca a ti,
igual que revisar el código y comprobar que cumple.

Ver el [Artículo 13 del RD 1007/2023](https://www.boe.es/buscar/act.php?id=BOE-A-2023-24840#a1-5).

## Licencia

[MIT](LICENSE). Se puede utilizar este proyecto para cualquier uso, incluso comercial, siempre que se haga referencia al uso y autoría del mismo. No se ofrece ninguna garantía de funcionamiento ni soporte. El uso de este proyecto es bajo la responsabilidad del usuario.
