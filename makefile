default: air

build:
	go build -race -o bin/main src/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/main.exe src/main.go

test:
	go test ./...

air:
	air -build.cmd "make build" -build.bin "bin/main"
