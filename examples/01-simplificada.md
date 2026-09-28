# 01 — Factura simplificada (F2)

Un ticket: sin datos del cliente. **No admite `Destinatarios`** (error 1190).

```json
{
  "IDFactura": {
    "IDEmisorFactura": "89890001K",
    "NumSerieFactura": "T-2026-0001",
    "FechaExpedicionFactura": "10-09-2026"
  },
  "NombreRazonEmisor": "EMPRESA DE PRUEBAS SL",
  "TipoFactura": "F2",
  "DescripcionOperacion": "Consumicion en barra",
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
}
```

100 € de base, 21 € de IVA, 121 € de total.

Respuesta, `201`:

```json
{
  "entry": { "Secuencia": 1, "Huella": "483DA000…", "…": "…" },
  "avisos": [],
  "qr": "https://prewww2.aeat.es/wlpl/TIKE-CONT/ValidarQR?fecha=10-09-2026&importe=121.00&nif=89890001K&numserie=T-2026-0001"
}
```

El resto de ejemplos solo muestra el cuerpo: la respuesta tiene siempre esta
forma.

## Simplificada sin IVA

Cambia el desglose: sin `TipoImpositivo` ni `CuotaRepercutida`, y el motivo en
`OperacionExenta`. La base pasa a ser el total.

```json
"Desglose": {
  "DetalleDesglose": [
    {
      "Impuesto": "01",
      "ClaveRegimen": "01",
      "OperacionExenta": "E1",
      "BaseImponibleOimporteNoSujeto": 10000
    }
  ]
},
"CuotaTotal": 0,
"ImporteTotal": 10000
```

Los motivos de exención están en [03 — el desglose](03-desglose.md).

## Simplificada cualificada

Si la simplificada lleva los datos del destinatario porque él los pidió —el
art. 7.2 y 7.3 del reglamento—, se marca con:

```json
"FacturaSimplificadaArt7273": "S"
```

Y entonces el tipo pasa a ser `F1`, no `F2`.
