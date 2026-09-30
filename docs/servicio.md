# El servicio `verifactud`

Un binario HTTP sobre la librería, para usarla desde cualquier lenguaje.

Los endpoints y sus cuerpos están en [api.md](api.md); la casuística de
facturación, en [../examples/](../examples/).

## Configuración

Copia `cmd/verifactud/verifactud.example.json` y rellena el sistema informático
y, por cada NIF, su token y su certificado:

```json
"tenants": {
  "89890001K": {
    "nombre": "EMPRESA DE PRUEBAS SL",
    "token": "cambia-este-token",
    "certificado": "/etc/verifactud/89890001K.pem",
    "tipo_certificado": "representante"
  }
}
```

**El `nombre` tiene que ser exactamente el del censo de la AEAT**, o los envíos
se rechazan con el error 4104.

Cada NIF lleva su certificado porque el certificado es de quien presenta. Si
presentas por varios clientes con el mismo, como una gestoría apoderada, se
repite la ruta en cada uno.

El certificado va en PEM. Un `.p12` de la FNMT se convierte una vez:

```
openssl pkcs12 -in certificado.p12 -out certificado.pem -nodes
```

Si OpenSSL 3 se queja del cifrado (los de la FNMT suelen usar el antiguo), añade
`-legacy`.

### El bloque `sistema`

Describe el software ante la AEAT, no a tus clientes. Va dentro de cada factura
que se remite, para que la AEAT sepa qué programa la generó y quién responde de
él: el productor, que es quien firma la declaración responsable.

| campo | qué es |
| --- | --- |
| `NombreRazon`, `NIF` | el productor del software: tú, si eres quien lo pone en marcha |
| `NombreSistemaInformatico` | el nombre del programa |
| `IdSistemaInformatico` | dos letras o cifras que el productor pone a este programa para distinguirlo de otros suyos. Con uno solo, `01` |
| `Version` | la versión del programa: se sube al actualizar |
| `NumeroInstalacion` | distingue esta instalación de otras del mismo programa |
| `TipoUsoPosibleSoloVerifactu` | `S`: este programa solo funciona en modo VERI\*FACTU |
| `TipoUsoPosibleMultiOT` | `S`: puede facturar por varios NIF |
| `IndicadorMultiplesOT` | `S` si esta instalación tiene más de un NIF en `tenants` |

**`IdSistemaInformatico` se elige una vez y no se cambia.** Además de ir a la
AEAT, el servicio lo usa para separar las cadenas: cada NIF encadena bajo ese
valor (el fichero `datos/<NIF>-01.jsonl`). Con otro, el servicio arranca con una
cadena nueva y vacía, sin avisar.

## Cuándo se envía a la AEAT

No hay endpoint para enviar. Cada alta se remite al momento, salvo que la AEAT
haya marcado un tiempo de espera; entonces lo pendiente sale junto al acabar la
espera. Por si acaso, el servicio revisa la cola cada `remision_cada`, 60 s por
defecto.

Cada envío queda en el log con su CSV, y los rechazos con su código.

## Con systemd

```
git clone https://github.com/cristianemek/go-verifactu
cd go-verifactu
CGO_ENABLED=0 go build -o verifactud ./cmd/verifactud
sudo cp verifactud /usr/local/bin/
sudo cp cmd/verifactud/verifactud.service /etc/systemd/system/
sudo systemctl enable --now verifactud
```

La unidad espera el binario en `/usr/local/bin`, la configuración y el
certificado en `/etc/verifactud`, y los datos en `/var/lib/verifactud`.

## Con Docker

Con la configuración y el certificado en `/etc/verifactud`:

```
docker compose up -d
```

Actualizar, `docker compose up -d --build`. Los logs, `docker compose logs -f`.

## El log en un fichero

`"log": "/var/log/verifactud/verifactud.log"` en la configuración. Sale en JSON,
una línea por evento, así que se consulta con `jq`. Para rotarlo:

```
sudo cp cmd/verifactud/verifactud.logrotate /etc/logrotate.d/verifactud
```

Guarda un año, por semanas y comprimido. No lo pongas dentro del directorio de
datos: ese no se borra nunca, y el log sí.

## Dónde se guardan los datos

Dos almacenes, se elige con `"store"`:

| `store` | `data` es | cuándo |
| --- | --- | --- |
| `"ledger"` (por defecto) | un **directorio** con ficheros JSONL | lo sencillo: se lee con `cat`, sin dependencias |
| `"sqlite"` | la **ruta de un fichero** `.db` | cuando la cadena crece: escribe el envío y sus líneas en una sola transacción y no carga todo en memoria |

Los dos pasan la misma batería de pruebas (`store/storetest`), así que se
comportan igual. SQLite vive en su propio módulo para que la librería siga sin
dependencias.

**El directorio de datos es todo el estado: copiarlo es la copia de seguridad.**

## Cambiar de almacén

Cambiarlo sin más arranca una cadena vacía, y el servicio no lo detecta: la
siguiente factura saldría como primer registro y la AEAT la marcaría con el
error 2007, cadena desincronizada.

Para pasar del `ledger` a SQLite hay un comando, **con el servicio parado**:

```
verifactud migrar -config /etc/verifactud/verifactud.json -a /var/lib/verifactud/verifactu.db
```

Copia las cadenas y los envíos de todos los NIF de la configuración tal cual: no
recalcula huellas, porque tienen que seguir cuadrando con lo que ya tiene la
AEAT. Al terminar comprueba que la cadena nueva es válida, que lo pendiente
coincide y que el último envío es el mismo. Si algo no cuadra, para y lo dice; el
origen no se toca en ningún caso.

Después, `"store": "sqlite"` y `"data"` con la ruta del `.db`, y a arrancar. **No
borres el directorio del `ledger`:** es la copia de seguridad hasta que SQLite
lleve semanas funcionando.
