# API de verifactud

`verifactud` registra facturas en VERI*FACTU y las remite a la AEAT. Se habla
con él por HTTP y JSON, desde cualquier lenguaje.

Este documento es la referencia de los cinco endpoints, con un ejemplo de
`curl` y la respuesta real de cada uno. La definición OpenAPI está en
[openapi.yaml](openapi.yaml): se puede abrir en Swagger UI o generar un cliente
con ella.

## Cómo funciona

1. Registras la factura con `POST /v1/{nif}/alta`. Responde al momento, con la
   factura ya encadenada y la URL del QR que tienes que imprimir.
2. El servicio la remite a la AEAT por su cuenta, normalmente en el mismo
   segundo. No hay que pedirlo.
3. Consultas el resultado con `GET /v1/{nif}/estado`.

La remisión es asíncrona a propósito: tu factura queda registrada aunque la
AEAT esté caída, que es lo que exige la norma.

## Autenticación

Todos los endpoints menos `/healthz` piden un token, y cada NIF tiene el suyo:

```
Authorization: Bearer <token>
```

El NIF va en la ruta y puede ir en minúsculas. Si el NIF no está configurado o
el token no es el suyo, la respuesta es `401`, sin distinguir entre los dos
casos.

## Convenciones

**Importes: enteros, en céntimos.** `12100` son 121,00 €. No se usan decimales
en ningún campo de importe, para que no haya errores de redondeo.

**Porcentajes: igual.** `2100` es el 21 %.

**Fechas: `dd-mm-aaaa`.** `10-09-2026`. Una fecha en otro formato devuelve
`400`.

**Campos que no tienes que mandar.** El servicio rellena, y sobrescribe si los
mandas: `IDVersion`, `Encadenamiento`, `SistemaInformatico`,
`FechaHoraHusoGenRegistro`, `TipoHuella` y `Huella`. Son la cadena de huellas y
la identificación del software: los pone el servicio porque de eso responde él.

**Repetir un alta es seguro.** Dos altas con el mismo NIF, número de serie y
fecha de expedición devuelven el mismo registro, sin duplicarlo. Si una
petición se corta por un timeout, se puede reintentar tal cual.

---

## POST /v1/{nif}/alta

Registra una factura. El cuerpo es el registro de alta en JSON.

### Factura simplificada (F2), un ticket

Una `F2` **no puede llevar destinatarios**. La AEAT la rechazaría con el
error 1190.

```bash
curl -X POST http://localhost:9009/v1/89890001K/alta \
  -H "Authorization: Bearer TU_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "IDFactura": {
      "IDEmisorFactura": "89890001K",
      "NumSerieFactura": "T-2026-0001",
      "FechaExpedicionFactura": "10-09-2026"
    },
    "NombreRazonEmisor": "EMPRESA DE PRUEBAS SL",
    "TipoFactura": "F2",
    "DescripcionOperacion": "Consumición en barra",
    "Desglose": {
      "DetalleDesglose": [
        {
          "Impuesto": "01",
          "ClaveRegimen": "01",
          "CalificacionOperacion": "S1",
          "TipoImpositivo": 2100,
          "BaseImponibleOimporteNoSujeto": 10000,
          "CuotaRepercutida": 2100
        }
      ]
    },
    "CuotaTotal": 2100,
    "ImporteTotal": 12100
  }'
```

Respuesta `201`:

```json
{
  "entry": {
    "Operacion": "alta",
    "Alta": { "...": "el registro completo, ya encadenado" },
    "Secuencia": 7,
    "Huella": "CD7D37A146990D6F8247903BA02C4CF64BF8DCD6B8B198BBC7092C1FA1963DB4",
    "IDFactura": {
      "NIF": "89890001K",
      "NumSerie": "T-2026-0001",
      "Fecha": "10-09-2026"
    },
    "Correccion": false
  },
  "avisos": [],
  "qr": "https://prewww2.aeat.es/wlpl/TIKE-CONT/ValidarQR?fecha=10-09-2026&importe=121.00&nif=89890001K&numserie=T-2026-0001"
}
```

`Secuencia` es la posición en la cadena de ese NIF y `Huella` es la huella
encadenada. Los dos son del registro, no tuyos: no hay que guardarlos para
nada, pero sirven para auditar.

`avisos` son descuadres que la AEAT acepta pero conviene mirar, por ejemplo que
la suma del desglose no cuadre con el importe total. **No son errores:** la
factura queda registrada igual.

### Factura completa (F1), con destinatario

