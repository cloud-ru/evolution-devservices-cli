BINARY        := eds
PKG           := github.com/cloud-ru/evolution-devservices-cli
BUILD_DIR     := bin
DIST_DIR      := dist

# Target platforms: Linux + macOS + Windows (developers locally + CI runners).
DEFAULT_OSES  := darwin linux windows
DEFAULT_ARCHS := amd64 arm64

OSES          ?= $(DEFAULT_OSES)
ARCHS         ?= $(DEFAULT_ARCHS)
VERSION       ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GOFLAGS       ?=
CGO_ENABLED   ?= 0

# Must come after VERSION: LDFLAGS uses immediate (:=) expansion, so it has
# to be defined once VERSION is already resolved, or it silently bakes in
# an empty version string.
LDFLAGS       := -s -w -X main.version=$(VERSION)

# ---- targets --------------------------------------------------------------

.PHONY: all
all: lint test build

.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) go build $(GOFLAGS) \
		-trimpath -ldflags '$(LDFLAGS)' \
		-o $(BUILD_DIR)/$(BINARY) ./cmd/eds
	@echo "Built $(BUILD_DIR)/$(BINARY) (version=$(VERSION))"

# Helper target - invoke recursively for each os/arch combination.
# Usage: make build-one GOOS=darwin GOARCH=arm64
.PHONY: build-one
build-one:
	@mkdir -p $(DIST_DIR)
	$(eval EXE := $(if $(filter windows,$(GOOS)),.exe,))
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build $(GOFLAGS) -trimpath -ldflags '$(LDFLAGS)' \
		-o $(DIST_DIR)/$(BINARY)-$(GOOS)-$(GOARCH)$(EXE) ./cmd/eds
	@echo "  ✔ $(DIST_DIR)/$(BINARY)-$(GOOS)-$(GOARCH)$(EXE)"

.PHONY: build-all
build-all:
	@status=0; \
	for os in $(OSES); do \
	  for arch in $(ARCHS); do \
	    if ! $(MAKE) --no-print-directory build-one GOOS=$$os GOARCH=$$arch; then \
	      status=1; \
	    fi; \
	  done; \
	done; \
	exit $$status
	@echo
	@echo "Artifacts in $(DIST_DIR)/:"
	@ls -1 $(DIST_DIR)

.PHONY: release
release: clean build-all
	@cd $(DIST_DIR) && shasum -a 256 $(BINARY)-* > checksums.txt
	@echo
	@echo "Checksums:"
	@cat $(DIST_DIR)/checksums.txt

# --- publish to S3-compatible object storage --------------------------------
#
# Requires the `aws` CLI configured with the right endpoint and credentials,
# for example:
#
#   export AWS_ENDPOINT_URL=https://storage.cloud.ru
#   export AWS_ACCESS_KEY_ID=...
#   export AWS_SECRET_ACCESS_KEY=...
#
# Layout produced in the bucket:
#
#   s3://BUCKET/PREFIX/VERSION/eds-<os>-<arch>[.exe]
#   s3://BUCKET/PREFIX/latest                  # contains VERSION
#   s3://BUCKET/PREFIX/install.sh              # copy of scripts/install.sh
#   s3://BUCKET/PREFIX/install.ps1             # copy of scripts/install.ps1
#
# Usage:
#   make upload BUCKET=my-bucket PREFIX=evolution-devservices-cli VERSION=v0.2.0
#   make upload-latest BUCKET=my-bucket PREFIX=evolution-devservices-cli VERSION=v0.2.0

BUCKET        ?=
PREFIX        ?= evolution-devservices-cli
S3_OPTS       := $(if $(AWS_ENDPOINT_URL),--endpoint-url $(AWS_ENDPOINT_URL),)

.PHONY: upload
upload: build-all
	@if [[ -z "$(BUCKET)" ]]; then \
	  echo "BUCKET is required, e.g. make upload BUCKET=my-bucket VERSION=v0.2.0"; \
	  exit 1; \
	fi
	@echo "Uploading version=$(VERSION) to s3://$(BUCKET)/$(PREFIX)/"
	@for f in $(DIST_DIR)/$(BINARY)-*; do \
	  name=$$(basename $$f); \
	  aws $(S3_OPTS) s3 cp $$f s3://$(BUCKET)/$(PREFIX)/$(VERSION)/$$name \
	    --acl public-read >/dev/null && \
	    echo "  ✔ s3://$(BUCKET)/$(PREFIX)/$(VERSION)/$$name"; \
	done
	@aws $(S3_OPTS) s3 cp scripts/install.sh s3://$(BUCKET)/$(PREFIX)/install.sh \
	  --acl public-read >/dev/null && \
	  echo "  ✔ s3://$(BUCKET)/$(PREFIX)/install.sh"
	@aws $(S3_OPTS) s3 cp scripts/install.ps1 s3://$(BUCKET)/$(PREFIX)/install.ps1 \
	  --acl public-read >/dev/null && \
	  echo "  ✔ s3://$(BUCKET)/$(PREFIX)/install.ps1"
	@printf "%s\n" "$(VERSION)" > $(DIST_DIR)/.latest && \
	  aws $(S3_OPTS) s3 cp $(DIST_DIR)/.latest s3://$(BUCKET)/$(PREFIX)/latest \
	    --acl public-read --content-type "text/plain" >/dev/null && \
	  rm -f $(DIST_DIR)/.latest && \
	  echo "  ✔ s3://$(BUCKET)/$(PREFIX)/latest -> $(VERSION)"

.PHONY: upload-latest
upload-latest:
	@if [[ -z "$(BUCKET)" ]]; then \
	  echo "BUCKET is required"; exit 1; \
	fi
	@printf "%s\n" "$(VERSION)" > $(DIST_DIR)/.latest && \
	  aws $(S3_OPTS) s3 cp $(DIST_DIR)/.latest s3://$(BUCKET)/$(PREFIX)/latest \
	    --acl public-read --content-type "text/plain" >/dev/null && \
	  rm -f $(DIST_DIR)/.latest && \
	  echo "  ✔ latest -> $(VERSION)"

.PHONY: goreleaser-build
goreleaser-build: clean
	goreleaser build --snapshot --clean

.PHONY: goreleaser-release
goreleaser-release: clean
	goreleaser release --clean

.PHONY: install
install:
	CGO_ENABLED=$(CGO_ENABLED) go install $(GOFLAGS) \
		-trimpath -ldflags '$(LDFLAGS)' $(PKG)/cmd/eds

.PHONY: test
test:
	go test $(GOFLAGS) ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR) $(DIST_DIR)

openapi-generator:
	openapi-generator-cli generate -i openapi-public.yaml -g go -o ./internal/workflow_client -c .openapi-generator.yaml

.PHONY: lint
lint:
	golangci-lint run --fix

.PHONY: gen-install
gen-install:
	binst gen -o scripts/install.sh --config binstaller.yml
