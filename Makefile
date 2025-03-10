GOLANGCI_LINT_CACHE?=praktikum-golangci-lint-cache
current_dir := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))

.PHONY: lint
lint:
	-docker run --rm -v .:/source -v $(GOLANGCI_LINT_CACHE):/root/.cache -w //source golangci/golangci-lint golangci-lint run -c .golangci.yml
	-bash -c 'cat ./.golangci-lint/report-unformatted.json | jq > ./.golangci-lint/report.json'


.PHONY: build-server
build-server:
	go build -C cmd/server \
		-ldflags "-X 'main.buildVersion=${appver}' -X 'main.buildDate=$(shell date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=${shell git rev-parse --short HEAD}'"

.PHONY: build-client
build-client:
	go build -C cmd/client \
		-ldflags "-X 'main.buildVersion=${appver}' -X 'main.buildDate=$(shell date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=${shell git rev-parse --short HEAD}'"

.PHONY: build
build: build-server, build-client

.PHONY: test
test:
	go test ./...
.PHONY: test-cover
test-cover: test
	go test ./... -cover

.PHONY: run-server
run-server:
	go run ./cmd/server/main.go -c .var/conf/server.cfg

.PHONY: run-client
run-client:
	go run ./cmd/client/main.go -c .var/conf/client.cfg

.PHONY: db-start
db-start:
	docker compose -f "scripts/db/docker-compose.yaml" up -d --build

.PHONY: db-stop
db-stop:
	docker compose -f "scripts/db/docker-compose.yaml" down

.PHONY: db-migration-new
db-migration-new:
	docker run --rm \
    -v $(realpath ./internal/adapter/storage/migrations):/migrations \
    migrate/migrate:v4.16.2 \
        create \
        -dir /migrations \
        -ext .sql \
        -seq -digits 5 \
        $(name)

.PHONY: db-migration-up
db-migration-up:
	docker run --rm \
    -v $(realpath ./internal/adapter/storage/migrations):/migrations \
	--network host \
    migrate/migrate:v4.16.2 \
        -path=/migrations \
        -database postgres://metrics:metrics@localhost:5432/metrics_db?sslmode=disable \
        up 

.PHONY: db-migration-down
db-migration-down:
	docker run --rm \
    -v $(realpath ./internal/adapter/storage/migrations):/migrations \
	--network host \
    migrate/migrate:v4.16.2 \
        -path=/migrations \
        -database postgres://metrics:metrics@localhost:5432/metrics_db?sslmode=disable \
        down 1

.PHONY: swagger-editor
swagger-editor:
	docker run --rm \
	-d -p 80:8080 \
	-v $(realpath ./docs/swagger):/specs \
	-e SWAGGER_FILE=/specs/swagger.yaml \
	swaggerapi/swagger-editor
