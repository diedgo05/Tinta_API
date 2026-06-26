# ---------- Build stage ----------
FROM golang:1.22-alpine AS builder

WORKDIR /workspace

ENV GOWORK=off

COPY shared/ shared/
COPY services/recommendations/ services/recommendations/

WORKDIR /workspace/services/recommendations
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/recommendations ./cmd/api

# ---------- Runtime stage ----------
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/recommendations /app/recommendations
COPY keys /app/keys
EXPOSE 8003
ENTRYPOINT ["/app/recommendations"]