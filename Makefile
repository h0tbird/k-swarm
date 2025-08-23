# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Local registry host:port
REGISTRY_HOST = $(shell $(CTLPTL) get cluster kind-dev -o template --template '{{.status.localRegistryHosting.host}}' || echo 'localhost:5000')
# Image URL to use for pushing image targets
PUSH_IMG ?= ${REGISTRY_HOST}/k-swarm:latest
# Image URL to use for pulling image targets
PULL_IMG ?= dev-registry:5000/k-swarm:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CRD directory
CRD_DIR ?= config/crd

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Platform list for multi-arch buildx
PLATFORMS ?= linux/arm64,linux/amd64

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

#------------------------------------------------------------------------------
##@ General
#------------------------------------------------------------------------------

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

#------------------------------------------------------------------------------
##@ Development
#------------------------------------------------------------------------------

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=$(CRD_DIR)/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet setup-envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test $$(go list ./... | grep -v /e2e) -coverprofile cover.out

# TODO(user): To use a different vendor for e2e tests, modify the setup under 'tests/e2e'.
# The default setup assumes Kind is pre-installed and builds/loads the Manager Docker image locally.
# CertManager is installed by default; skip with:
# - CERT_MANAGER_INSTALL_SKIP=true
KIND_CLUSTER ?= k-swarm-test-e2e

.PHONY: setup-test-e2e
setup-test-e2e: ## Set up a Kind cluster for e2e tests if it does not exist
	@command -v $(KIND) >/dev/null 2>&1 || { \
		echo "Kind is not installed. Please install Kind manually."; \
		exit 1; \
	}
	@case "$$($(KIND) get clusters)" in \
		*"$(KIND_CLUSTER)"*) \
			echo "Kind cluster '$(KIND_CLUSTER)' already exists. Skipping creation." ;; \
		*) \
			echo "Creating Kind cluster '$(KIND_CLUSTER)'..."; \
			$(KIND) create cluster --name $(KIND_CLUSTER) ;; \
	esac

.PHONY: test-e2e
test-e2e: setup-test-e2e manifests generate fmt vet ## Run the e2e tests. Expected an isolated environment using Kind.
	KIND_CLUSTER=$(KIND_CLUSTER) go test ./test/e2e/ -v -ginkgo.v
	$(MAKE) cleanup-test-e2e

.PHONY: cleanup-test-e2e
cleanup-test-e2e: ## Tear down the Kind cluster used for e2e tests
	@$(KIND) delete cluster --name $(KIND_CLUSTER)

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter & yamllint
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

#------------------------------------------------------------------------------
##@ Build
#------------------------------------------------------------------------------

.PHONY: build-devel
build-devel: generate ## Build a manager binary without optimizations and inlining for Alpine musl linux/ARCH.
	GO111MODULE=on go build -gcflags "-N -l" -o bin/manager cmd/main.go

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: swarmctl
swarmctl: ## Build swarmctl binary.
	$(GORELEASER) build --snapshot --single-target --clean -o bin/swarmctl

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

# If you wish to build the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	- $(CONTAINER_TOOL) buildx create --name k-swarm-builder
	$(CONTAINER_TOOL) buildx use k-swarm-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${PUSH_IMG} .
	- $(CONTAINER_TOOL) buildx rm k-swarm-builder

.PHONY: release
release: ## Create a new release
	git checkout -B ${BRANCH}
	git push -u origin ${BRANCH}
	git tag -a ${TAG} -m "Release ${TAG}"
	$(GORELEASER) release --clean
	PUSH_IMG=ghcr.io/h0tbird/k-swarm:$$(jq -r '.tag' dist/metadata.json) make docker-buildx

