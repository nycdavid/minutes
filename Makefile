.PHONY: test
test:
	go test ./...

.PHONY: generate-migration
generate-migration:
	@migrate create -ext sql -dir db/migrations -seq $(NAME)


.PHONY: migrate-all
migrate-all:
	ENV="dev" go run ./cmd/migrate
	ENV="test" go run ./cmd/migrate
