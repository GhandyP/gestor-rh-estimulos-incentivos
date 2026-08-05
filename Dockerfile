# syntax=docker/dockerfile:1
# Multi-stage, CGO-free: build estático + runtime mínimo con templates.
# Imágenes fijadas para builds reproducibles.

# ---- Etapa de build ----
FROM golang:1.26.2-alpine AS build
WORKDIR /src

# Dependencias primero (capa cacheable).
COPY go.mod go.sum ./
RUN go mod download

# Código fuente.
COPY . .

# Binario estático, sin CGO (modernc.org/sqlite es Go puro).
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---- Etapa runtime ----
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
ENV APP_ROOT=/app

COPY --from=build /out/server /app/server
# Templates dentro de la imagen: el arranque los carga desde /app (cwd-independiente).
COPY web/ /app/web/

EXPOSE 8080
ENV PORT=8080
ENV DB_PATH=/data/estimulos.db
VOLUME ["/data"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/readyz || exit 1

ENTRYPOINT ["/app/server"]
