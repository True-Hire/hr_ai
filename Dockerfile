FROM golang:1.25-alpine AS builder

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN swag init -g cmd/main.go -o docs
RUN CGO_ENABLED=0 go build -o /app/hr-ai ./cmd/main.go
RUN CGO_ENABLED=0 go build -o /app/seeder ./cmd/seeder/

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/hr-ai .
COPY --from=builder /app/seeder .
COPY db/migrations ./db/migrations
COPY web ./web

EXPOSE 11911

CMD ["./hr-ai"]
