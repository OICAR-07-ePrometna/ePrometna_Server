FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -v -ldflags="-s -w" -o bin/ePrometna_Server .
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bin/ePrometna_Server .

EXPOSE 8090
ENTRYPOINT ["./ePrometna_Server"]
