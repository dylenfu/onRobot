# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

prepare:
	mkdir -p build/target
	cp -r build/keystore build/target
	cp -r build/setup build/target
	cp -r build/poly_keystore build/target
	cp -r scripts/*.sh build/target/

compile:
	cp -r cases build/target
	$(GOBUILD) -o build/target/robot cmd/main.go

compile-local:
	cp config/local.json build/target/config.json
	cp build/target/setup/local-nodes.json build/target/setup/static-nodes.json
	cp build/target/setup/local-genesis.json build/target/setup/genesis.json
	make compile

compile-remote:
	cp config/remote.json build/target/config.json
	cp build/target/setup/remote-nodes.json build/target/setup/static-nodes.json
	cp build/target/setup/remote-genesis.json build/target/setup/genesis.json
	make compile

compile-dev:
	cp config/dev.json build/target/config.json
	cp build/target/setup/dev-nodes.json build/target/setup/static-nodes.json
	cp build/target/setup/dev-genesis.json build/target/setup/dev-genesis.json
	make compile

compile-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o build/target/robot-linux cmd/main.go

robot:
	@echo test case $(t)
	./build/target/robot -config=build/target/config.json -t=$(t)

clean:
	rm -rf build/target/*