# ---------- Build stage ----------
FROM golang:1.22-alpine AS builder

WORKDIR /workspace

ENV GOWORK=off

COPY shared/ shared/
COPY services/community/ services/community/

WORKDIR /workspace/services/community
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/community ./cmd/api

# ---------- Runtime stage ----------
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/community /app/community
COPY keys /app/keys
EXPOSE 8002
ENTRYPOINT ["/app/community"]