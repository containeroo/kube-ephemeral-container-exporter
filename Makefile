# Setting SHELL to bash allows bash commands to be executed by recipes.
SHELL := /usr/bin/env bash -o pipefail
.SHELLFLAGS := -ec

##@ Paths

LOCALBIN ?= $(CURDIR)/bin

$(LOCALBIN):
	mkdir -p $(LOCALBIN)

##@ Tool binaries

ENVTEST := $(LOCALBIN)/setup-envtest
GOLANGCI_LINT := $(LOCALBIN)/golangci-lint
YAMLFMT := $(LOCALBIN)/yamlfmt
GINKGO := $(LOCALBIN)/ginkgo
KIND := $(LOCALBIN)/kind

##@ Tool versions

ENVTEST_K8S_VERSION ?= 1.30.0
# renovate: datasource=github-releases depName=kubernetes-sigs/controller-runtime
ENVTEST_VERSION ?= release-0.18
# renovate: datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.12.1
# renovate: datasource=github-releases depName=google/yamlfmt
YAMLFMT_VERSION ?= v0.21.0
# renovate: datasource=github-releases depName=onsi/ginkgo
GINKGO_VERSION ?= v2.28.3
# renovate: datasource=github-releases depName=kubernetes-sigs/kind
KIND_VERSION ?= 0.31.0

## Detect platform for Kind binary
UNAME_S := $(shell uname -s | tr '[:upper:]' '[:lower:]')
KIND_BINARY := kind-$(UNAME_S)-amd64

# Tagging
VERSION_PREFIX ?= v
LATEST_TAG := $(shell git tag --list '$(VERSION_PREFIX)*' --sort=-v:refname | head -n 1)
CURRENT_VERSION := $(if $(LATEST_TAG),$(patsubst $(VERSION_PREFIX)%,%,$(LATEST_TAG)),0.0.0)
NEXT_PATCH := $(shell echo "$(CURRENT_VERSION)" | awk -F. '{printf "%d.%d.%d", $$1, $$2, $$3 + 1}')
NEXT_MINOR := $(shell echo "$(CURRENT_VERSION)" | awk -F. '{printf "%d.%d.0", $$1, $$2 + 1}')
NEXT_MAJOR := $(shell echo "$(CURRENT_VERSION)" | awk -F. '{printf "%d.0.0", $$1 + 1}')


##@ General

.PHONY: all
all: build

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} \
	/^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } \
	/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Tagging

.PHONY: tag tag-patch tag-minor tag-major push-tags
tag-patch: ## Create the next patch tag locally.
	@git tag -a "$(VERSION_PREFIX)$(NEXT_PATCH)" -m "Release $(VERSION_PREFIX)$(NEXT_PATCH)"
	@echo "Created tag $(VERSION_PREFIX)$(NEXT_PATCH)"

tag-minor: ## Create the next minor tag locally.
	@git tag -a "$(VERSION_PREFIX)$(NEXT_MINOR)" -m "Release $(VERSION_PREFIX)$(NEXT_MINOR)"
	@echo "Created tag $(VERSION_PREFIX)$(NEXT_MINOR)"

tag-major: ## Create the next major tag locally.
	@git tag -a "$(VERSION_PREFIX)$(NEXT_MAJOR)" -m "Release $(VERSION_PREFIX)$(NEXT_MAJOR)"
	@echo "Created tag $(VERSION_PREFIX)$(NEXT_MAJOR)"

push-tags: ## Push commits and tags to origin.
	@git push --follow-tags

tag: ## Show latest tag.
	@echo "Latest version: $(if $(LATEST_TAG),$(LATEST_TAG),none (next: $(VERSION_PREFIX)$(NEXT_PATCH)))"


##@ Development

.PHONY: download
download: ## Download go packages and list them.
	go mod download
	go list -m all

.PHONY: verify-deps
verify-deps: ## Verify go.mod and go.sum are tidy
	go mod tidy
	git diff --exit-code go.mod go.sum

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet envtest ## Run unit tests.
	KUBEBUILDER_ASSETS="$$( $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path )" \
	go test -coverprofile=cover.out -covermode=atomic -count=1 -parallel=4 -timeout=5m ./internal/...

.PHONY: lint
lint: golangci-lint ## Run golangci-lint.
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint and apply fixes.
	$(GOLANGCI_LINT) run --fix

.PHONY: check-header
check-header: ## Verify that all *.go files have the boilerplate header.
	@missing_files=0; \
	for file in $$(find . -type f -name '*.go'); do \
		package_line=$$(grep -n -m1 '^package[[:space:]]' "$$file" | cut -d: -f1); \
		if [ -z "$$package_line" ]; then \
			echo "Could not find package declaration in $$file"; \
			missing_files=$$((missing_files + 1)); \
			continue; \
		fi; \
		header_tmp=$$(mktemp); \
		sed -n "1,$$((package_line - 1))p" "$$file" | sed '$${/^$$/d;}' > "$$header_tmp"; \
		if ! diff -u hack/boilerplate.go.txt "$$header_tmp" > /dev/null; then \
			echo "Missing or incorrect header in $$file"; \
			missing_files=$$((missing_files + 1)); \
		fi; \
		rm -f "$$header_tmp"; \
	done; \
	if [ $$missing_files -ne 0 ]; then \
		echo "ERROR: Some files are missing the required boilerplate header."; \
		exit 1; \
	fi; \
	echo "All files have the correct boilerplate header."

