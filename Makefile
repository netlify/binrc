.PONY: all build deps image lint release test

all: test build ## Run the tests and build the binary.

build: ## Build the binary.
	@go build -ldflags "-X github.com/netlify/binrc/cmd.Version=`git rev-parse HEAD`"

deps: ## Install dependencies.
	@go get -u github.com/golang/lint/golint
	@go get -u github.com/golang/dep && dep ensure -update

help: ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

image: ## Build the Docker image.
	@docker build .

lint: ## Run golint to ensure the code follows Go styleguide.
	@golint -set_exit_status `go list ./... | grep -v /vendor/`

release: build ## Build the linux binary and upload it to GitHub releases as a tarball
	@rm -rf releases/*
	@mkdir -p releases/binrc_${TAG}_linux_amd64
	@mv binrc releases/binrc_${TAG}_linux_amd64/binrc_${TAG}_linux_amd64
	@cp LICENSE releases/binrc_${TAG}_linux_amd64/
	@tar -C releases -czvf releases/binrc_${TAG}_Linux-64bit.tar.gz binrc_${TAG}_linux_amd64

test: lint ## Run tests.
	@go test -v `go list ./... | grep -v /vendor/`
