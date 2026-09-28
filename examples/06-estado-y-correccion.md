# 06 — Estado y corrección

## Qué contestó la AEAT

```http
GET /v1/89890001K/estado?serie=T-2026-0001&fecha=10-09-2026
```

`op=anulacion` para consultar la baja en lugar del alta.

```json
"aeat": { "estado": "Correcto", "csv": "A-3JLMAM3L8XVZD9" }
```

Los cuatro estados posibles y qué significa cada uno están en
[../docs/api.md](../docs/api.md).

Cuando hay problema, el bloque lo dice:

```json
"aeat": {
  "estado": "Incorrecto",
  "codigo": "1189",
  "descripcion": "Es obligatorio que se informe del bloque Destinatarios..."
}
```

## Corregir

El mismo `POST /alta`, con el **mismo número de serie y fecha**, los datos
completos y correctos, y un parámetro:

| | cuándo |
| --- | --- |
| `POST /v1/{nif}/alta?tras_rechazo` | estaba `Incorrecto` |
| `POST /v1/{nif}/alta?subsanacion` | estaba `Correcto` o `AceptadoConErrores` |

La respuesta trae `"Correccion": true`. Sin el parámetro, el alta sería
idempotente y te devolvería la misma factura de antes sin cambiar nada.

## Códigos que no siguen la regla

| código | llega como | qué hacer |
| --- | --- | --- |
| `3000` | `Incorrecto` | duplicado, ya consta: **nada** |
| `2007` | `AceptadoConErrores` | cadena desincronizada: avisar a quien mantiene el servicio |
| `2000` | `AceptadoConErrores` | la huella no cuadra: fallo del software |
| `4104` | se queda `Pendiente` | el nombre no coincide con el censo: revisar la configuración |

Y dos errores admisibles que la AEAT **exime** de subsanar: falta de
`ClaveRegimen` con IPSI, y fecha de generación adelantada respecto a su reloj.
