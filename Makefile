ATLAS_DEV_DB_URL = "postgres://postgres:test123@localhost:5435/postgres?sslmode=disable"
MIGRATIONS_DIR = "file://postgres/migrations"
DB_SCHEMA = "file://postgres/schema.sql"
DB_URL = "postgres://postgres:test123@localhost:5434/postgres?sslmode=disable"

define migrate_diff
	@read -p "Enter migration name: " NAME; \
	atlas migrate diff $$NAME --dir $(1) --to $(2) --dev-url $(3)
endef

define migrate_apply
	atlas migrate apply --dir $(1) --url $(2)
endef

.PHONY: makemigrations \
 		migrate \


makemigrations:
	$(call migrate_diff, ${MIGRATIONS_DIR}, ${DB_SCHEMA}, ${ATLAS_DEV_DB_URL})



migrate:
	$(call migrate_apply, ${MIGRATIONS_DIR}, $(DB_URL))