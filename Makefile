.EXPORT_ALL_VARIABLES:

COMPOSE_CONVERT_WINDOWS_PATHS=1

DBSCHEMAPATH = ./internal/data/repository/mysql/sql
DBCONNSTRING = "root:rootpw@tcp(localhost:3306)/networkcore"
MIGRATIONPATH = ./internal/data/repository/mysql/migrations
MYSQL_ROOT_PASSWORD = rootpw
MYSQL_PASSWORD = userpw

tidy:
	go mod tidy
timeout:
	timeout 5
dep_install:
	go get -d ./...
test: dep_install
	go test -race ./...
testv: dep_install
	go test -v -race ./...
prepare_lint:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0
lint: prepare_lint fmt tidy
	golangci-lint run ./...
fmt:
	go fmt ./...
run: #dep_install
	go run cmd/core/main.go --config=cmd/core/config/network-dev.yaml --logdest std --loglevel debug
build: dep_install
	go build -o ncmonolit.exe  cmd/core/main.go
start:
	ncmonolit.exe  --config=cmd/core/config/network-dev.yaml --logdest std --loglevel debug
run_local: build start
prepare_migrate:
	go get -u github.com/pressly/goose/cmd/goose@v2.7.0-rc4
migrate_db:timeout goose_up
goose_up:
	goose -dir $(MIGRATIONPATH) mysql $(DBCONNSTRING) up

docker-up:
	docker-compose -f docker-compose.yml up -d --build
docker-down:
	docker-compose -f docker-compose.yml down
docker-down-hard:
	docker-compose -f docker-compose.yml down -v --remove-orphans
	docker rmi  hw_otus_architect_ncmonolit:latest

up: docker-up prepare_migrate migrate_db
down: docker-down

up_prod: docker-up
down_prod: docker-down

.PHONY: build, all, fmt, lint, test, run,tidy