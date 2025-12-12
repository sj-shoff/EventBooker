FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o event-booker ./cmd/event-booker/main.go

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata curl postgresql-client

COPY --from=builder /app/event-booker /app/
COPY static /app/static
COPY templates /app/templates

EXPOSE 8005

CMD ["./event-booker"]