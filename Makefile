SHELL=/usr/bin/env bash

GOCC?=go
BINS:=

#version=$(shell bash ./version.sh)
#ldflags=-X=github.com/SaoNetwork/sao-node/build.CurrentCommit=$(subst -,.,"$(version)")
#GOFLAGS+=-ldflags="$(ldflags)"

all: payment-gateway

payment-gateway:
	$(GOCC) build $(GOFLAGS) -o saopayment ./cmd/payment-gateway
.PHONY: payment-gateway
BINS+=saopayment

cbor-gen:
	$(GOCC) run ./gen/cbor/cbor_gen.go
.PHONY: cbor-gen

api-gen:
	$(GOCC) run ./gen/api
	goimports -w api
	goimports -w api
.PHONY: api-gen

docsgen-md-bin:
	$(GOCC) build $(GOFLAGS) -o docgen-md ./gen/apidoc

docsgen-md: docsgen-md-bin
	./docgen-md "api/api_gateway.go" "SaoApi" "api" "./api" > docs/api.md

docsgen-cfg:
	$(GOCC) run ./gen/cfgdoc > ./payment-gateway/config/doc_gen.go

clean:
	rm -rf $(BINS)
.PHONY: clean
