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
go test -bench=. ./cmd/swarmctl/pkg/k8sctx > old.txt
go test -bench=. ./cmd/swarmctl/pkg/k8sctx > new.txt
benchstat old.txt new.txt
```
