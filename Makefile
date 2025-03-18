TEST_DIR = ./test

.PHONY: proto migrate migrate-reset run test
migrate:
	go run main.go migrate

migrate-reset:
	go run main.go migrate-reset

air:
	air
run:
	go run main.go


test:
	@echo "Running Go tests..."
	go test -v $(TEST_DIR)/...

.PHONY: test
