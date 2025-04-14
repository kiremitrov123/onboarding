# Build stage
FROM golang:1.23 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o commenting ./cmd

# Runtime stage
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/commenting .
EXPOSE 8080
CMD ["./commenting"]
