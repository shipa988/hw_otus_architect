.EXPORT_ALL_VARIABLES:

COMPOSE_CONVERT_WINDOWS_PATHS=1

tidy:
	go mod tidy
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
run: dep_install
	go run main.go --config=config/network-dev.yaml --logdest std --loglevel debug
build: dep_install
	go build -o networkcore.exe  main.go
start:
	networkcore.exe  --config=config/network-dev.yaml --logdest std --loglevel debug
run_local: build start
.PHONY: build, all, fmt, lint, test, run,tidy