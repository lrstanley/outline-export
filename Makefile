.DEFAULT_GOAL := build-all

export PROJECT := "YOUR_PROJECT_NAME"
export PACKAGE := "github.com/lrstanley/YOUR_PROJECT_NAME"

license:
	curl -sL https://liam.sh/-/gh/g/license-header.sh | bash -s

prepare: clean go-prepare
	@echo

build-all: prepare go-build
	@echo

clean:
	/bin/rm -rfv ${PROJECT}

go-fetch:
	go mod download
	go mod tidy

up:
	go get -u ./...
	go get -u -t ./...
	go mod tidy

go-prepare: go-fetch
	go generate -x ./...

go-dlv: go-prepare
	dlv debug \
		--headless --listen=:2345 \
		--api-version=2 --log \
		--allow-non-terminal-interactive \
		${PACKAGE} -- --debug

go-debug: go-prepare
	go run ${PACKAGE} \
		--debug \
		--dry-run \
		--check-dir ./tests/

go-build: go-prepare
	CGO_ENABLED=0 \
	go build \
		-ldflags '-d -s -w -extldflags=-static' \
		-tags=netgo,osusergo,static_build \
		-installsuffix netgo \
		-trimpath \
		-o ${PROJECT} \
		${PACKAGE}
