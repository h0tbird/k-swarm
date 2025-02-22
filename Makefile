
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
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

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
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.64.5
golangci-lint:
	@[ -f $(GOLANGCI_LINT) ] || { \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell dirname $(GOLANGCI_LINT)) $(GOLANGCI_LINT_VERSION) ;\
	}

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
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${PULL_IMG}
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
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
CTLPTL ?= $(LOCALBIN)/ctlptl
GORELEASER ?= $(LOCALBIN)/goreleaser

## Tool Versions
KUSTOMIZE_VERSION ?= v5.5.0          # https://github.com/kubernetes-sigs/kustomize/releases
CONTROLLER_TOOLS_VERSION ?= v0.17.2  # https://github.com/kubernetes-sigs/controller-tools/releases
ENVTEST_K8S_VERSION = 1.32.0         # https://github.com/kubernetes-sigs/controller-runtime/tree/main/tools/setup-envtest
CTLPTL_VERSION ?= v0.8.39            # https://github.com/tilt-dev/ctlptl/releases
GORELEASER_VERSION ?= v2.7.0         # https://github.com/goreleaser/goreleaser/releases

## Tool install scripts
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"

.PHONY: tooling
tooling: kustomize controller-gen envtest ctlptl goreleaser ## Install all the tooling.

.PHONY: kustomize
kustomize: $(LOCALBIN) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
	@ test -s $(KUSTOMIZE) && $(KUSTOMIZE) version | grep -q $(KUSTOMIZE_VERSION) || \
	{ rm -rf $(KUSTOMIZE); curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); } \
	> /dev/null

.PHONY: controller-gen
controller-gen: $(LOCALBIN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
	@ test -s $(CONTROLLER_GEN) && $(CONTROLLER_GEN) --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(LOCALBIN) ## Download envtest-setup locally if necessary.
	@ test -s $(ENVTEST) || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@$(ENVTEST_K8S_VERSION)

.PHONY: ctlptl
ctlptl: $(LOCALBIN) ## Download ctlptl locally if necessary. If wrong version is installed, it will be overwritten.
	@ test -s $(CTLPTL) && $(CTLPTL) version | grep -q $(CTLPTL_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/tilt-dev/ctlptl/cmd/ctlptl@$(CTLPTL_VERSION)

.PHONY: goreleaser
goreleaser: ## Download goreleaser locally if necessary. If wrong version is installed, it will be overwritten.
	@ test -s $(GORELEASER) && $(GORELEASER) --version | grep -q $(GORELEASER_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION)
