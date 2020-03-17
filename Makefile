BINARY_NAME=bin/proxy
COMMIT_ID := $(shell git rev-parse HEAD )


all:
	mkdir -p bin/
	cd cmd && go build -ldflags "-X main.COMMIT_ID='${COMMIT_ID}'" -o ../$(BINARY_NAME) -v &&cd -

clean:
	go clean
	rm -rf ./bin