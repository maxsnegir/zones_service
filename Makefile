MIGRATOR_NAME?=migrator


.PHONY: make-migrations
make-migrations:
	migrate create -ext sql -dir migrations -seq $(NAME)


.PHONY: migrate
migrate:
	cd cmd/migrator/ && go run main.go --op up


.PHONY: lint
lint: tools ## Check the project with lint.
	@golangci-lint run --fix ./...

tools: ## Install all needed tools, e.g.
	@go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2


.PHONY: cover
cover:
	go test -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out


.PHONY: test
test:
	go test -v ./...


.PHONY: gen
gen:
	mockgen -source=internal/app/http/handlers.go \
	-destination=internal/app/http/mocks/mock_storage.go

