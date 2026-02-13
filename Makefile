include .env
export

MIGRATIONS_DIR = db/migrations

.PHONY: run swag-gen migrate-create migrate-up migrate-down migrate-force

swag-gen:
	swag init -g cmd/main.go -o docs

run:
	go run ./cmd/main.go

migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=create_something"; exit 1; fi
	@next=$$(printf "%03d" $$(( $$(ls $(MIGRATIONS_DIR)/*.sql 2>/dev/null | wc -l) + 1 ))); \
	touch $(MIGRATIONS_DIR)/$${next}_$(name).sql; \
	echo "Created $(MIGRATIONS_DIR)/$${next}_$(name).sql"

migrate-up:
	@for f in $$(ls $(MIGRATIONS_DIR)/*.sql | sort); do \
		echo "Running $$f ..."; \
		psql "$(DATABASE_URL)" -f "$$f" || exit 1; \
	done
	@echo "All migrations applied."

migrate-down:
	@latest=$$(ls $(MIGRATIONS_DIR)/*.sql | sort | tail -1); \
	table=$$(head -1 "$$latest" | grep -oP '(?i)(?:CREATE TABLE IF NOT EXISTS |CREATE TABLE )\K\S+' | tr -d '('); \
	if [ -z "$$table" ]; then echo "Could not detect table from $$latest"; exit 1; fi; \
	echo "Dropping table $$table from $$latest ..."; \
	psql "$(DATABASE_URL)" -c "DROP TABLE IF EXISTS $$table CASCADE;"

migrate-force:
	@latest=$$(ls $(MIGRATIONS_DIR)/*.sql | sort | tail -1); \
	echo "Force re-running $$latest ..."; \
	psql "$(DATABASE_URL)" -f "$$latest"
