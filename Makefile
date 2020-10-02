GOCMD=go 
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get 
NAME=Echidna
TARGET=./build
ARCHS=amd64 386
LDFLAGS="-s -w"
GCFLAGS="all=-trimpath=$(pwd)"
ASMFLAGS="all=-trimpath=$(pwd)"

all: clean update lint test darwin windows linux

local: lint test install

install:
	go install

clean:
	rm -rf  current/  inspect/ error.log cmd/echidna/current/ cmd/echidna/inspect build/; \
	go clean ./... ; \
	echo "Cleaning Complete."

test: 
	$(GOTEST) ./... -v -race;\
	echo "Testing Complete."

lint:
	golangci-lint run; \
	go mod tidy; \
	echo "Linting Complete."

update:
	go get -u; \
	go mod tidy -v; \
	echo "Updating Complete."

windows:
	@for GOARCH in ${ARCHS}; do \
		echo "Building $(NAME) for windows $${GOARCH} ..." ; \
		mkdir -p ${TARGET}/$(NAME)-windows-$${GOARCH} ; \
		GOOS=windows GOARCH=$${GOARCH} GO111MODULE=on CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -gcflags=${GCFLAGS} -asmflags=${ASMFLAGS} -o ${TARGET}/$(NAME)-windows-$${GOARCH}/$(NAME).exe ; \
	done; \
	echo "Done."

linux:
	@for GOARCH in ${ARCHS}; do \
		echo "Building $(NAME) for linux $${GOARCH} ..." ; \
		mkdir -p ${TARGET}/$(NAME)-linux-$${GOARCH} ; \
		GOOS=linux GOARCH=$${GOARCH} GO111MODULE=on CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -gcflags=${GCFLAGS} -asmflags=${ASMFLAGS} -o ${TARGET}/$(NAME)-linux-$${GOARCH}/$(NAME) ; \
	done; \
	echo "Done."

darwin:
	@for GOARCH in ${ARCHS}; do \
		echo "Building $(NAME) for darwin $${GOARCH} ..." ; \
		mkdir -p ${TARGET}/$(NAME)-darwin-$${GOARCH} ; \
		GOOS=darwin GOARCH=$${GOARCH} GO111MODULE=on CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -gcflags=${GCFLAGS} -asmflags=${ASMFLAGS} -o ${TARGET}/$(NAME)-darwin-$${GOARCH}/$(NAME) ; \
	done; \
	echo "Done."


.PHONY: test