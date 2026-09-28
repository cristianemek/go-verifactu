# 03 — El desglose

`DetalleDesglose` es una lista de hasta 12 líneas: **una por cada combinación de
impuesto, régimen, calificación y tipo** de la factura.

Cada línea lleva `CalificacionOperacion` **o** `OperacionExenta`, nunca las dos.

| `CalificacionOperacion` | |
| --- | --- |
| `S1` | sujeta y no exenta, IVA normal |
| `S2` | sujeta y no exenta **con inversión del sujeto pasivo** |
| `N1` | no sujeta, art. 7, 14 y otros |
| `N2` | no sujeta por reglas de localización |

## Sujeta con IVA (S1)

```json
{
  "Impuesto": "01",
  "ClaveRegimen": "01",
  "CalificacionOperacion": "S1",
  "TipoImpositivo": 2100,
  "BaseImponibleOimporteNoSujeto": 100000,
  "CuotaRepercutida": 21000
}
```

`TipoImpositivo` y `CuotaRepercutida` son obligatorios aquí, y solo aquí puede la
cuota ser distinta de cero.

## Varios tipos en la misma factura

Una línea por tipo. Los totales suman todas:

```json
"Desglose": { "DetalleDesglose": [
  { "Impuesto": "01", "ClaveRegimen": "01", "CalificacionOperacion": "S1", "TipoImpositivo": 2100, "BaseImponibleOimporteNoSujeto": 100000, "CuotaRepercutida": 21000 },
  { "Impuesto": "01", "ClaveRegimen": "01", "CalificacionOperacion": "S1", "TipoImpositivo": 1000, "BaseImponibleOimporteNoSujeto": 50000, "CuotaRepercutida": 5000 },
  { "Impuesto": "01", "ClaveRegimen": "01", "CalificacionOperacion": "S1", "TipoImpositivo": 400, "BaseImponibleOimporteNoSujeto": 20000, "CuotaRepercutida": 800 }
]},
"CuotaTotal": 26800,
"ImporteTotal": 196800
```

## Inversión del sujeto pasivo (S2)

El IVA lo declara el destinatario, así que tú no lo repercutes: **tipo y cuota a
cero**, obligatoriamente.

```json
{
  "Impuesto": "01",
  "ClaveRegimen": "01",
  "CalificacionOperacion": "S2",
  "TipoImpositivo": 0,
  "BaseImponibleOimporteNoSujeto": 100000,
  "CuotaRepercutida": 0
}
```

`"CuotaTotal": 0` y `"ImporteTotal": 100000`.

Con `S2` el tipo de factura no puede ser `F2` ni `R5`: va como `F1`, `F3` o
`R1` a `R4`.

## Exenta

El motivo va en `OperacionExenta`, y **no se informan** tipo, cuota, ni nada de
recargo:

```json
{
  "Impuesto": "01",
  "ClaveRegimen": "01",
  "OperacionExenta": "E1",
  "BaseImponibleOimporteNoSujeto": 100000
}
```

| | artículo de la Ley del IVA |
| --- | --- |
| `E1` | art. 20 — exenciones en operaciones interiores (sanidad, educación, alquiler de vivienda…) |
| `E2` | art. 21 — exportaciones |
| `E3` | art. 22 — operaciones asimiladas a las exportaciones |
| `E4` | art. 23 y 24 — zonas francas, depósitos y regímenes aduaneros |
| `E5` | art. 25 — entregas intracomunitarias |
| `E6` | otros |

Con `ClaveRegimen` `01` no se admiten `E2` ni `E3`.

## No sujeta

Igual: sin tipo ni cuota. La base es el importe no sujeto.

```json
{
  "Impuesto": "01",
  "ClaveRegimen": "01",
  "CalificacionOperacion": "N1",
  "BaseImponibleOimporteNoSujeto": 100000
}
```

`N1` para las no sujetas del art. 7 y 14; `N2` cuando no lo están por las reglas
de localización, por ejemplo un servicio prestado fuera del territorio.

## Recargo de equivalencia

Dos campos más en la línea, y el recargo suma al total:

```json
{
  "Impuesto": "01",
  "ClaveRegimen": "01",
  "CalificacionOperacion": "S1",
  "TipoImpositivo": 2100,
  "BaseImponibleOimporteNoSujeto": 100000,
  "CuotaRepercutida": 21000,
  "TipoRecargoEquivalencia": 520,
  "CuotaRecargoEquivalencia": 5200
}
```

`"CuotaTotal": 26200` y `"ImporteTotal": 126200`: la cuota total incluye el
recargo, y el importe total es base más cuota más recargo.

Los pares que admite la AEAT:

| `TipoImpositivo` | `TipoRecargoEquivalencia` |
| --- | --- |
| 21 % | 5,2 % |
| 10 % | 1,4 % |
| 4 % | 0,5 % |

## IGIC e IPSI

Cambia el impuesto; en Canarias los tipos son los del IGIC:

```json
{
  "Impuesto": "03",
  "ClaveRegimen": "01",
  "CalificacionOperacion": "S1",
  "TipoImpositivo": 700,
  "BaseImponibleOimporteNoSujeto": 100000,
  "CuotaRepercutida": 7000
}
```

| `Impuesto` | |
| --- | --- |
| `01` | IVA (o vacío) |
| `02` | IPSI, Ceuta y Melilla |
| `03` | IGIC, Canarias |
| `05` | otros |

## Regímenes especiales

`ClaveRegimen` es `01` en el régimen general. Para los especiales, la AEAT
impone además estas reglas, que conviene tener a mano:

| clave | regla |
| --- | --- |
| `02` | solo se puede informar `OperacionExenta` |
| `03` | solo `S1` |
| `04` | solo `S2` o exenta |
| `06` | exige `BaseImponibleACoste`, y no vale `F2`, `F3` ni `R5` |
| `07` | no admite `E2`, `E3`, `E4`, `E5`, ni `S2`, `N1`, `N2` |
| `08` | exige `N2` |
| `10` | exige `N1`, `F1` y destinatario con NIF |
| `11` | el tipo tiene que ser el 21 % |
| `14` | exige `FechaOperacion` posterior a la de expedición, `F1`/`R1`-`R4`, y NIF del destinatario empezando por P, Q, S o V |
| `20` | exige `N2` |

La lista completa de claves y su significado está en el documento de listas de
la AEAT; los códigos de error están en
[../docs/aeat/errores.properties](../docs/aeat/errores.properties).
