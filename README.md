# swarm
A k8s service swarm

Build and publish a multi-arch docker image:
```
PUSH_IMG=h0tbird/swarm make docker-buildx
```

Deploy the informer:
```
kustomize build ./config/informer | k --context little-sunshine-1-admin apply -f -
```

Deploy the workers:
```
cd config/worker
kustomize edit set namespace foo-1
kustomize build . | k --context little-sunshine-1-admin apply -f -
```

Check the workload endpoints:
```
istioctl --context little-sunshine-1-admin -n foo-1 pc endpoint deploy/controller-manager | grep worker
```
