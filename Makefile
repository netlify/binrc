.PHONY: all build deps image lint release generate test

help: ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: deps generate test build ## Run the tests and build the binary.

generate:
	@go generate

build: ## Build the binary.
	@go build -ldflags "-X github.com/netlify/binrc/cmd.Version=`git rev-parse HEAD`"

deps: ## Install dependencies.
	@GO111MODULE=off go get -u github.com/myitcv/gobin
	@go mod download

image: ## Build the Docker image.
	@docker build .

lint: ## Run golint to ensure the code follows Go styleguide.
	@gobin -m -run github.com/golang/lint/golint -set_exit_status ./...

publish: release upload ## Build and upload a release to GitHub releases.

release: ## Build the linux binary and prepares the release as a tarball.
	$(eval TAG := $(shell git describe --tags --abbrev=0 | cut -c 2-))
	@GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/netlify/binrc/cmd.Version=`git rev-parse HEAD`"
	@rm -rf releases/*
	@mkdir -p releases/binrc_$(TAG)_linux_amd64
	@mv binrc releases/binrc_$(TAG)_linux_amd64/binrc_$(TAG)_linux_amd64
	@cp LICENSE releases/binrc_$(TAG)_linux_amd64/
	@tar -C releases -czvf releases/binrc_$(TAG)_Linux-64bit.tar.gz binrc_$(TAG)_linux_amd64

test: lint ## Run tests.
	@go test -v ./...

upload: ## Upload release to GitHub releases.
	$(eval TAG := $(shell git describe --tags --abbrev=0 | cut -c 2-))
	@hub release create -a releases/binrc_$(TAG)_Linux-64bit.tar.gz v$(TAG)
