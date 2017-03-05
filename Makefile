run:
	go run main.go

test:
	go test -v ./...

build:
	env GOOS=linux GOARCH=386 go build -v