#------------------------------------------------------------------------------
##@ Deployment
#------------------------------------------------------------------------------

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	@[ -d "$(CRD_DIR)" ] && $(KUSTOMIZE) build "$(CRD_DIR)" | $(KUBECTL) apply -f - || { \
		echo "No CRDs found in $(CRD_DIR) (skipping install)"; \
	}

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	@[ -d "$(CRD_DIR)" ] && $(KUSTOMIZE) build "$(CRD_DIR)" | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f - || { \
		echo "No CRDs found in $(CRD_DIR) (skipping uninstall)"; \
	}

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: overlay
overlay: kustomize ## Render a kustomize overlay to stdout.
	@ cd config/manager && $(KUSTOMIZE) edit set image controller=${PULL_IMG}
	@ $(KUSTOMIZE) build config/overlays/$(OVERLAY)

#------------------------------------------------------------------------------
##@ Tilt / Kind
#------------------------------------------------------------------------------

.PHONY: kind-create
kind-create: ctlptl  ## Create a kind cluster with a local registry.
	$(CTLPTL) apply -f hack/dev-cluster.yaml

.PHONY: tilt-up
tilt-up: kind-create ## Start kind and tilt.
	tilt up -- --flags '--leader-elect=false --enable-informer=true --enable-worker=true --worker-bind-address=:8082 --informer-bind-address=:8083 --informer-url=http://k-swarm-informer.k-swarm-system --informer-poll-interval=10s --worker-request-interval=2s --zap-devel'

.PHONY: kind-delete
kind-delete: ctlptl ## Delete the local development cluster.
	$(CTLPTL) delete --cascade true -f hack/dev-cluster.yaml

#------------------------------------------------------------------------------
##@ Build Dependencies
#------------------------------------------------------------------------------

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KIND ?= kind
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
CTLPTL ?= $(LOCALBIN)/ctlptl
GORELEASER ?= $(LOCALBIN)/goreleaser

## Tool Versions
KUSTOMIZE_VERSION ?= v5.6.0          # https://github.com/kubernetes-sigs/kustomize/releases
CONTROLLER_TOOLS_VERSION ?= v0.17.2  # https://github.com/kubernetes-sigs/controller-tools/releases
CTLPTL_VERSION ?= v0.8.39            # https://github.com/tilt-dev/ctlptl/releases
GORELEASER_VERSION ?= v2.7.0         # https://github.com/goreleaser/goreleaser/releases
#ENVTEST_VERSION is the version of controller-runtime release branch to fetch the envtest setup script (i.e. release-0.20)
ENVTEST_VERSION ?= $(shell go list -m -f "{{ .Version }}" sigs.k8s.io/controller-runtime | awk -F'[v.]' '{printf "release-%d.%d", $$2, $$3}')
#ENVTEST_K8S_VERSION is the version of Kubernetes to use for setting up ENVTEST binaries (i.e. 1.31)
ENVTEST_K8S_VERSION ?= $(shell go list -m -f "{{ .Version }}" k8s.io/api | awk -F'[v.]' '{printf "1.%d", $$3}')
GOLANGCI_LINT_VERSION ?= v2.1.6

## Tool install scripts
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"

.PHONY: tooling
tooling: kustomize controller-gen envtest ctlptl goreleaser ## Install all the tooling.

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(LOCALBIN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
	@ test -s $(CONTROLLER_GEN) && $(CONTROLLER_GEN) --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: setup-envtest
setup-envtest: envtest ## Download the binaries required for ENVTEST in the local bin directory.
	@echo "Setting up envtest binaries for Kubernetes version $(ENVTEST_K8S_VERSION)..."
	@$(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path || { \
		echo "Error: Failed to set up envtest binaries for version $(ENVTEST_K8S_VERSION)."; \
		exit 1; \
	}

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: ctlptl
ctlptl: $(LOCALBIN) ## Download ctlptl locally if necessary. If wrong version is installed, it will be overwritten.
	@ test -s $(CTLPTL) && $(CTLPTL) version | grep -q $(CTLPTL_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/tilt-dev/ctlptl/cmd/ctlptl@$(CTLPTL_VERSION)

.PHONY: goreleaser
goreleaser: ## Download goreleaser locally if necessary. If wrong version is installed, it will be overwritten.
	@ test -s $(GORELEASER) && $(GORELEASER) --version | grep -q $(GORELEASER_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION)

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef
