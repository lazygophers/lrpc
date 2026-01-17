.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make up           - Start MongoDB container"
	@echo "  make down         - Stop MongoDB container"
	@echo "  make logs         - Show MongoDB logs"
	@echo "  make test         - Run tests with MongoDB"
	@echo "  make test-cov     - Run tests with coverage"
	@echo "  make clean        - Clean up containers and volumes"
	@echo "  make shell        - Open MongoDB shell"

.PHONY: up
up:
	docker-compose up -d
	@echo "MongoDB started on port 27017"
	@echo "Mongo Express available at http://localhost:8081"
	@echo "Username: admin, Password: admin"

.PHONY: down
down:
	docker-compose down

.PHONY: logs
logs:
	docker-compose logs -f mongo

.PHONY: shell
shell:
	docker-compose exec mongo mongosh -u root -p root --authenticationDatabase admin mongodb://localhost:27017

.PHONY: test
test: up
	@sleep 2
	go test ./... -v -race
	$(MAKE) down

.PHONY: test-cov
test-cov: up
	@sleep 2
	go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	$(MAKE) down

.PHONY: clean
clean:
	docker-compose down -v
	rm -f coverage.out coverage.html

.PHONY: restart
restart: down up

.PHONY: status
status:
	docker-compose ps

# Development targets
.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: build
build:
	go build ./...

.PHONY: vet
vet:
	go vet ./...
