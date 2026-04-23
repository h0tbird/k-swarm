# k-swarm Architecture

This document is the entry point for new contributors who need to understand how
`k-swarm` is structured and how its pieces fit together. It complements the
top-level [README.md](../README.md) (which focuses on usage) and the in-progress
[ambient-mode-plan.md](./ambient-mode-plan.md).

## 1. What problem does k-swarm solve?

`k-swarm` deploys a cooperating set of Kubernetes workloads that automatically
discover each other and exchange synthetic HTTP traffic. The resulting traffic
mesh is used as a foundation for service-mesh experiments — for example,
validating Istio sidecar vs. ambient configurations, multi-cluster failover, or
telemetry pipelines at scale.

The design optimizes for two things:

- **Easy fan-out across many clusters and namespaces**, driven by a single
  command-line tool.
- **Self-organizing traffic**: once deployed, workers discover their peers and
  start talking with no further input.

## 2. High-level component map

`k-swarm` ships **two binaries** built from this repository:

| Binary      | Source                                  | Distribution                | Audience               |
| ----------- | --------------------------------------- | --------------------------- | ---------------------- |
| `swarmctl`  | [cmd/swarmctl/main.go](../cmd/swarmctl/main.go) | Homebrew / Go binary | Humans, CI, automation |
| `manager`   | [cmd/main.go](../cmd/main.go)           | Container image (`ghcr.io/h0tbird/k-swarm`) | Runs inside the cluster |

The `manager` binary is a single image that can run in **two distinct roles**
selected by command-line flags:

- `--enable-informer` — runs the **informer** (one Deployment per cluster)
- `--enable-worker`  — runs the **worker**   (many Deployments per cluster)

```mermaid
flowchart LR
    Operator([Human or CI])
    SC[swarmctl CLI]

    subgraph Cluster[Kubernetes cluster]
        direction TB
        subgraph NSI[namespace informer]
            Inf[informer Deployment]
        end
        subgraph NS1[namespace sidecar-n1]
            W1[worker Deployment]
        end
        subgraph NS2[namespace sidecar-n2]
            W2[worker Deployment]
        end
        subgraph NSN[namespace sidecar-nN]
            WN[worker Deployment]
        end
    end

    Operator --> SC
    SC -->|server-side apply| Inf
    SC -->|server-side apply| W1
    SC -->|server-side apply| W2
    SC -->|server-side apply| WN

    W1 -.->|poll services| Inf
    W2 -.->|poll services| Inf
    WN -.->|poll services| Inf

    W1 ==>|fetch data| W2
    W2 ==>|fetch data| W1
    W2 ==>|fetch data| WN
    WN ==>|fetch data| W2
    W1 ==>|fetch data| WN
    WN ==>|fetch data| W1

    classDef informer fill:#dbeafe,stroke:#1d4ed8,color:#1e3a8a;
    classDef worker fill:#dcfce7,stroke:#15803d,color:#14532d;
    classDef tool fill:#fef3c7,stroke:#b45309,color:#78350f;
    class Inf informer;
    class W1,W2,WN worker;
    class SC tool;
```

## 3. The `swarmctl` CLI

`swarmctl` is a thin Cobra-based CLI whose only job is to **render Go templates
into Kubernetes manifests and server-side apply them** against a fan-out of
contexts selected by a regex.

Key packages:

- [cmd/swarmctl/cmd/cmd.go](../cmd/swarmctl/cmd/cmd.go) — wires the Cobra
  command tree and flags.
- [cmd/swarmctl/pkg/swarmctl/swarmctl.go](../cmd/swarmctl/pkg/swarmctl/swarmctl.go) —
  `Generate*` and `Install*` handlers; renders templates and drives the apply
  loop.
- [cmd/swarmctl/pkg/k8sctx/k8sctx.go](../cmd/swarmctl/pkg/k8sctx/k8sctx.go) —
  per-kubeconfig-context wrapper holding a REST config plus discovery and
  dynamic clients (used for SSA).
- [cmd/swarmctl/assets/](../cmd/swarmctl/assets/) — embedded `*.goyaml`
  templates: `informer.goyaml`, `worker.goyaml`, `telemetry.goyaml` and the
  Istio `crds.yaml` snapshot.

### Command tree