Una `F1` **exige destinatarios**, o la AEAT la rechaza con el error 1189. La
diferencia con el ejemplo anterior es el bloque `Destinatarios`:

```json
{
  "IDFactura": {
    "IDEmisorFactura": "89890001K",
    "NumSerieFactura": "F-2026-0001",
    "FechaExpedicionFactura": "10-09-2026"
  },
  "NombreRazonEmisor": "EMPRESA DE PRUEBAS SL",
  "TipoFactura": "F1",
  "DescripcionOperacion": "Servicios de desarrollo",
  "Destinatarios": {
    "IDDestinatario": [
      { "NombreRazon": "CLIENTE SL", "NIF": "B12345674" }
    ]
  },
  "Desglose": {
    "DetalleDesglose": [
      {
        "Impuesto": "01",
        "ClaveRegimen": "01",
        "CalificacionOperacion": "S1",
        "TipoImpositivo": 2100,
        "BaseImponibleOimporteNoSujeto": 100000,
        "CuotaRepercutida": 21000
      }
    ]
  },
  "CuotaTotal": 21000,
  "ImporteTotal": 121000
}
```

Un cliente extranjero sin NIF español va con `IDOtro` en lugar de `NIF`:

```json
{
  "NombreRazon": "FOREIGN CORP",
  "IDOtro": { "CodigoPais": "FR", "IDType": "02", "ID": "FR12345678901" }
}
```

### Factura rectificativa (R1 a R5)

Una rectificativa es un alta más, con tres campos de más y el tipo `R1` a `R5`.

**Cuál elegir.** El tipo depende del motivo por el que rectificas, no de cómo
lo rectificas:

| tipo | cuándo |
| --- | --- |
| `R1` | error fundado en derecho, y los casos del art. 80.1 y 80.2 de la Ley del IVA: devoluciones, descuentos posteriores, operaciones que quedan sin efecto |
| `R2` | el cliente entra en concurso de acreedores, art. 80.3 |
| `R3` | deuda incobrable, art. 80.4 |
| `R4` | el resto de casos |
| `R5` | rectifica una factura **simplificada** |

`R5` es la única que no lleva destinatarios, igual que la `F2` que rectifica.

Ante la duda entre `R1` y `R4`, `R1` cubre la mayoría de los casos del día a
día: una devolución, un descuento pactado después, una operación anulada. `R4`
es el cajón de sastre.

**Rectificativa o subsanación.** Si el error es un dato mal escrito y la
operación es la que es, no es una rectificativa: es una subsanación, que se
manda con `?subsanacion` (más abajo). La rectificativa es para cuando cambia la
operación.

`TipoRectificativa` dice cómo rectifica. Es **obligatorio** en las `R1` a `R5`,
y no se admite en las demás:

- **`"I"`, por diferencias.** El desglose lleva **solo la diferencia**, que
  suele ser negativa.
- **`"S"`, por sustitución.** El desglose lleva los importes **correctos
  completos**, y en `ImporteRectificacion` van los de la factura que se
  rectifica.

Ejemplo por diferencias, corrigiendo 100 € de más:

```json
{
  "IDFactura": {
    "IDEmisorFactura": "89890001K",
    "NumSerieFactura": "R-2026-0001",
    "FechaExpedicionFactura": "15-09-2026"
  },
  "NombreRazonEmisor": "EMPRESA DE PRUEBAS SL",
  "TipoFactura": "R1",
  "TipoRectificativa": "I",
  "FacturasRectificadas": {
    "IDFacturaRectificada": [
      {
        "IDEmisorFactura": "89890001K",
        "NumSerieFactura": "F-2026-0001",
        "FechaExpedicionFactura": "10-09-2026"
      }
    ]
  },
  "DescripcionOperacion": "Rectificacion de la factura F-2026-0001",
  "Destinatarios": {
    "IDDestinatario": [
      { "NombreRazon": "CLIENTE SL", "NIF": "B12345674" }
    ]
  },
  "Desglose": {
    "DetalleDesglose": [
      {
        "Impuesto": "01",
        "ClaveRegimen": "01",
        "CalificacionOperacion": "S1",
        "TipoImpositivo": 2100,
        "BaseImponibleOimporteNoSujeto": -10000,
        "CuotaRepercutida": -2100
      }
    ]
  },
  "CuotaTotal": -2100,
  "ImporteTotal": -12100
}
```

Por sustitución, el mismo registro con `"TipoRectificativa": "S"`, el desglose
con los importes correctos y, además:

