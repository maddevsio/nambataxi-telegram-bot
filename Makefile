run:
	go run main.go

test:
	go test -v ./...

build:
	env CGO_ENABLED=1 GOOS=linux GOARCH=386 go build -v -ldflags -linkmode=external
