APP            ?= dynatrace-sqs-events
PKG            ?= ./cmd/lambda
ZIP            ?= function.zip
ARCH           ?= arm64
GOOS           ?= linux
CGO_ENABLED    ?= 0
BIN            ?= bootstrap

.PHONY: build zip clean

build:
	GOOS=$(GOOS) GOARCH=$(ARCH) CGO_ENABLED=$(CGO_ENABLED) go build -tags lambda.norpc -trimpath -ldflags '$(LDFLAGS)' -gcflags '$(GCFLAGS)' -o $(BIN) $(PKG)
	
zip: build
	go run ./tools/zipper/main.go

clean:
	-go run ./tools/cleaner/main.go $(BIN) $(ZIP)