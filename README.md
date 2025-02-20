[![Build and publish to ghcr.io](https://github.com/h0tbird/k-swarm/actions/workflows/docker-build-and-publish.yml/badge.svg)](https://github.com/h0tbird/k-swarm/actions/workflows/docker-build-and-publish.yml)
[![Cleanup ghcr.io images](https://github.com/h0tbird/k-swarm/actions/workflows/cleanup-ghcr-images.yml/badge.svg)](https://github.com/h0tbird/k-swarm/actions/workflows/cleanup-ghcr-images.yml)
[![Dependabot Updates](https://github.com/h0tbird/k-swarm/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/h0tbird/k-swarm/actions/workflows/dependabot/dependabot-updates)

# k-swarm
`k-swarm` is used for deploying a series of k8s services that are capable of identifying and communicating with one another, thus establishing a network of synthetic traffic. This interconnected traffic among various workloads provides a foundational platform for a range of laboratory experiments, including the testing and validation of diverse service mesh configurations at scale.

### Install

```
HOMEBREW_GITHUB_API_TOKEN=${GITHUB_TOKEN} brew install h0tbird/tap/swarmctl
```

### Usage

Install the `informer` with two replicas to all `kind` clusters:
```
swarmctl i --context 'kind-*' --replicas 2
```

Install services 1 to 5 with 2 `workers` each to all `kind` clusters:
```
swarmctl w --context 'kind-*' 1:5 --replicas 2
```

Enable telemetry for `service-1` on all `kind` clusters:
```
swarmctl w t --context 'kind-*' 1:1 on
```

## Developing

Download all the `Makefile` tooling to `./bin/`:
```
make tooling
```

Upgrade the telemetry CRD:
```
k apply view-last-applied crd telemetries.telemetry.istio.io > cmd/swarmctl/assets/crds.yaml
```

Bring up a local dev environment:
```
make tilt-up
```

Create a new release:
```
make release BRANCH='release-0.1' TAG='v0.1.0'
```

### Performance Profiling and Benchmarking
CPU Profiling
```
swarmctl w --context 'kind-foo-*' 1:10 --cpu-profile
go tool pprof --http localhost:3000 cpu.prof
```

Memory Profiling
```
swarmctl w --context 'kind-foo-*' 1:10 --mem-profile
go tool pprof --http localhost:3000 mem.prof
```

Tracing
```
swarmctl i --contextkind-foo-1 --tracing
go tool trace trace.out
```

Benchmarking
```
go test -bench=. -benchmem -memprofile 0-mem.prof -cpuprofile 0-cpu.prof -benchtime=100x -count=10 ./cmd/swarmctl/pkg/k8sctx | tee 0-bench.txt
go test -bench=. -benchmem -memprofile 1-mem.prof -cpuprofile 1-cpu.prof -benchtime=100x -count=10 ./cmd/swarmctl/pkg/k8sctx | tee 1-bench.txt
go test -bench=. -benchmem -memprofile 2-mem.prof -cpuprofile 2-cpu.prof -benchtime=100x -count=10 ./cmd/swarmctl/pkg/k8sctx | tee 2-bench.txt
benchstat 0-bench.txt 1-bench.txt 2-bench.txt
```
