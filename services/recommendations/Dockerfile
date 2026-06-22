# ---------- Build stage ----------
FROM golang:1.22-alpine AS builder

WORKDIR /workspace

COPY go.work ./
COPY shared/go.mod shared/
COPY services/identity/go.mod services/identity/
COPY services/community/go.mod services/community/
COPY services/recommendations/go.mod services/recommendations/

COPY shared/ shared/
COPY services/identity/ services/identity/
COPY services/community/ services/community/
COPY services/recommendations/ services/recommendations/

WORKDIR /workspace/services/recommendations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/recommendations ./cmd/api

# ---------- Runtime stage ----------
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/recommendations /app/recommendations
COPY keys /app/keys

EXPOSE 8003
ENTRYPOINT ["/app/recommendations"]
