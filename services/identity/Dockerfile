# ---------- Build stage ----------
FROM golang:1.22-alpine AS builder

WORKDIR /workspace

# Copy go workspace and module manifests first for better layer caching
COPY go.work ./
COPY shared/go.mod shared/
COPY services/identity/go.mod services/identity/
COPY services/community/go.mod services/community/
COPY services/recommendations/go.mod services/recommendations/

# Now copy source code
COPY shared/ shared/
COPY services/identity/ services/identity/
COPY services/community/ services/community/
COPY services/recommendations/ services/recommendations/

WORKDIR /workspace/services/identity
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/identity ./cmd/api

# ---------- Runtime stage ----------
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/identity /app/identity
EXPOSE 8001
ENTRYPOINT ["/app/identity"]
