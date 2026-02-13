FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/hr-ai ./cmd/main.go

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/hr-ai .
COPY db/migrations ./db/migrations

EXPOSE 8080

CMD ["./hr-ai"]
