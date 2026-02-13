include .env
export

.PHONY: run swag-gen

swag-gen:
	swag init -g cmd/main.go -o docs

run:
	go run ./cmd/main.go
