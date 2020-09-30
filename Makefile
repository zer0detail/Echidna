GOCMD=go 
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get 
GOMOD=$(GOCMD) mod
BINARY_NAME=Echidna
LINTER=golangci-lint 
TARGET=build
ARCHS=amd64 386
LDFLAGS="-s -w"
GCFLAGS="all=-trimpath=$(shell pwd)"
ASMFLAGS="all=-trimpath=$(s hell pwd)"

all: test build 

local: lint tests install

current:
	$(GOBUILD) -o ./$(BINARY_NAME)

windows:
	@for GOARCH in ${ARCHS}; do \
		echo "Building for windows $${GOARCH} ..." ; \
		mkdir -p ${TARGET}/gobuster-windows-$${GOARCH} ; \
		GOOS=windows GOARCH=$${GOARCH} GO111MODULE=on CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -gcflags=${GCFLAGS} -asmflags=${ASMFLAGS} -o ${TARGET}/gobuster-windows-$${GOARCH}/gobuster.exe ; \
	done; \
	echo "Done."

clean:
	if exist current rmdir  current /Q /S 
	if exist inspect rmdir inspect  /Q /S
	if exist cmd\echidna\current rmdir  cmd\echidna\current /Q /S 
	if exist cmd\echidna\inspect rmdir cmd\echidna\inspect  /Q /S
	if exist error.log del /f error.log
	if exist $(TARGET) rmdir $(TARGET) /Q /S
	go clean ./...

tests: 
	$(GOTEST) ./... -v -race

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

lint:
	$(LINTER) run ./...
	$(GOMOD) tidy

install:
	$(GOINSTALL)

