# k-swarm
`k-swarm` is used for deploying a series of k8s services that are capable of identifying and communicating with one another, thus establishing a network of synthetic traffic. This interconnected traffic among various workloads provides a foundational platform for a range of laboratory experiments, including the testing and validation of diverse service mesh configurations at scale.

## Install

```
HOMEBREW_GITHUB_API_TOKEN=${GITHUB_TOKEN} brew install octoroot/tap/swarmctl
```

## Developing

Download all the `Makefile` tooling to `./bin/`:
```
make tooling
```

Bring up a local dev environment:
```
make tilt-up
```

Release
```
git checkout -b release-0.1
git push -u origin release-0.1
git tag -a v0.1.0 -m "Release v0.1.0"
make release
```

## Performance Profiling and Benchmarking
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
