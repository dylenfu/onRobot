SHELL=/bin/bash

# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
ENV=$(ONROBOT)

prepare:
	@cp config/$(ENV).json build/$(ENV)/config.json
	@cp -r scripts/*.sh build/$(ENV)/

compile:
	@$(GOBUILD) -o build/$(ENV)/robot cmd/main.go

compile-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o build/$(ENV)/robot-linux cmd/main.go

robot:
	@echo test case $(t)
	./build/$(ENV)/robot -config=build/$(ENV)/config.json -t=$(t)

clean: