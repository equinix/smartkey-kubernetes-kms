# Parameters for Go
BINARY_NAME=smartkey-kms

all: smartkey-kms

smartkey-kms: build

test:
	go test -v

build:
	go build -o $(BINARY_NAME) -v

clean:
	go clean
	rm -f $(BINARY_NAME)

