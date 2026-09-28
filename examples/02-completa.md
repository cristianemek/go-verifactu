# 02 — Factura completa (F1)

Lleva destinatario, y es obligatorio: sin `Destinatarios` la AEAT la rechaza
(error 1189).

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
    "IDDestinatario": [{ "NombreRazon": "CLIENTE SL", "NIF": "B12345674" }]
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

## Cliente sin NIF español

```json
"Destinatarios": {
  "IDDestinatario": [
    {
      "NombreRazon": "FOREIGN CORP",
      "IDOtro": { "CodigoPais": "FR", "IDType": "02", "ID": "FR12345678901" }
    }
  ]
}
```

| `IDType` | |
| --- | --- |
| `02` | NIF-IVA |
| `03` | pasaporte |
| `04` | documento oficial del país de residencia |
| `05` | certificado de residencia fiscal |
| `06` | otro documento probatorio |
| `07` | no censado |

`NIF` o `IDOtro`, nunca los dos. Hasta 1000 destinatarios.

## Factura emitida por un tercero o por el destinatario

```json
"EmitidaPorTerceroODestinatario": "T",
"Tercero": { "NombreRazon": "GESTORIA SL", "NIF": "B12345674" }
```

`"T"` para un tercero, `"D"` si la emite el propio destinatario.

## Fecha de la operación distinta a la de expedición

```json
"FechaOperacion": "05-09-2026"
```
