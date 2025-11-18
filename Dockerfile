FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/api/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/docs /app/docs
COPY --from=builder /app/.env /app/.env

RUN chmod +x ./server

CMD ["./server"]