```json
"ImporteRectificacion": {
  "BaseRectificada": 100000,
  "CuotaRectificada": 21000
}
```

Las rectificativas `R1` a `R4` llevan destinatarios. La `R5`, que rectifica una
simplificada, no.

`FacturasRectificadas` no es obligatoria para la AEAT, pero conviene mandarla:
es lo que ata la rectificativa con la factura que corrige.

Un detalle útil de las rectificativas por diferencias: la AEAT comprueba que
base por tipo dé la cuota, con un margen de 10 €, y que base y cuota tengan el
mismo signo. **Esa comprobación no se aplica** cuando `TipoRectificativa` es
`"I"` ni en las `R2` y `R3`, que es donde los importes no cuadran por su
naturaleza.

### Corregir una factura rechazada

Si la AEAT rechazó una factura, **no se reenvía la misma**. Se manda otra vez
corregida, con `?tras_rechazo`:

```bash
curl -X POST "http://localhost:9009/v1/89890001K/alta?tras_rechazo" \
  -H "Authorization: Bearer TU_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "...": "el registro corregido, con el mismo numero de serie" }'
```

El servicio marca el registro como subsanación y lo encadena de nuevo. Sin ese
parámetro, el alta sería idempotente y te devolvería la factura rechazada tal
cual.

Hay un segundo parámetro, `?subsanacion`, para corregir una factura que la AEAT
**sí tiene registrada**, aceptada o aceptada con errores.

La diferencia está en si la factura llegó a entrar en la AEAT:

| parámetro | cuándo | qué manda |
| --- | --- | --- |
| `?tras_rechazo` | la AEAT **rechazó** el último envío de esa factura | `Subsanacion: S` y `RechazoPrevio: X` o `S` |
| `?subsanacion` | la AEAT la **tiene registrada** y hay que corregir un dato | `Subsanacion: S` |

Con `?tras_rechazo` no tienes que distinguir qué se rechazó. El servicio mira
el historial de esa factura: si ninguna versión llegó a aceptarse, manda `X`;
si alguna se aceptó y lo rechazado fue una subsanación posterior, manda `S`.
Son las dos operativas que la AEAT distingue, y el servicio ya sabe cuál toca.

Si te equivocas de parámetro, la AEAT lo rechaza: con `?subsanacion` sobre algo
que no consta, y con `?tras_rechazo` sobre algo que sí consta.

Dos cosas que la AEAT deja claras sobre la subsanación:

- El registro corregido va **con el mismo número de serie y la misma fecha**, y
  con **todos los datos completos y correctos**, no solo el campo cambiado.
- El registro original no se modifica ni desaparece: queda en la cadena, y
  encima queda el nuevo.

---

## POST /v1/{nif}/anular

Anula una factura ya registrada. El cuerpo es el registro de anulación.

Antes de usarlo: **anular no es lo habitual**. Para deshacer una factura, lo
normal es emitir una rectificativa, que deja los dos documentos en la cadena.

La AEAT reserva la anulación para las operaciones incorrectas que **no admiten
factura rectificativa**, y pone un ejemplo claro: **una factura emitida por
error cuando no ha habido una venta de verdad**. Si hubo operación y lo que
cambia es el importe, el destinatario o las condiciones, es una rectificativa.

```bash
curl -X POST http://localhost:9009/v1/89890001K/anular \
  -H "Authorization: Bearer TU_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "IDFactura": {
      "IDEmisorFacturaAnulada": "89890001K",
      "NumSerieFacturaAnulada": "F-2026-0001",
      "FechaExpedicionFacturaAnulada": "10-09-2026"
    }
  }'
```

Responde `201` con la misma forma que el alta, pero sin `qr`: una anulación no
lleva QR.

La anulación normal exige que la factura **conste en la AEAT**. Si no consta,
la rechaza.

### Anular algo que la AEAT no tiene

Para ese caso está `"SinRegistroPrevio": "S"`, un campo tuyo que el servicio no
toca. La AEAT contempla dos situaciones:

- La factura se emitió cuando el sistema **todavía no era VERI\*FACTU**, así que
  nunca se remitió.
- La AEAT **rechazó el alta**, y en lugar de corregirla se decide que esa
  factura hay que anularla.

**Cómo saber si consta o no:** con `GET /estado`. El bloque `aeat` te lo dice
sin ambigüedad. `Correcto` o `AceptadoConErrores` quiere decir que la AEAT la
tiene, así que la anulación va normal. `Incorrecto` quiere decir que fue
rechazada y **no consta**, así que toca `SinRegistroPrevio`. Mientras esté
`Pendiente`, espera: todavía no se sabe.