.PHONY: check-header-fix
check-header-fix: ## Fix missing or incorrect headers in all *.go files.
	@for file in $$(find . -type f -name '*.go'); do \
		package_line=$$(grep -n -m1 '^package[[:space:]]' "$$file" | cut -d: -f1); \
		if [ -z "$$package_line" ]; then \
			echo "Skipping $$file: could not find package declaration"; \
			continue; \
		fi; \
		header_tmp=$$(mktemp); \
		sed -n "1,$$((package_line - 1))p" "$$file" | sed '$${/^$$/d;}' > "$$header_tmp"; \
		if ! diff -u hack/boilerplate.go.txt "$$header_tmp" > /dev/null; then \
			echo "Fixing header in $$file"; \
			body_tmp=$$(mktemp); \
			sed -n "$${package_line},\$$p" "$$file" > "$$body_tmp"; \
			{ \
				cat hack/boilerplate.go.txt; \
				echo; \
				cat "$$body_tmp"; \
			} > "$$file"; \
			rm -f "$$body_tmp"; \
		fi; \
		rm -f "$$header_tmp"; \
	done; \
	echo "Headers have been fixed for all *.go files."

##@ Build

.PHONY: build
build: fmt vet ## Build binary.
	go build -o bin/kube-ephemeral-container-exporter ./cmd/main.go

.PHONY: run
run: fmt vet ## Run locally.
	go run ./cmd/main.go --leader-elect=false

.PHONY: kustomize
kustomize: yamlfmt ## Render kustomize manifests into a single file.
	kustomize build deploy/kubernetes/ > deploy/kubernetes/kube-ephemeral-container-exporter.yaml
	$(YAMLFMT) deploy/kubernetes/kube-ephemeral-container-exporter.yaml

##@ Kind / E2E

.PHONY: kind
kind: $(KIND) ## Create a Kind cluster.
	@echo "Setting up Kind cluster..."
	@$(KIND) create cluster --name kube-ephemeral-container-exporter-test --wait 60s
	@kubectl cluster-info

.PHONY: delete-kind
delete-kind: ## Delete the Kind cluster.
	@echo "Deleting Kind cluster..."
	@$(KIND) delete cluster --name kube-ephemeral-container-exporter-test
	@echo "Kind cluster teardown complete."

.PHONY: e2e
e2e: ginkgo ## Run all e2e tests sequentially.
	@echo "Running e2e tests with Ginkgo..."
	PATH=$(LOCALBIN):$$PATH USE_EXISTING_CLUSTER=true \
	$(GINKGO) --procs=1 --timeout=30m --tags=e2e -v --focus='$(FOCUS)' ./test/e2e/...

.PHONY: e2e-generic
e2e-generic: ## Run generic e2e tests.
	@$(MAKE) e2e FOCUS="Pods generic"

.PHONY: e2e-namespaced
e2e-namespaced: ## Run namespaced-mode e2e tests.
	@$(MAKE) e2e FOCUS="Namespaced mode"

##@ Dependencies

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.

$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.

$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: yamlfmt
yamlfmt: $(YAMLFMT) ## Download yamlfmt locally if necessary.

$(YAMLFMT): $(LOCALBIN)
	$(call go-install-tool,$(YAMLFMT),github.com/google/yamlfmt/cmd/yamlfmt,$(YAMLFMT_VERSION))

.PHONY: ginkgo
ginkgo: $(GINKGO) ## Download ginkgo locally if necessary.

$(GINKGO): $(LOCALBIN)
	$(call go-install-tool,$(GINKGO),github.com/onsi/ginkgo/v2/ginkgo,$(GINKGO_VERSION))

$(KIND): $(LOCALBIN)
	@if [ ! -f "$(KIND)" ]; then \
		echo "Downloading Kind $(KIND_VERSION) for $(UNAME_S)..."; \
		curl -fsSL -o "$(KIND)" "https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/$(KIND_BINARY)"; \
		chmod +x "$(KIND)"; \
		echo "Kind $(KIND_VERSION) installed at $(KIND)."; \
	fi

.PHONY: tools
tools: ginkgo envtest golangci-lint yamlfmt kind ## Install all tools.

# go-install-tool installs a tool versioned by suffix and symlinks the stable name.
# $1 - target path with binary name
# $2 - package url
# $3 - version
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
	set -e; \
	package=$(2)@$(3); \
	echo "Downloading $$package"; \
	rm -f "$(1)" || true; \
	GOBIN=$(LOCALBIN) go install $$package; \
	mv "$(1)" "$(1)-$(3)"; \
}; \
ln -sf "$(1)-$(3)" "$(1)"
endef
