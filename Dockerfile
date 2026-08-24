# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o /main ./cmd/app

FROM gcr.io/distroless/static-debian12

COPY --from=builder --chown=nonroot:nonroot /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder --chown=nonroot:nonroot /main /app/main
# Файл .env не копируется, чтобы не запекать секреты.
COPY --from=builder --chown=nonroot:nonroot /src/configs/.env.example /app/configs/.env.example

USER nonroot:nonroot
ENTRYPOINT ["/app/main"]
