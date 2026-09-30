# API de verifactud

Referencia de endpoints. Los cuerpos de cada caso de facturación están en
[../examples/](../examples/); la definición para Swagger o para generar clientes,
en [openapi.yaml](openapi.yaml).

## Cómo funciona

1. `POST /alta` registra la factura y responde al momento, con la URL del QR.
2. El servicio la remite a la AEAT por su cuenta, normalmente en el mismo segundo.
3. `GET /estado` dice qué contestó.

La remisión es asíncrona a propósito: la factura queda registrada aunque la AEAT
esté caída, que es lo que exige la norma.

## Autenticación

```
Authorization: Bearer <token>
```

Un token por NIF. El NIF va en la ruta y admite minúsculas. Si el NIF no está
configurado o el token no es el suyo: `401`, sin distinguir los dos casos.

`/healthz` no pide token.

## Convenciones

- **Importes y porcentajes en enteros:** `12100` son 121,00 €, `2100` es el 21 %.
- **Fechas como `dd-mm-aaaa`:** `10-09-2026`.
- **Idempotencia:** dos altas con el mismo NIF, número de serie y fecha devuelven
  el mismo registro. Reintentar tras un timeout es seguro.
- **Campos que pone el servicio**, y que se ignoran si los mandas: `IDVersion`,
  `Encadenamiento`, `SistemaInformatico`, `FechaHoraHusoGenRegistro`,
  `TipoHuella` y `Huella`.

## Endpoints

| | |
| --- | --- |
| `POST /v1/{nif}/alta` | registra una factura. Cuerpo: `RegistroAlta` en JSON |
| `POST /v1/{nif}/anular` | registra una anulación. Cuerpo: `RegistroAnulacion` |
| `GET /v1/{nif}/estado` | devuelve un registro y lo que contestó la AEAT |
| `GET /v1/{nif}/conexion` | prueba el certificado de ese NIF sin enviar nada: `204` o `503` |
| `GET /healthz` | el proceso está vivo |

### Parámetros de `/alta` y `/anular`

| | cuándo |
| --- | --- |
| `?tras_rechazo` | la AEAT rechazó el último envío de esa factura |
| `?subsanacion` | la AEAT la tiene registrada y hay que corregir un dato |

Con `?tras_rechazo` el servicio decide solo el indicador que toca según el
historial de la factura. Ver [ejemplo 06](../examples/06-estado-y-correccion.md).

### Parámetros de `/estado`

| | |
| --- | --- |
| `serie` | obligatorio |
| `fecha` | obligatorio, `dd-mm-aaaa` |
| `op` | `alta` (por defecto) o `anulacion` |

## La respuesta

Alta, anulación y estado responden igual:

```json
{
  "entry": { "Operacion": "alta", "Alta": {}, "Secuencia": 1, "Huella": "…", "IDFactura": {}, "Correccion": false },
  "avisos": [],
  "aeat": { "estado": "Correcto", "csv": "A-3JLMAM3L8XVZD9" },
  "qr": "https://prewww2.aeat.es/wlpl/TIKE-CONT/ValidarQR?…"
}
```

- **`entry`** es el registro tal como quedó en la cadena. `Correccion` marca que
  sustituye a otro anterior con el mismo número.
- **`avisos`** son descuadres que la AEAT acepta pero marca. No son errores.
- **`aeat`** solo en `/estado`.
- **`qr`** es la URL que va dentro del QR. Las anulaciones no lo llevan.

### El bloque `aeat`

| `estado` | qué hacer |
| --- | --- |
| `Pendiente` | aún no remitida o sin respuesta: esperar |
| `Correcto` | nada |
| `AceptadoConErrores` | está registrada, pero hay que subsanarla con `?subsanacion` |
| `Incorrecto` | rechazada: corregir y mandarla con `?tras_rechazo` |

Con problema, añade `codigo` y `descripcion` de la AEAT. El `csv` solo sale si no
está rechazada; qué es y por qué hay que guardarlo, en el
[ejemplo 07](../examples/07-qr-y-csv.md).

Hay cuatro códigos cuya reacción no es la obvia —`3000`, `2007`, `2000` y
`4104`—: están en el [ejemplo 06](../examples/06-estado-y-correccion.md).

## Errores

```json
{ "error": "descripcion del problema" }
```

| código | cuándo |
| --- | --- |
| `400` | JSON mal formado, fecha inválida, el NIF emisor no es el de la ruta, o el registro no cumple las reglas de la AEAT |
| `401` | falta el token, no es el del NIF, o el NIF no está configurado |
| `404` | la factura no existe |
| `409` | dos peticiones a la vez sobre la misma cadena: reintentar |
| `422` | la AEAT rechazó el mensaje entero: hay datos mal, no reintentar igual |
| `429` | la AEAT marcó tiempo de espera; el servicio ya lo gestiona |
| `502` | fallo del lado de la AEAT |
| `503` | el certificado falta o no lo aceptan |
| `500` | fallo del servicio: mirar el log |

El `400` por validación es el más común al integrar, y el mensaje dice el campo.
Las dos reglas que explican la mitad de los casos:

- **1189:** `F1`, `F3` y `R1` a `R4` exigen `Destinatarios`.
- **1190:** `F2` y `R5` no lo admiten.

## Tipos de factura

| | |
| --- | --- |
| `F1` | completa (ordinaria) |
| `F2` | simplificada (ticket) |
| `F3` | sustituye a varias simplificadas |
| `R1` a `R5` | rectificativas |

Los tipos de rectificativa, el desglose, las exenciones y los regímenes están en
[examples/03](../examples/03-desglose.md) y
[examples/04](../examples/04-rectificativas.md). La lista completa de códigos de
error de la AEAT, en [aeat/errores.properties](aeat/errores.properties).
