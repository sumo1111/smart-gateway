VERSION := $(shell cat VERSION)

.PHONY: all build-linux build-windows build-mac dev clean docker

all: build-linux

build-linux:
	CGO_ENABLED=1 go build -trimpath -ldflags "-s -w -X 'main.Version=$(VERSION)'" -o smart-gateway .

build-windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc 		go build -trimpath -ldflags "-s -w -X 'main.Version=$(VERSION)'" -o smart-gateway.exe .

build-mac:
	CGO_ENABLED=1 go build -trimpath -ldflags "-s -w -X 'main.Version=$(VERSION)'" -o smart-gateway .

dev:
	go run .

clean:
	rm -f smart-gateway smart-gateway.exe

docker:
	docker build -t sumo1111/smart-gateway:$(VERSION) .
	docker tag sumo1111/smart-gateway:$(VERSION) sumo1111/smart-gateway:latest

frontend:
	cd web && npm install && npm run build

setup-mingw:
	# Ubuntu: sudo apt install gcc-mingw-w64-x86-64
	@echo "Install: sudo apt install gcc-mingw-w64-x86-64"
