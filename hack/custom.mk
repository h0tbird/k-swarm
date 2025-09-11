# Local registry host:port
REGISTRY_HOST = $(shell $(CTLPTL) get cluster kind-dev -o template --template '{{.status.localRegistryHosting.host}}' || echo 'localhost:5000')
# Image URL to use for pushing image targets
PUSH_IMG ?= ${REGISTRY_HOST}/k-swarm:latest
# Image URL to use for pulling image targets
PULL_IMG ?= dev-registry:5000/k-swarm:latest

# CRD directory
CRD_DIR ?= config/crd

# Platform list for multi-arch buildx
PLATFORMS ?= linux/arm64,linux/amd64

.PHONY: build-devel
build-devel: generate ## Build a manager binary without optimizations and inlining for Alpine musl linux/ARCH.
	GO111MODULE=on go build -gcflags "-N -l" -o bin/manager cmd/main.go

.PHONY: swarmctl
swarmctl: ## Build swarmctl binary.
	$(GORELEASER) build --snapshot --single-target --clean -o bin/swarmctl

.PHONY: release
release: ## Create a new release
	git checkout -B ${BRANCH}
	git push -u origin ${BRANCH}
	git tag -a ${TAG} -m "Release ${TAG}"
	$(GORELEASER) release --clean
	PUSH_IMG=ghcr.io/h0tbird/k-swarm:$$(jq -r '.tag' dist/metadata.json) make docker-buildx

.PHONY: overlay
overlay: kustomize ## Render a kustomize overlay to stdout.
	@ cd config/manager && $(KUSTOMIZE) edit set image controller=${PULL_IMG}
	@ $(KUSTOMIZE) build config/overlays/$(OVERLAY)

##@ Tilt / Kind

.PHONY: kind-create
kind-create: ctlptl  ## Create a kind cluster with a local registry.
	$(CTLPTL) apply -f hack/dev-cluster.yaml

.PHONY: tilt-up
tilt-up: kind-create ## Start kind and tilt.
	tilt up -- --flags '--leader-elect=false --enable-informer=true --enable-worker=true --worker-bind-address=:8082 --informer-bind-address=:8083 --informer-url=http://k-swarm-informer.k-swarm-system --informer-poll-interval=10s --worker-request-interval=2s --zap-devel'

.PHONY: kind-delete
kind-delete: ctlptl ## Delete the local development cluster.
	$(CTLPTL) delete --cascade true -f hack/dev-cluster.yaml

CTLPTL ?= $(LOCALBIN)/ctlptl
GORELEASER ?= $(LOCALBIN)/goreleaser

CTLPTL_VERSION ?= v0.8.39            # https://github.com/tilt-dev/ctlptl/releases
GORELEASER_VERSION ?= v2.7.0         # https://github.com/goreleaser/goreleaser/releases

.PHONY: goreleaser
goreleaser: ## Download goreleaser locally if necessary. If wrong version is installed, it will be overwritten.
	@ test -s $(GORELEASER) && $(GORELEASER) --version | grep -q $(GORELEASER_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION)
