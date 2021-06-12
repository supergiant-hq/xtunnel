clean:
	@echo == Cleanup ==
	rm -rf bin
	@echo

proto:
	@echo == Generating protobuf code ==
	protoc --go_out=. model/*.proto
	@echo

bin-linux:
	@echo == Compiling Binaries for Linux ==
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/tunnel-linux-amd64 ./
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ./bin/tunnel-linux-arm64 ./
	@echo

bin-darwin:
	@echo == Compiling Binaries for Darwin ==
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/tunnel-darwin-amd64 ./
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ./bin/tunnel-darwin-arm64 ./
	@echo

bin-windows:
	@echo == Compiling Binaries for Windows ==
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/tunnel-windows-amd64.exe ./
	@echo

bin: bin-linux bin-darwin bin-windows

all: clean bin

