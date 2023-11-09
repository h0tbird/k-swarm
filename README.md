# swarm
A k8s service swarm

Build and publish a multi-arch docker image:
```
PUSH_IMG=h0tbird/swarm make docker-buildx
```

Deploy informer and worker:
```
kustomize build ./config/informer | k --context little-sunshine-1-admin apply -f - 
kustomize build ./config/worker | k --context little-sunshine-1-admin apply -f - 
```