Esto es justo lo que el servicio guarda por ti. La AEAT no tiene forma de
consultar una factura concreta a posteriori en todos los casos, así que el
registro local es la fuente.

Y al revés: usar `SinRegistroPrevio: "S"` sobre una factura que **sí** consta
también lo rechaza la AEAT. No es un campo de "por si acaso".

---

## GET /v1/{nif}/estado

Devuelve una factura registrada y lo que contestó la AEAT.

```bash
curl "http://localhost:9009/v1/89890001K/estado?serie=T-2026-0001&fecha=10-09-2026" \
  -H "Authorization: Bearer TU_TOKEN"
```

| parámetro | obligatorio | qué es |
| --- | --- | --- |
| `serie` | sí | el número de serie de la factura |
| `fecha` | sí | la fecha de expedición, `dd-mm-aaaa` |
| `op` | no | `alta` (por defecto) o `anulacion` |

Respuesta `200`:

```json
{
  "entry": { "...": "el registro" },
  "avisos": [],
  "aeat": {
    "estado": "Correcto",
    "csv": "A-3JLMAM3L8XVZD9"
  },
  "qr": "https://prewww2.aeat.es/wlpl/TIKE-CONT/ValidarQR?..."
}
```

El bloque `aeat` tiene cuatro estados:

| estado | qué significa | qué hacer |
| --- | --- | --- |
| `Pendiente` | aún no se ha remitido, o se remitió y no hay respuesta | esperar y volver a preguntar |
| `Correcto` | aceptada | nada |
| `AceptadoConErrores` | **registrada**, pero con errores admisibles | corregirla y mandarla con `?subsanacion` |
| `Incorrecto` | **rechazada**, no consta en la AEAT | corregirla y mandarla con `?tras_rechazo` |

Cuando hay rechazo, `codigo` y `descripcion` traen el error de la AEAT:

```json
"aeat": {
  "estado": "Incorrecto",
  "codigo": "1189",
  "descripcion": "Es obligatorio que se informe del bloque Destinatarios..."
}
```

Una factura registrada hace un segundo suele salir como `Pendiente`: la
remisión es asíncrona. Si sigue `Pendiente` pasados unos minutos, mira el log
del servicio.

### Qué hacer con «AceptadoConErrores»

No es un aviso que se pueda ignorar. La factura **queda registrada**, pero la
AEAT dice que hay que subsanarla: mandar el registro otra vez, con el mismo
número de serie y fecha, ya con todos los datos correctos, usando
`?subsanacion`.

Los casos típicos son un NIF de destinatario correcto pero no censado, una
huella que no le cuadra a la AEAT, o un total que no cuadra con el desglose.

Hay dos excepciones que la AEAT dice expresamente que **no** hace falta
subsanar:

- falta `ClaveRegimen` con impuesto IPSI;
- la fecha de generación del registro va adelantada respecto al reloj de la
  AEAT.

Fuera de esas dos, mira la `descripcion` y subsana.

### El CSV

El CSV, «código seguro de verificación», es el justificante de que la remisión
llegó. Son 16 caracteres, y la AEAT dice algo importante: **no se puede
recuperar después**, no hay forma de volver a pedirlo.

Por eso el servicio lo guarda en el momento del envío y te lo devuelve aquí.
Guárdalo tú también junto a la factura, en tu base de datos.

Dos detalles: es **del envío, no de la factura**, así que varias facturas
remitidas juntas comparten CSV, y no se emite si la AEAT rechaza el envío
entero. Por eso no sale en las facturas rechazadas.

La AEAT no obliga a imprimirlo en la factura ni a enseñárselo al cliente: es
tuyo, para acreditar la remisión si alguna vez hace falta.

---

## GET /v1/{nif}/conexion

Prueba el certificado de ese NIF contra la AEAT, sin enviar ninguna factura.
Útil al configurar un cliente nuevo o para vigilar que el certificado no ha
caducado.

```bash
curl -i "http://localhost:9009/v1/89890001K/conexion" \
  -H "Authorization: Bearer TU_TOKEN"
```

`204` si la AEAT acepta el certificado. `503` si lo rechaza o si falta.

---

## GET /healthz

Sin token. `200` si el proceso está vivo. No comprueba la AEAT.

---

## El QR

Cada factura tiene que llevar un QR impreso. El servicio devuelve **la URL** que
va dentro, en el campo `qr` de `/alta` y de `/estado`. La imagen la pintas tú,
con la librería de QR de tu lenguaje.

