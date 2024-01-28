# swarm
A k8s service swarm

Build and publish a multi-arch docker image:
```
PUSH_IMG=h0tbird/swarm make docker-buildx
```

## Performance Profiling and Benchmarking
CPU Profiling
```
swarmctl --context 'kind-foo-*' --cpu-profile worker 1:30
go tool pprof --http localhost:3000 cpu.prof
```

Tracing
```
swarmctl --context kind-foo-1 informer --tracing
go tool trace trace.out
```

Benchmarking
```
go test -bench=. -benchmem -memprofile old-mem.prof -cpuprofile old-cpu.prof -benchtime=100x -count=10 ./cmd/swarmctl/pkg/k8sctx | tee old-bench.txt
go test -bench=. -benchmem -memprofile new-mem.prof -cpuprofile new-cpu.prof -benchtime=100x -count=10 ./cmd/swarmctl/pkg/k8sctx | tee new-bench.txt
benchstat old-bench.txt new-bench.txt
```

```
goos: darwin
goarch: arm64
pkg: github.com/octoroot/swarm/cmd/swarmctl/pkg/k8sctx
             │ old-bench.txt │         new-bench.txt         │
             │    sec/op     │   sec/op     vs base          │
ApplyYaml-10     180.1m ± 0%   180.1m ± 0%  ~ (p=0.315 n=10)

             │ old-bench.txt │            new-bench.txt             │
             │     B/op      │     B/op      vs base                │
ApplyYaml-10   115.18Ki ± 0%   38.37Ki ± 1%  -66.69% (p=0.000 n=10)

             │ old-bench.txt │           new-bench.txt            │
             │   allocs/op   │ allocs/op   vs base                │
ApplyYaml-10     1310.0 ± 0%   601.0 ± 0%  -54.12% (p=0.000 n=10)
```
