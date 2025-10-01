.DEFAULT_GOAL := build

export PROJECT := "outline-export"
export PACKAGE := "github.com/lrstanley/outline-export"

license:
	curl -sL https://liam.sh/-/gh/g/license-header.sh | bash -s

clean:
	/bin/rm -rfv ${PROJECT}

fetch:
	go mod tidy

up:
	go get -u ./...
	go get -u -t ./...
	go mod tidy

prepare: clean fetch
	go generate -x ./...
	go run . generate-markdown > USAGE.md

dlv: prepare
	dlv debug \
		--headless --listen=:2345 \
		--api-version=2 --log \
		--allow-non-terminal-interactive \
		${PACKAGE} -- \
		--debug

debug: prepare
	rm -rf ./tmp
	go run ${PACKAGE} \
		--debug \
		--extract \
		--format markdown \
		--export-path ./tmp/

build: prepare
	CGO_ENABLED=0 \
	go build \
		-ldflags '-d -s -w -extldflags=-static' \
		-tags=netgo,osusergo,static_build \
		-installsuffix netgo \
		-trimpath \
		-o ${PROJECT} \
		${PACKAGE}
