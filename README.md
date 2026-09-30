# go-verifactu

[![Go Reference](https://pkg.go.dev/badge/github.com/cristianemek/go-verifactu.svg)](https://pkg.go.dev/github.com/cristianemek/go-verifactu)
[![CI](https://github.com/cristianemek/go-verifactu/actions/workflows/ci.yml/badge.svg)](https://github.com/cristianemek/go-verifactu/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/cristianemek/go-verifactu/branch/main/graph/badge.svg)](https://codecov.io/gh/cristianemek/go-verifactu)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

VERI\*FACTU en Go: registra tus facturas, las encadena con su huella, te da la
URL del QR y las manda a Hacienda.

Viene de dos formas: como **librería** de Go, o como **servicio HTTP** si tu
software está en otro lenguaje.

Está pensado para montarlo en tu propio servidor, con systemd o con Docker, y
que hable con tus proyectos, estén en el lenguaje que estén, sin tener que
reimplementar en cada uno toda la lógica y las comprobaciones de VERI\*FACTU.
Levantas el servicio, te comunicas con él por HTTP, y él se encarga de
encadenar, remitir y registrarlo todo, y te devuelve el estado de la remisión y
la URL del QR.

No necesitas saber Go para usarlo: la librería es para quien quiera meterlo en
su software en Go, y el servicio, para quien use cualquier otro lenguaje.

Sin dependencias. Go 1.27+. En desarrollo: la API puede cambiar hasta la v1.0.0.

## La librería

```
go get github.com/cristianemek/go-verifactu
```

```go
engine.Alta(ctx, tenant, factura)   // registra y encadena
engine.Remitir(ctx, tenant)         // envía lo pendiente a la AEAT
```

Ejemplos completos en el
[godoc](https://pkg.go.dev/github.com/cristianemek/go-verifactu#pkg-examples).

## El servicio

```
verifactud -config verifactud.json
```

```
POST /v1/{nif}/alta      registra una factura
POST /v1/{nif}/anular    registra una anulación
GET  /v1/{nif}/estado    qué contestó la AEAT
GET  /v1/{nif}/conexion  prueba el certificado
GET  /healthz
```

Cómo montarlo, con systemd o con Docker: [docs/servicio.md](docs/servicio.md).

## Documentación

| | |
| --- | --- |
| [examples/](examples/) | casos reales: simplificadas, exentas, ISP, rectificativas, anulaciones |
| [docs/api.md](docs/api.md) | todos los endpoints y campos |
| [docs/openapi.yaml](docs/openapi.yaml) | para Swagger UI o para generar un cliente |
| [docs/servicio.md](docs/servicio.md) | configuración, despliegue, logs y almacenes |

## Qué no hace

Sólo la modalidad VERI\*FACTU y territorio común. Nada de firma XAdES, TicketBAI
ni factura electrónica B2B.

## Aviso

Esto no es un SIF: es una herramienta para construir uno. No lleva declaración
responsable, así que si la usas en tu software de facturación esa parte te toca a
ti, igual que revisar el código y comprobar que cumple.

Ver el [Artículo 13 del RD 1007/2023](https://www.boe.es/buscar/act.php?id=BOE-A-2023-24840#a1-5).

## Licencia

[MIT](LICENSE). Cualquier uso, incluido comercial, citando el uso y la autoría.
Sin garantía ni soporte: lo usas bajo tu responsabilidad.
