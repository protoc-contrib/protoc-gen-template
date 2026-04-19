GOPKG ?= github.com/protoc-contrib/protoc-gen-go-template
DOCKER_IMAGE ?= protoc-contrib/protoc-gen-go-template
GOBINS ?= .
GOLIBS ?= .

all: test install

include rules.mk

.PHONY: examples
examples:	install
	cd examples/time && make
	cd examples/enum && make
	cd examples/import && make
	cd examples/dummy && make
	cd examples/flow && make
	cd examples/concat && make
	cd examples/flow && make
	cd examples/sitemap && make
	cd examples/go-generate && make
  #cd examples/single-package-mode && make
	cd examples/helpers && make
	cd examples/arithmetics && make
  #cd examples/go-kit && make
