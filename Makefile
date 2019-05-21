# Parameters for Go
BINARY_NAME=smartkey-kms

all: smartkey-kms

smartkey-kms: build

test:
	go get ./...
	go test -v

build:
	go get ./...
	go build -o $(BINARY_NAME) -v

clean:
	go get ./...
	go clean
	rm -f $(BINARY_NAME)

