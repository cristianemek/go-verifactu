# Ejemplos

Casuística de facturación, con el JSON que se manda en cada caso.

| | |
| --- | --- |
| [01](01-simplificada.md) | factura simplificada (F2) |
| [02](02-completa.md) | factura completa (F1) |
| [03](03-desglose.md) | el desglose: IVA, exentas, no sujetas, ISP, recargo, IGIC e IPSI |
| [04](04-rectificativas.md) | rectificativas: total y parcial |
| [05](05-anulacion.md) | anulación |
| [06](06-estado-y-correccion.md) | ver qué contestó la AEAT y corregir |
| [07](07-qr-y-csv.md) | el QR y el CSV |

Todo va a `POST /v1/{nif}/alta` con `Authorization: Bearer TU_TOKEN`, salvo la
anulación y las consultas.

Endpoints, convenciones y códigos de error: [../docs/api.md](../docs/api.md).
