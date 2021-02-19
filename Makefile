# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
ENV=$(env)

prepare:
	make clean
	mkdir -p build/target
	cp -r build/env/$(ENV)/keystore build/target
	cp -r build/env/$(ENV)/poly_keystore build/target
	cp -r build/env/$(ENV)/eth_keystore build/target
	cp -r build/env/$(ENV)/setup build/target
	cp config/$(ENV).json build/target/config.json
	cp -r scripts/*.sh build/target/

compile:
	rm -rf build/target/case
	cp -r build/env/$(ENV)/cases build/target
	$(GOBUILD) -o build/target/robot cmd/main.go

compile-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o build/target/robot-linux cmd/main.go

robot:
	@echo test case $(t)
	./build/target/robot -config=build/target/config.json -t=$(t)

clean:
	rm -rf build/target/keystore
	rm -rf build/target/poly_keystore
	rm -rf build/target/eth_keystore
	rm -rf build/target/setup
	rm -rf build/target/case