# swarm
A k8s service swarm

Build and publish a multi-arch docker image:
```
PUSH_IMG=h0tbird/swarm make docker-buildx
```

Deploy the informer:
```
for CLUSTER in {1..2}; do
  kustomize build ./config/informer | k --context little-sunshine-${CLUSTER}-admin apply -f -
done
```

Deploy some workers:
```
cd config/worker

for SERVICE in {1..5}; do
  kustomize edit set namespace foo-${SERVICE}
  for CLUSTER in {1..2}; do
    kustomize build . | k --context little-sunshine-${CLUSTER}-admin apply -f -
  done
done
```

Check the workload endpoints:
```
istioctl --context little-sunshine-1-admin -n foo-1 pc endpoint deploy/controller-manager | grep worker
```