La AEAT pide, además:

- Código QR según ISO/IEC 18004:2015, nivel de corrección de errores **M**.
- Entre 30x30 mm y 40x40 mm impreso.
- Al menos 2 mm en blanco alrededor, mejor 6.
- En la primera página, con el texto **«QR tributario:»** encima y
  **«VERI*FACTU»** debajo.

La URL apunta al entorno de pruebas o al de producción según la configuración
del servicio, así que no hay que tocar nada al pasar a producción.

---

## Errores

Todos tienen la misma forma:

```json
{ "error": "descripcion del problema" }
```

| código | cuándo | qué hacer |
| --- | --- | --- |
| `400` | JSON mal formado, fecha inválida, o el registro no cumple las reglas de la AEAT | corregir la petición; el mensaje dice qué campo |
| `401` | falta el token, no es el del NIF, o el NIF no está configurado | revisar la configuración |
| `404` | la factura no existe | revisar serie, fecha y `op` |
| `409` | dos peticiones a la vez sobre la misma cadena | reintentar |
| `422` | la AEAT rechazó el mensaje entero | no reintentar tal cual: hay algo mal en los datos |
| `429` | la AEAT marcó tiempo de espera | esperar; el servicio ya lo gestiona solo |
| `502` | fallo del lado de la AEAT | reintentar más tarde |
| `503` | el certificado falta o la AEAT no lo acepta | revisar el certificado |
| `500` | fallo del servicio | mirar el log |

El `400` por validación es el más común al integrar. Estas dos reglas explican
la mitad de los casos:

- **1189:** las facturas `F1`, `F3`, `R1`, `R2`, `R3` y `R4` exigen
  `Destinatarios`.
- **1190:** las `F2` y `R5` no lo admiten.

---

## Tablas de códigos

### TipoFactura

| valor | qué es |
| --- | --- |
| `F1` | factura completa (ordinaria) |
| `F2` | factura simplificada (ticket) |
| `F3` | factura emitida en sustitución de simplificadas |
| `R1` | rectificativa por error fundado en derecho, art. 80.1 y 80.2 |
| `R2` | rectificativa por concurso de acreedores, art. 80.3 |
| `R3` | rectificativa por deuda incobrable, art. 80.4 |
| `R4` | rectificativa por el resto de casos |
| `R5` | rectificativa de una factura simplificada |

### TipoRectificativa

| valor | qué es |
| --- | --- |
| `S` | por sustitución: el desglose lleva los importes correctos |
| `I` | por diferencias: el desglose lleva solo la diferencia |

### Impuesto

| valor | qué es |
| --- | --- |
| `01` | IVA |
| `02` | IPSI (Ceuta y Melilla) |
| `03` | IGIC (Canarias) |
| `05` | otros |

### CalificacionOperacion

| valor | qué es |
| --- | --- |
| `S1` | sujeta y no exenta, sin inversión del sujeto pasivo |
| `S2` | sujeta y no exenta, con inversión del sujeto pasivo |
| `N1` | no sujeta, art. 7, 14 y otros |
| `N2` | no sujeta por las reglas de localización |

Una línea del desglose lleva `CalificacionOperacion` **o** `OperacionExenta`,
nunca las dos.

### OperacionExenta

De `E1` a `E6`, según el artículo de la Ley del IVA que declara la exención, y
`E7` y `E8` para los casos restantes. El detalle está en el documento de diseño
de registro de la AEAT.

### ClaveRegimen

`01` es el régimen general, que es el de casi todas las facturas. El resto
—`02` a `11`, `14`, `15` y `17` a `21`— son regímenes especiales:
exportaciones, bienes usados, agencias de viaje, recargo de equivalencia,
criterio de caja y demás. La lista con su significado está en el documento de
diseño de registro de la AEAT.

---

## Notas para integrar

**Guarda `NumSerieFactura` y la fecha.** Son la clave para consultar el estado
luego. El servicio no da un identificador propio.

**Consulta el estado en diferido.** Lo razonable es preguntar por las facturas
del día un rato después, no bloquear la venta esperando a la AEAT.

**Los rechazos hay que atenderlos.** Una factura rechazada no cumple hasta que
se manda corregida. Merece la pena avisar a quien factura, no solo apuntarlo en
un log.

**Un NIF, una cadena.** Las facturas de un NIF van encadenadas por orden de
registro. Por eso el directorio de datos del servicio es todo el estado: hacerle
una copia es la copia de seguridad, y borrarlo rompe la cadena.
