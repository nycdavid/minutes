.PHONY: test
test:
	go test ./...

.PHONY: generate-migration
generate-migration:
	@migrate create -ext sql -dir db/migrations -seq $(NAME)


.PHONY: migrate-all-up
migrate-all-up:
	ENV="dev" go run ./cmd/migrate up
	ENV="test" go run ./cmd/migrate up

.PHONY: migrate-all-down
migrate-all-down:
	ENV="dev" go run ./cmd/migrate down
	ENV="test" go run ./cmd/migrate down
