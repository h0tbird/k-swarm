# swarm
A k8s service swarm

Build and publish a multi-arch docker image:
```
PUSH_IMG=h0tbird/swarm make docker-buildx
```

CPU Profiling
```
swarmctl --context 'kind-foo-*' --cpu-profile worker 1:30
go tool pprof --http localhost:3000 cpu.prof
```

Check the workload endpoints:
```
istioctl --context little-sunshine-1-admin -n foo-1 pc endpoint deploy/controller-manager | grep worker
```