```
swarmctl
├── manifest (m)
│   ├── dump (d)                          # write embedded templates to ~/.swarmctl
│   ├── generate (g)                      # render to stdout
│   │   ├── informer (i)
│   │   │   └── telemetry (t) on|off
│   │   └── worker (w) <start:end>
│   │       └── telemetry (t) on|off
│   └── install (i)                       # render + server-side apply
│       ├── informer (i)
│       │   └── telemetry (t) on|off
│       └── worker (w) <start:end>
│           └── telemetry (t) on|off
```

`install` accepts `--context '<regex>'`; matching kubeconfig contexts are
discovered, the user is prompted (unless `--yes`), and the rendered manifests
are server-side applied to **every** matching cluster.

### Typical flow

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant SC as swarmctl
    participant KC as kubeconfig
    participant API as kube-apiserver

    User->>SC: swarmctl w --context 'kind-*' 1:5 --replicas 2 --dataplane-mode sidecar
    SC->>KC: enumerate contexts, filter by regex
    SC-->>User: list matched contexts, prompt y/N
    User-->>SC: y
    loop for each context
        SC->>SC: render worker.goyaml for i in 1..5 (namespace sidecar-nN)
        SC->>API: server-side apply manifests (dynamic client)
    end
```

The `worker` subcommand takes a numeric range (`<start:end>`); for each `i` it
renders the worker template into namespace `<dataplane-mode>-n<i>` (e.g.
`sidecar-n1`, `ambient-n3`). This is how a single `swarmctl w 1:5` produces
five Deployments / Services across five namespaces.

## 4. The `manager` binary

[cmd/main.go](../cmd/main.go) is a deliberately small entry point:

```mermaid
flowchart TD
    M([main]) --> F[initFlags]
    F --> D{flags}
    D -->|EnableInformer| I[informer.Start]
    D -->|EnableWorker| W[worker.Start]
    I -.-> Wait[wg.Wait]
    W -.-> Wait

    classDef role fill:#dbeafe,stroke:#1d4ed8,color:#1e3a8a;
    class I,W role;
```

The two roles do not share state; they are simply gated by independent
booleans, and both can technically run in the same process (tests do this).

## 5. The informer

Source: [pkg/informer/informer.go](../pkg/informer/informer.go) and the
controller in [internal/controller/service_controller.go](../internal/controller/service_controller.go).

There is **one informer Deployment per cluster**, in the `informer` namespace.
Internally it runs two cooperating components inside a single
`controller-runtime` manager:

1. A **Kubebuilder controller** (`ServiceReconciler`) that watches
   `core/v1/Service` objects labeled `app=k-swarm`.
2. A **Gin HTTP server** (`Informer` runnable) that exposes `GET /services`.

They are stitched together by an unbuffered `chan []string`:

```mermaid
flowchart LR
    subgraph InformerPod[Informer pod]
        direction LR
        K[(Kubernetes API)]
        R[ServiceReconciler]
        C{{commChan}}
        G[Informer runnable]
        H[/services HTTP endpoint/]

        K -->|watch app=k-swarm| R
        R -->|host ns port| C
        C --> G
        G --> H
    end

    Worker[worker pods] -->|HTTP poll| H

    classDef ctrl fill:#dbeafe,stroke:#1d4ed8,color:#1e3a8a;
    classDef http fill:#fef3c7,stroke:#b45309,color:#78350f;
    classDef worker fill:#dcfce7,stroke:#15803d,color:#14532d;
    class R,G ctrl;
    class H http;
    class Worker worker;
