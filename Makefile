GOCMD=go 
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get 
BINARY_NAME=echidna
LINTER=golangci-lint 

all: test build 

local: lint tests install

clean:
	cmd /c if exist current rmdir  current /Q /S 
	cmd /c if exist inspect rmdir inspect  /Q /S
	cmd /c if exist cmd\echidna\current rmdir  cmd\echidna\current /Q /S 
	cmd /c if exist cmd\echidna\inspect rmdir cmd\echidna\inspect  /Q /S
	cmd /c if exist error.log del /f error.log

tests: 
	$(GOTEST) ./... -v

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

lint:
	$(LINTER) run

install:
	$(GOINSTALL)

