include .env
export

MIGRATIONS_DIR = db/migrations
MIGRATE = migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)?sslmode=disable"

.PHONY: run swag-gen migrate-create migrate-up migrate-down migrate-force

swag-gen:
	swag init -g cmd/main.go -o docs

run:
	go run ./cmd/main.go

migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=create_something"; exit 1; fi
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq -digits 3 $(name)

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

migrate-force:
	@version=$$(ls $(MIGRATIONS_DIR)/*.up.sql | sort | tail -1 | grep -oP '\d+' | head -1); \
	echo "Forcing version $$version ..."; \
	$(MIGRATE) force $$version
