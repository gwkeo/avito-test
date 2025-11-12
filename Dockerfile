FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/api

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .

CMD ["./server"]