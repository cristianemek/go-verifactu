# go-verifactu

[![Go Reference](https://pkg.go.dev/badge/github.com/cristianemek/go-verifactu.svg)](https://pkg.go.dev/github.com/cristianemek/go-verifactu)
[![CI](https://github.com/cristianemek/go-verifactu/actions/workflows/ci.yml/badge.svg)](https://github.com/cristianemek/go-verifactu/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/cristianemek/go-verifactu/branch/main/graph/badge.svg)](https://codecov.io/gh/cristianemek/go-verifactu)
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

Tres cosas antes de empezar: los importes van en enteros —`record.Amount(2100)`
son 21,00 €—, `Alta` es idempotente, así que reintentar tras un timeout es
seguro, y `factura.Avisos()` comprueba los descuadres que la AEAT marca pero no
rechaza.

## Qué cubre

Sólo la modalidad VERI*FACTU y territorio común. No hace firma XAdES, ni
TicketBAI, ni factura electrónica B2B.

El certificado se carga en PEM. Si tienes un `.p12` de la FNMT, se convierte una
vez:

```
openssl pkcs12 -in certificado.p12 -out certificado.pem -nodes
```

Si OpenSSL 3 se queja del cifrado del `.p12` (los de la FNMT suelen usar el
antiguo), añade `-legacy` al comando.

## Aviso

Esto no es un SIF, es una herramienta para construir uno. No lleva declaración
responsable: si la usas en tu software de facturación, esa parte te toca a ti,
igual que revisar el código y comprobar que cumple.

Ver el [Artículo 13 del RD 1007/2023](https://www.boe.es/buscar/act.php?id=BOE-A-2023-24840#a1-5).

## Licencia

[MIT](LICENSE). Se puede utilizar este proyecto para cualquier uso, incluso comercial, siempre que se haga referencia al uso y autoría del mismo. No se ofrece ninguna garantía de funcionamiento ni soporte. El uso de este proyecto es bajo la responsabilidad del usuario.

## Servicio

`cmd/verifactud` es un binario HTTP sobre la librería, para usarla desde
cualquier lenguaje. Un fichero de configuración
(`cmd/verifactud/verifactud.example.json`) con el sistema informático y, por
cada NIF, su token y su certificado.

Cada NIF lleva el suyo porque el certificado es de quien presenta. Si presentas
por varios clientes con el mismo certificado, como una gestoría apoderada, se
repite la ruta en cada uno.

La referencia con ejemplos está en [docs/api.md](docs/api.md), y la definición
OpenAPI, para Swagger UI o para generar un cliente, en
[docs/openapi.yaml](docs/openapi.yaml).

```
POST /v1/{nif}/alta      registra una factura (cuerpo: RegistroAlta en JSON)
POST /v1/{nif}/anular    registra una anulación
GET  /v1/{nif}/estado    devuelve un registro por serie y fecha
GET  /v1/{nif}/conexion  prueba el certificado contra la AEAT sin enviar nada
GET  /healthz
```

Todas menos `/healthz` piden `Authorization: Bearer <token>`. Las tres primeras
responden `{"entry": ..., "avisos": [...]}`, con `"qr"`, la URL que va dentro
del QR de la factura, y `/estado` añade lo que contestó la AEAT:

```json
"aeat": {"estado": "Pendiente"}
"aeat": {"estado": "Correcto", "csv": "A-3JLMAM3L8XVZD9"}
"aeat": {"estado": "Incorrecto", "codigo": "1189", "descripcion": "..."}
```

Una factura rechazada se corrige enviándola otra vez con `?tras_rechazo`.

El envío a la AEAT no tiene endpoint: cada alta se remite al momento, salvo que
la AEAT haya marcado un tiempo de espera; entonces lo pendiente sale junto al
acabar la espera. Por si acaso, el servicio revisa la cola cada `remision_cada`
(60 s por defecto). Cada envío queda en el log con su CSV y los rechazos.

Con systemd:

```
go install github.com/cristianemek/go-verifactu/cmd/verifactud@latest
sudo cp cmd/verifactud/verifactud.service /etc/systemd/system/
sudo systemctl enable --now verifactud
```

La unidad espera el binario en `/usr/local/bin`, la configuración y el
certificado en `/etc/verifactud`, y los datos en `/var/lib/verifactud`.

Para tener el log en un fichero, `"log": "/var/log/verifactud/verifactud.log"`
en la configuración. Sale en JSON, una línea por evento. Para rotarlo:

```
sudo cp cmd/verifactud/verifactud.logrotate /etc/logrotate.d/verifactud
```

Guarda un año, por semanas y comprimido. No lo pongas dentro del directorio de
datos: ese no se borra nunca, y el log sí.

Con Docker, con la configuración y el certificado en `/etc/verifactud`:

```
docker compose up -d
```

Para actualizar, `docker compose up -d --build`; los logs, `docker compose logs -f`.

El directorio de datos es todo el estado: copiarlo es la copia de seguridad.
