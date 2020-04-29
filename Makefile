
# This Makefile is adopted from https://github.com/hashicorp/consul/blob/master/Makefile

DEPS = $(shell go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

PACKAGES = $(shell go list ./...)
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods \
         -nilfunc -rangeloops -shift -structtags -unsafeptr
         #-printf

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
	@go fmt $(PACKAGES)

vet:
	@go tool vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@go tool vet $(VETARGS) . ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for reviewal."; \
	fi

.PHONY: all cov build test vet web web-push
