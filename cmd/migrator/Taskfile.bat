REM task:
REM     migrate:
REM         description: "Migrations"
REM         cmd:
go run ./cmd/migrator/main.go --storage-path=./storage/sso.db --migrations-path=./migrations

REM         description: "TestMigrations"
REM         cmd:
go run ./cmd/migrator/main.go --storage-path=./storage/sso.db --migrations-path=./tests/migrations --migrations-table=migrations_test
