
# This Makefile is adopted from https://github.com/hashicorp/consul/blob/master/Makefile

all: format build

cov:
	gocov test | gocov-html > /tmp/coverage.html
	open /tmp/coverage.html

build: test
	cd cmd/gryffin-standalone; go build

test:
	go test ./...
	@$(MAKE) vet

test-mono:
	go run cmd/gryffin-standalone/main.go "http://127.0.0.1:8081"
	go run cmd/gryffin-standalone/main.go "http://127.0.0.1:8082/dvwa/vulnerabilities/sqli/?id=1&Submit=Submit"


test-integration:
	INTEGRATION=1 go test ./...

test-cover:
	go test --cover ./...

format:
	@gofmt -l .

vet:
	@go vet ./...

.PHONY: all cov build test vet web web-push
