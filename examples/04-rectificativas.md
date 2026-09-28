# 04 — Rectificativas

Una rectificativa es un alta normal con tres campos más. Dos decisiones:

- **El tipo** (`R1` a `R5`) dice **por qué** rectificas.
- **`TipoRectificativa`** dice **cómo**: `"S"` por sustitución (total) o `"I"` por
  diferencias (parcial).

| tipo | motivo |
| --- | --- |
| `R1` | error fundado en derecho y art. 80.1 y 80.2: devoluciones, descuentos posteriores, operaciones que quedan sin efecto |
| `R2` | concurso de acreedores, art. 80.3 |
| `R3` | deuda incobrable, art. 80.4 |
| `R4` | el resto de casos |
| `R5` | rectifica una **simplificada**: la única sin `Destinatarios` |

## Parcial, por diferencias («I»)

El desglose lleva **solo lo que cambia**. Un descuento de 100 € sobre una factura
de 1.000 € va en negativo:

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
      { "IDEmisorFactura": "89890001K", "NumSerieFactura": "F-2026-0001", "FechaExpedicionFactura": "10-09-2026" }
    ]
  },
  "DescripcionOperacion": "Descuento sobre la factura F-2026-0001",
  "Destinatarios": {
    "IDDestinatario": [{ "NombreRazon": "CLIENTE SL", "NIF": "B12345674" }]
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

Con `"I"`, la AEAT **no comprueba** que base por tipo dé la cuota. Tampoco lo
comprueba en las `R2` y `R3`.

## Total, por sustitución («S»)

El desglose lleva los importes **correctos completos**, como si la factura se
volviera a emitir, y `ImporteRectificacion` los de la factura que sustituye.

Cambia esto respecto al ejemplo anterior:

```json
"TipoRectificativa": "S",
"Desglose": { "DetalleDesglose": [
  { "Impuesto": "01", "ClaveRegimen": "01", "CalificacionOperacion": "S1", "TipoImpositivo": 2100, "BaseImponibleOimporteNoSujeto": 90000, "CuotaRepercutida": 18900 }
]},
"CuotaTotal": 18900,
"ImporteTotal": 108900,
"ImporteRectificacion": { "BaseRectificada": 100000, "CuotaRectificada": 21000 }
```

`ImporteRectificacion` solo se admite con `"S"`, y con `"S"` es obligatorio.

## Anular por completo una factura

Una rectificativa que deja la operación a cero: por diferencias, con el desglose
en negativo por el importe entero.

```json
"TipoRectificativa": "I",
"Desglose": { "DetalleDesglose": [
  { "Impuesto": "01", "ClaveRegimen": "01", "CalificacionOperacion": "S1", "TipoImpositivo": 2100, "BaseImponibleOimporteNoSujeto": -100000, "CuotaRepercutida": -21000 }
]},
"CuotaTotal": -21000,
"ImporteTotal": -121000
```

No confundir con la [anulación](05-anulacion.md): eso es otra cosa, para facturas
que no debieron existir.

## Rectificar varias facturas a la vez

`IDFacturaRectificada` es una lista:

```json
"FacturasRectificadas": { "IDFacturaRectificada": [
  { "IDEmisorFactura": "89890001K", "NumSerieFactura": "F-2026-0001", "FechaExpedicionFactura": "10-09-2026" },
  { "IDEmisorFactura": "89890001K", "NumSerieFactura": "F-2026-0002", "FechaExpedicionFactura": "11-09-2026" }
]}
```

## Sustituir simplificadas por una completa (F3)

No es una rectificativa: es una factura nueva que sustituye a varios tickets.

```json
"TipoFactura": "F3",
"FacturasSustituidas": { "IDFacturaSustituida": [
  { "IDEmisorFactura": "89890001K", "NumSerieFactura": "T-2026-0001", "FechaExpedicionFactura": "10-09-2026" }
]}
```
