.PHONY: all fmt test docker-build offline-zip

all: fmt test

fmt:
	go fmt ./...

test:
	go test ./...

docker-build:
	docker build -t adi-ber/vjal-platform:latest .

offline-zip:
	./build.sh
