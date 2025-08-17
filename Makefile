BUILD_DIR := build
PROVIDER_NAME := interfaces-v1alpha1
TF_PROVIDER_NAME := terraform-provider-${PROVIDER_NAME}
TERRAFORMRC := "${HOME}/.terraformrc"
TF_RC_DEV_KEY := "nokia-eda/${PROVIDER_NAME}"


# Detect host OS/ARCH at makefile parse time (used by targets below)
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Darwin)
OS := darwin
else
OS := linux
endif

ifeq ($(UNAME_M),x86_64)
ARCH := amd64
else
	ifeq ($(UNAME_M),aarch64)
	ARCH := arm64
	else
		ifeq ($(UNAME_M),arm64)
		ARCH := arm64
		else
		ARCH := $(UNAME_M)
		endif
	endif
endif

# tfplugindocs release to download when requested
TFPLUGINDOCS_VERSION := 0.22.0
TFPLUGINDOCS_NAME := tfplugindocs_${TFPLUGINDOCS_VERSION}_${OS}_${ARCH}
TFPLUGINDOCS_ZIP := ${TFPLUGINDOCS_NAME}.zip
TFPLUGINDOCS_URL := https://github.com/hashicorp/terraform-plugin-docs/releases/download/v${TFPLUGINDOCS_VERSION}/${TFPLUGINDOCS_ZIP}

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: all
all: build

##@ General
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts"
	@rm -rf ${BUILD_DIR}


.PHONY: fmt
fmt: ## Run go fmt against and terraform fmt.
	go fmt ./...
	terraform fmt -recursive examples

.PHONY: gen-docs
gen-docs: tfplugindocs ## Generate docs using local tfplugindocs binary.
	@echo "Generating documentation"
	@./${BUILD_DIR}/tfplugindocs generate --provider-dir . --provider-name ${PROVIDER_NAME}

.PHONY: vet
vet: ## Run go vet against code.
	@go mod tidy
	go vet ./...

.PHONY: build-dir
build-dir: ## Ensure the ./build directory exists
	@mkdir -p ${BUILD_DIR}

.PHONY: build
build: build-dir fmt vet ## Build the terraform provider
	@echo "Building ${TF_PROVIDER_NAME}"
	go build -ldflags="-s -w" -o ${BUILD_DIR}/${TF_PROVIDER_NAME} main.go

.PHONY: tfplugindocs
tfplugindocs: build-dir ## Download tfplugindocs binary for host OS/ARCH and unpack into build/
	@if [ ! -f "${BUILD_DIR}/tfplugindocs" ]; then \
		echo "Detected OS=$(OS) ARCH=$(ARCH)"; \
		echo "Downloading tfplugindocs..."; \
		./scripts/tfplugindocs.sh "${TFPLUGINDOCS_URL}" "${TFPLUGINDOCS_ZIP}" "${BUILD_DIR}"; \
	else \
		echo "tfplugindocs already exists at ${BUILD_DIR}/tfplugindocs, skipping download"; \
	fi

##@ Deployment

.PHONY: install
install: build ## Install the terraform provider
	@echo "Installing ${TF_PROVIDER_NAME} dev override"
	@export TERRAFORMRC=${TERRAFORMRC} KEY=${TF_RC_DEV_KEY} BUILD_PATH="$$(realpath ${BUILD_DIR})"; \
	[ -f "$${TERRAFORMRC}" ] || { echo "Creating $${TERRAFORMRC}"; echo 'provider_installation {\n  dev_overrides {\n  }\n  direct {}\n}' > "$${TERRAFORMRC}"; }; \
	if grep -q "$${KEY}" "$${TERRAFORMRC}"; then \
		echo "Key $${KEY} already present in $${TERRAFORMRC}"; \
	else \
		awk -v key="\"$${KEY}\"" -v value="\"$${BUILD_PATH}\"" '\
		BEGIN { in_block=0 } \
		/dev_overrides[[:space:]]*{/ { in_block=1 } \
		in_block && /}/ { print "      " key " = " value; in_block=0 } \
		{ print }' "$${TERRAFORMRC}" > "$${TERRAFORMRC}.tmp" && mv "$${TERRAFORMRC}.tmp" "$${TERRAFORMRC}"; \
		echo "Added key $${KEY} to dev_overrides block in $${TERRAFORMRC} pointing to $${BUILD_PATH}"; \
	fi

.PHONY: uninstall
uninstall: ## Uninstall the terraform provider
	@echo "Uninstalling ${TF_PROVIDER_NAME} dev override"
	@rm -rf ${BUILD_DIR}
	@export TERRAFORMRC=${TERRAFORMRC} KEY=${TF_RC_DEV_KEY}; \
	awk -v key="$${KEY}" '$$0 ~ key { next } { print }' "$${TERRAFORMRC}" > "$${TERRAFORMRC}.tmp" && mv "$${TERRAFORMRC}.tmp" "$${TERRAFORMRC}"; \
	echo "Removed key $${KEY} from dev_overrides block in $${TERRAFORMRC};"

.PHONY: gen-docs
gen-docs: ## Generate documentation
	@echo "Generating documentation"
	@./${BUILD_DIR}/tfplugindocs generate --provider-dir . --provider-name ${PROVIDER_NAME}
