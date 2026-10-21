# ---- build ----
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o stress-test ./cmd


# ---- runtime ----
FROM alpine:3.20
WORKDIR /app
# Certificados CA para testar URLs https
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/stress-test .
ENTRYPOINT ["/app/stress-test"]