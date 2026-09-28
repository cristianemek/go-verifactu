# 07 — El QR y el CSV

## QR

El campo `qr` de la respuesta es la URL que va dentro del código:

```
https://prewww2.aeat.es/wlpl/TIKE-CONT/ValidarQR?fecha=10-09-2026&importe=121.00&nif=89890001K&numserie=T-2026-0001
```

La imagen la pintas tú. La AEAT pide: ISO/IEC 18004:2015 con corrección **M**,
entre 30x30 y 40x40 mm, 2 mm en blanco alrededor como mínimo, en la primera
página, con **«QR tributario:»** encima y **«VERI\*FACTU»** debajo.

Apunta a pruebas o a producción según el `entorno` del servicio.

## CSV

Sale en `/estado` cuando la AEAT acepta la factura:

```json
"aeat": { "estado": "Correcto", "csv": "A-3JLMAM3L8XVZD9" }
```

Es el justificante de la remisión, 16 caracteres, **del envío y no de la
factura**: varias facturas remitidas juntas comparten CSV. La AEAT **no lo
devuelve más adelante**, así que guárdalo. No hay obligación de imprimirlo.