```

Notable details:

- The reconciler filters with a `predicate.NewPredicateFuncs` that only admits
  objects bearing `app=k-swarm`, so reconcile is noisy only on relevant
  Services.
- On every reconcile it `List()`s **all** matching Services and rebuilds the
  full set; entries are formatted as `<name>.<namespace>:<port>` for the
  Service port named `http`.
- The HTTP server is `endless`-based so the process can hot-reload without
  dropping connections.
- The endpoint is intentionally trivial (no auth, no pagination) because it
  lives entirely behind cluster-internal networking.

## 6. The worker

Source: [pkg/worker/worker.go](../pkg/worker/worker.go).

A worker pod is **simultaneously a client and a server**:

- **Server** (`server`): a Gin handler at `GET /data` that returns a small JSON
  blob describing the pod (`CLUSTER_NAME`, `POD_NAME`, `POD_NAMESPACE`,
  `POD_IP`, `NODE_NAME`, all from the downward API).
- **Client** (`client`): periodically polls the informer for the current peer
  list, then in a tight loop issues `GET /data` against every peer, sleeping
  `--worker-request-interval` between requests.

```mermaid
sequenceDiagram
    autonumber
    participant W as Worker pod, client side
    participant I as Informer /services
    participant P1 as Peer worker 1 /data
    participant P2 as Peer worker 2 /data

    loop every informer-poll-interval, default 10s
        W->>I: GET /services
        I-->>W: services list as JSON
    end

    loop forever, over current serviceList
        W->>P1: GET /data
        P1-->>W: 200 with clusterName, podName, ...
        Note over W: sleep worker-request-interval, default 2s
        W->>P2: GET /data
        P2-->>W: 200 with pod metadata
    end
```

The polling goroutine and the request goroutine share the package-level
`serviceList` slice; the polling goroutine atomically replaces it after each
successful fetch. This is intentional: a worker that briefly cannot reach the
informer keeps using the last known peer set.

Workers are deployed **once per namespace**, with multiple replicas inside each
namespace. A typical lab might have:

- `informer/`               — 1 Deployment, N replicas (HA via leader election)
- `sidecar-n1/` … `sidecar-n5/` — one worker Deployment each, R replicas

This namespace-per-service shape is what makes the synthetic mesh useful for
mesh experiments: each namespace can carry different Istio
revisions/dataplane modes/telemetry knobs, all driven by `swarmctl` flags.

## 7. End-to-end: what happens when an operator runs `swarmctl`

```mermaid
sequenceDiagram
    autonumber
    actor Op as Operator
    participant SC as swarmctl
    participant API as kube-apiserver
    participant CTRL as ServiceReconciler in informer pod
    participant SRV as Informer HTTP in informer pod
    participant W as Worker pods

    Op->>SC: swarmctl i --context kind-dev --replicas 1 --dataplane-mode sidecar
    SC->>API: SSA Namespace, RBAC, Deployment, Service for informer
    API-->>CTRL: pod starts, manager and runnable boot
    Op->>SC: swarmctl w --context kind-dev 1:3 --dataplane-mode sidecar
    SC->>API: SSA Namespace plus Deployment and Service for sidecar-n1..n3
    API-->>CTRL: Service add events with label app=k-swarm
    CTRL->>SRV: commChan receives new service list
    W->>SRV: GET /services
    SRV-->>W: peer list
    W->>W: fan out GET /data to peers
```

## 8. Where to look next

- **Add a CLI flag / template knob** → start in
  [cmd/swarmctl/cmd/cmd.go](../cmd/swarmctl/cmd/cmd.go) (register the flag),
  then plumb it through the `Generate*` handler in
  [cmd/swarmctl/pkg/swarmctl/swarmctl.go](../cmd/swarmctl/pkg/swarmctl/swarmctl.go),
  and finally consume it in the matching template under
  [cmd/swarmctl/assets/](../cmd/swarmctl/assets/).
- **Change discovery semantics** → edit the predicate / list logic in
  [internal/controller/service_controller.go](../internal/controller/service_controller.go).
- **Change the synthetic traffic pattern** → edit `client()` in
  [pkg/worker/worker.go](../pkg/worker/worker.go).
- **Add a new HTTP route** to the informer → register it in
  `Informer.Start()` in [pkg/informer/informer.go](../pkg/informer/informer.go).
- **Local end-to-end loop** → `make tilt-up` ([Tiltfile](../Tiltfile)).
- **Ambient-mode roadmap** → [docs/ambient-mode-plan.md](./ambient-mode-plan.md).

## 9. Glossary

- **informer** — cluster-scoped discovery service that lists all swarm
  workers via a Kubebuilder controller and serves them at `GET /services`.
- **worker**   — namespace-scoped HTTP service that polls the informer and
  fans out `GET /data` requests to all discovered peers.
- **manager**  — the Go binary that, depending on flags, runs as informer,
  worker, or both. Shipped as a single container image.
- **swarmctl** — operator-facing CLI that renders embedded templates and
  server-side applies them across a regex-selected fan-out of kube contexts.
