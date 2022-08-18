.PHONY: all build deps image lint release generate test

TAG = $(shell git describe --tags --abbrev=0 | cut -c 2-)

help: ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)


## $(1) binary-name
## $(2) entry file, path to main.go
## $(3) goos
## $(4) goarch
## $(5) extension (like ".exe")
define build_binary 
	@echo "Building $(1) for $(3)/$(4) $(2)"
	@GOOS=darwin GOARCH=$(arch) go build \
		-ldflags " \
			-X github.com/netlify/binrc/info.sha=`git rev-parse HEAD` \
			-X github.com/netlify/binrc/info.distro=$(3) \
			-X github.com/netlify/binrc/info.arch=$(4) \
			-X github.com/netlify/binrc/cmd.Version=`git rev-parse HEAD`" \
		-o builds/$(3)-$(4)-${TAG}/$(1)$(5) $(2)

	@echo "Built: builds/$(3)-$(4)-${TAG}/$(1)$(5)"
endef

## $(1) binary-name
## $(2) goos
## $(3) goarch
## $(4) artifact ending (like ".zip" or ".tar.gz")
## $(5) extension (like ".exe")
define package	
	@cp LICENSE releases/${TAG}/
	@tar -czf releases/${TAG}/$(1)-$(2)-$(3)$(4) -C builds/$(2)-$(3)-${TAG} $(1)$(5)

	@sha256sum releases/${TAG}/$(1)-$(2)-$(3)$(4) >> releases/${TAG}/checksums.txt
endef

all: deps generate test lint build ## Run the tests and build the binary.

generate:
	@go generate

deps: ## Install dependencies.
	@echo "Installing dependencies"
	@go get -u github.com/myitcv/gobin
	@go get -u golang.org/x/lint/golint
	@go mod verify
	@go mod tidy
	@go mod download

build: ## Build the client binary
	@rm -rf build
	$(call build_binary,binrc_$(TAG),main.go,linux,386)
	$(call build_binary,binrc_$(TAG),main.go,linux,amd64)
	$(call build_binary,binrc_$(TAG),main.go,linux,arm)
	$(call build_binary,binrc_$(TAG),main.go,linux,arm64)
	$(call build_binary,binrc_$(TAG),main.go,darwin,amd64)
	$(call build_binary,binrc_$(TAG),main.go,darwin,arm64)

lint: ## Run golint to ensure the code follows Go styleguide.
	golint -set_exit_status ./...

test: lint ## Run tests.
	@go test -v ./...

clean: ## Clean the build artifacts.
	@mkdir -p releases/${TAG}
	@rm -f releases/${TAG}/*
	@rm -rf pkg-build

package: clean build ## Build a release package for Linux.
	$(call package,binrc_$(TAG),linux,386,.tar.gz)
	$(call package,binrc_$(TAG),linux,amd64,.tar.gz)
	$(call package,binrc_$(TAG),linux,arm,.tar.gz)
	$(call package,binrc_$(TAG),linux,arm64,.tar.gz)
	$(call package,binrc_$(TAG),darwin,amd64,.tar.gz)
	$(call package,binrc_$(TAG),darwin,arm64,.tar.gz)

release_upload: package	
	@echo "Uploading release"

	@hub release create \
		-a releases/${TAG}/binrc_${TAG}-darwin-amd64.tar.gz \
		-a releases/${TAG}/binrc_${TAG}-darwin-arm64.tar.gz \
		-a releases/${TAG}/binrc_${TAG}-linux-386.tar.gz \
		-a releases/${TAG}/binrc_${TAG}-linux-amd64.tar.gz \
		-a releases/${TAG}/binrc_${TAG}-linux-arm.tar.gz \
		-a releases/${TAG}/binrc_${TAG}-linux-arm64.tar.gz \
		-a releases/${TAG}/LICENSE \
		-a releases/${TAG}/checksums.txt v${TAG} \
		-m "Release v${TAG}"
