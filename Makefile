# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

compile:
	mkdir -p build/target
	cp config/config.json build/target
	cp -r cases build/target
	cp -r build/keystore build/target
	cp -r build/setup build/target
	cp -r scripts/* build/target/
	$(GOBUILD) -o build/target/robot cmd/main.go

compile-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o build/target/robot-linux cmd/main.go

robot:
	@echo test case $(t)
	./build/target/robot -config=build/target/config.json -t=$(t)

clean:
	rm -rf build/target/*