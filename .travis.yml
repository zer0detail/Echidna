language: go

go:
  - 1.x

before_install:
  - curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.24.0

script:
  - make lint test