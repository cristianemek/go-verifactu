# Compila verifactud. Contexto: la raiz del repositorio (hace falta go.work).
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /verifactud ./cmd/verifactud
RUN mkdir -p /var/lib/verifactud

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /verifactud /verifactud
COPY --from=build --chown=65532:65532 /var/lib/verifactud /var/lib/verifactud
VOLUME /var/lib/verifactud
EXPOSE 9009
ENTRYPOINT ["/verifactud", "-config", "/etc/verifactud/verifactud.json"]
