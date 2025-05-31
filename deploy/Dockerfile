# Stage 1
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gophkeeper ./cmd/server

# Stage 2
FROM alpine:3.21
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /home/appuser
COPY --from=builder --chown=appuser:appgroup /app/gophkeeper .

RUN mkdir -p /etc/ssl/private && \
  chown appuser:appgroup /etc/ssl/private && \
  chmod 750 /etc/ssl/private

COPY --chown=appuser:appgroup cert.pem key.pem /etc/ssl/private/
RUN chmod 440 /etc/ssl/private/key.pem && \
  chmod 444 /etc/ssl/private/cert.pem

USER appuser
EXPOSE 8080

# hardcode cert and key files for simplicity
CMD ["./gophkeeper", "--c=/etc/ssl/private/cert.pem", "--k=/etc/ssl/private/key.pem"]
