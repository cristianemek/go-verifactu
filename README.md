# go-verifactu

Librería y servicio en Go para VERI*FACTU, el sistema de facturación de la AEAT.

Calcula la huella encadenada, monta el QR y envía los registros a Hacienda.

Dos formas de usarlo:

- **Binario** en tu VPS. Levantas el servicio, le hablas por HTTP desde PHP,
  Node, Python o lo que uses. No hace falta saber Go.
- **Librería** si tu proyecto ya es Go. `go get`.

 En desarrollo. La API cambia y todavía no vale para producción.

## Como servicio

```
go install github.com/cristianemek/go-verifactu/cmd/verifactu@latest
verifactu serve
```

Un binario estático, sin runtime ni dependencias. Cópialo al VPS y ya está.

```
curl -X POST localhost:8080/v1/alta \
  -H 'Content-Type: application/json' \
  -d @factura.json
```

Te devuelve el registro con su huella y la URL del QR para imprimir en la
factura. El envío a la AEAT va en cola por detrás.

## Como librería

```
go get github.com/cristianemek/go-verifactu
```

Go 1.22+. Sin dependencias.

## Qué cubre

Solo la modalidad VERI*FACTU y territorio común. No hace firma XAdES, ni
TicketBAI, ni factura electrónica B2B.

## Aviso

Esto no es un SIF, es una herramienta para construir uno. No lleva declaración
responsable: si la usas en tu software de facturación, esa parte te toca a ti,
igual que revisar el código y comprobar que cumple.

Ver el [Artículo 13 del RD 1007/2023](https://www.boe.es/buscar/act.php?id=BOE-A-2023-24840#a1-5).

## Licencia

[MIT](LICENSE). Se puede utilizar este proyecto para cualquier uso, incluso comercial, siempre que se haga referencia al uso y autoría del mismo. No se ofrece ninguna garantía de funcionamiento ni soporte. El uso de este proyecto es bajo la responsabilidad del usuario.