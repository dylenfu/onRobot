# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

prepare:
	mkdir -p build/target
	cp -r scripts/*.sh build/target/

compile:
	cp -r cases build/target
	$(GOBUILD) -o build/target/robot cmd/main.go

compile-local:
	make clean
	cp -r build/env/local/keystore build/target
	cp -r build/env/local/poly_keystore build/target
	cp -r build/env/local/setup build/target
	cp config/local.json build/target/config.json

compile-dev:
	make clean
	cp -r build/env/dev/keystore build/target
	cp -r build/env/dev/poly_keystore build/target
	cp -r build/env/dev/setup build/target
	cp config/dev.json build/target/config.json

compile-test:
	make clean
	cp -r build/env/test/keystore build/target
	cp -r build/env/test/poly_keystore build/target
	cp -r build/env/test/setup build/target
	cp config/test.json build/target/config.json

compile-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o build/target/robot-linux cmd/main.go

robot:
	@echo test case $(t)
	./build/target/robot -config=build/target/config.json -t=$(t)

clean:
	rm -rf build/target/keystore
	rm -rf build/target/poly_keystore
	rm -rf build/target/setup
	rm -rf build/target/case