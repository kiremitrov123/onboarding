FROM golang:1.23 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ogpreview ./cmd

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/ogpreview .
EXPOSE 8080
CMD ["./ogpreview"]
