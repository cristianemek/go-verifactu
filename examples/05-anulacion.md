# 05 — Anulación

Para una factura que **no debió existir**: emitida por error, sin venta detrás.
Si hubo operación y cambian los importes, es una
[rectificativa](04-rectificativas.md).

```http
POST /v1/89890001K/anular
```

```json
{
  "IDFactura": {
    "IDEmisorFacturaAnulada": "89890001K",
    "NumSerieFacturaAnulada": "T-2026-0001",
    "FechaExpedicionFacturaAnulada": "10-09-2026"
  }
}
```

`201`, y sin `qr`: una anulación no lo lleva. El alta original no se borra:
quedan los dos registros en la cadena.

## Si la AEAT no tiene esa factura

Cuando su alta fue rechazada, o se emitió antes de usar VERI\*FACTU:

```json
{
  "IDFactura": {
    "IDEmisorFacturaAnulada": "89890001K",
    "NumSerieFacturaAnulada": "ANTIGUA-0001",
    "FechaExpedicionFacturaAnulada": "10-09-2026"
  },
  "SinRegistroPrevio": "S"
}
```

Con la factura sí registrada, esto lo rechaza la AEAT. Mira antes el
[estado](06-estado-y-correccion.md).

## Si la anulación la genera otro

```json
"GeneradoPor": "T",
"Generador": { "NombreRazon": "GESTORIA SL", "NIF": "B12345674" }
```

`"E"` el expedidor, `"D"` el destinatario, `"T"` un tercero.
