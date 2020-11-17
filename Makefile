# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

prepare:
	mkdir -p build/target
	cp -r build/keystore build/target
	cp -r build/setup build/target
	cp -r scripts/*.sh build/target/

compile:
	cp -r cases build/target
	$(GOBUILD) -o build/target/robot cmd/main.go

compile-local:
	cp config/local.json build/target/config.json
	make compile

compile-remote:
	cp config/remote.json build/target/config.json
	rm -rf build/target/setup/static-nodes.json
	mv build/target/setup/scp-nodes.json build/target/setup/static-nodes.json
	make compile

compile-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o build/target/robot-linux cmd/main.go

robot:
	@echo test case $(t)
	./build/target/robot -config=build/target/config.json -t=$(t)

clean:
	rm -rf build/target/*