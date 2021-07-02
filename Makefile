-include environ.inc
.PHONY: deps dev build install image release test clean

CGO_ENABLED=0
VERSION=$(shell git describe --abbrev=0 --tags 2>/dev/null || echo "$VERSION")
COMMIT=$(shell git rev-parse --short HEAD || echo "$COMMIT")
GOCMD=go

all: build

deps:

dev : DEBUG=1
dev : build
	./start-cluster.sh

fbox: generate
	@$(GOCMD) build -tags "netgo static_build" -installsuffix netgo \
		-ldflags "-w \
		-X $(shell go list).Version=$(VERSION) \
		-X $(shell go list).Commit=$(COMMIT)" \
		.

build: fbox

generate:
	@if [ x"$(DEBUG)" = x"1"  ]; then		\
	  echo 'Running in debug mode...';	\
	fi

install: build
	@$(GOCMD) install .

ifeq ($(PUBLISH), 1)
image:
	@docker build --build-arg VERSION="$(VERSION)" --build-arg COMMIT="$(COMMIT)" -t prologic/fbox .
	@docker push prologic/fbox
else
image:
	@docker build --build-arg VERSION="$(VERSION)" --build-arg COMMIT="$(COMMIT)" -t prologic/fbox .
endif

release:
	@./tools/release.sh

test:
	@$(GOCMD) test -v -cover -race ./...

bench: bench-twtxt.txt
	go test -race -benchtime=1x -cpu 16 -benchmem -bench "^(Benchmark)" github.com/jointwt/twtxt/types

clean:
	@git clean -f -d -X
