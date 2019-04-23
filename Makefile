# Parameters for Go
BINARY_NAME=smartkey-kms

all: smartkey-kms

smartkey-kms: build

build:
	go build -o $(BINARY_NAME) -v

test:
	go test -v ./...

clean:
	go clean
	rm -f $(BINARY_NAME)

