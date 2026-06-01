FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache docker-cli

COPY --from=builder /app/server .
# We also need to copy the problems directory to be available to the backend!
COPY --from=builder /app/problems ./problems
COPY --from=builder /app/.env .env

EXPOSE 8080
CMD ["./server"]
