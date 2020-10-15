# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

compile:
	rm -rf build/target/*
	mkdir -p build/target/params
	cp cmd/config.json build/target/config.json
	cp cmd/wallet.dat build/target/wallet.dat
	cp cmd/transfer_wallet.dat build/target/transfer_wallet.dat
	cp -r cmd/params/* build/target/params/
	$(GOBUILD) -o build/target/robot cmd/main.go

compile-linux-robot:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o build/target/linux-robot cmd/main.go

robot:
	@echo test case $(t)
	./build/target/robot -config=build/target/config.json \
	-params=build/target/params \
	-wallet=build/target/wallet.dat \
	-transfer=build/target/transfer_wallet.dat \
	-t=$(t)

clean:
	rm -rf build/target/*