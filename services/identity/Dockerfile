# ---------- Build stage ----------
FROM golang:1.22-alpine AS builder

WORKDIR /workspace

# Ignorar go.work: cada servicio usa su propio go.mod con `replace` al shared
ENV GOWORK=off

# Copiar solo lo que identity necesita: shared + identity
COPY shared/ shared/
COPY services/identity/ services/identity/

WORKDIR /workspace/services/identity
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/identity ./cmd/api

# ---------- Runtime stage ----------
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/identity /app/identity
COPY keys /app/keys
EXPOSE 8001
ENTRYPOINT ["/app/identity"]