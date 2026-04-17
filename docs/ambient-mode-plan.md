# Plan: Add Istio Ambient mode to k-swarm

## TL;DR

Add a `--dataplane-mode {sidecar|ambient}` flag (default `ambient`) and a
`--waypoint-name` flag to `swarmctl`. Conditionally render templates so that
in ambient mode:

- Namespace gets `istio.io/dataplane-mode: ambient` instead of `istio-injection: enabled`.
- Sidecar resource annotations are dropped from pod templates.
- A Gateway-API `Gateway` (waypoint) is deployed per namespace and pods are
  labeled with `istio.io/use-waypoint: <name>` so L7 `AuthorizationPolicy`
  keeps working through the waypoint.
- `DestinationRule` is omitted (only effective with sidecars).
- The worker's external ingress migrates from Istio `Gateway` + `VirtualService`
  to Gateway API `Gateway` + `HTTPRoute`.
- `PeerAuthentication` (STRICT mTLS) is dropped (ambient is mTLS by default
  via ztunnel).

User decisions:

- Selectable via flag (keep sidecar working).
- Yes, deploy a waypoint to preserve L7 features.
- Migrate external ingress to Gateway API (`HTTPRoute`).

## Workflow conventions

- Work on a new branch `ambient-mode` (off `main`).
- Small, focused commits — one logical change each. Suggested commit
  boundaries are listed inside each phase below.
- Open a **draft PR** to `h0tbird/k-swarm` after the first commit so progress
  is visible early in the GitHub UI; push subsequent commits to the same
  branch.
- Mark the PR ready for review only after Phase 5 verification passes.

## Phases & Steps

### Phase 0 — Bootstrap

*Commits*: (0) "docs: add ambient-mode plan"; create branch `ambient-mode`,
add this file, push, open draft PR.

### Phase 1 — Plumbing: flags & template variables

*Commits*: (1.a) "swarmctl: add --dataplane-mode and --waypoint-name flags";
(1.b) "swarmctl: thread DataplaneMode/WaypointName into template data".

1. Add `--dataplane-mode` (string, default `ambient`, values `sidecar|ambient`)
   and `--waypoint-name` (string, default `waypoint`) as persistent flags on
   both `manifestGenerateCmd` and `manifestInstallCmd` in
   `cmd/swarmctl/cmd/cmd.go` `init()`.
2. Add completion funcs `dataplaneModeCompletion` and `waypointNameCompletion`,
   and validators `dataplaneModeIsValid` (only allow `sidecar`/`ambient`) and
   `waypointNameIsValid`. Wire them in `validateFlags`.
3. In `cmd/swarmctl/pkg/swarmctl/swarmctl.go`, read the new flags in all four
   entry points: `GenerateInformer`, `GenerateWorker`, `InstallInformer`,
   `InstallWorker`. Add `DataplaneMode string` and `WaypointName string` to
   each anonymous struct passed to `tmpl.Execute`.
4. Sanity check: when `--istio-revision` is set together with
   `--dataplane-mode=ambient`, still render the `istio.io/rev` label
   (revision tag selects the control plane regardless of mode).

### Phase 2 — Worker template (`cmd/swarmctl/assets/worker.goyaml`)

*Commits*: (2.a) "worker template: ambient namespace label + drop sidecar pod
annotations"; (2.b) "worker template: gate PeerAuthentication and
DestinationRule on sidecar mode"; (2.c) "worker template: ambient ingress via
Gateway API + per-namespace waypoint".

5. Namespace labels: add an `else if eq .DataplaneMode "ambient"` branch that
   emits `istio.io/dataplane-mode: ambient` (keep `istio-injection: enabled`
   as the sidecar/default branch).
6. Deployment pod template:
   - Wrap the four `sidecar.istio.io/proxy*` annotations in
     `{{- if ne .DataplaneMode "ambient" }} … {{- end }}`.
   - **(corrected during validation)** The `istio.io/use-waypoint`
     label belongs on the **`Service`**, not on the pod template — the
     waypoint Gateway carries the default `istio.io/waypoint-for: service`,
     so service-addressed traffic only transits the waypoint when the
     Service (or namespace) is labeled. Add
     `{{- if eq .DataplaneMode "ambient" }}istio.io/use-waypoint: {{ .WaypointName }}{{- end }}`
     to the worker `Service` metadata labels.
7. `DestinationRule worker`: wrap the entire document in
   `{{- if ne .DataplaneMode "ambient" }} … {{- end }}`.
8. `PeerAuthentication worker`: wrap in `{{- if ne .DataplaneMode "ambient" }}`
   (ambient default is mTLS via ztunnel).
9. `AuthorizationPolicy worker`: **(corrected during validation)** in
   ambient mode the policy must attach via `targetRefs: [{kind: Service,
   group: "", name: worker}]` so the waypoint enforces it at L7. With
   `selector` only, ztunnel enforces at L4 and rejects HTTP-rule policies
   (resulting in `503`). Sidecar mode keeps the existing `selector`-based
   form.
10. Replace Istio `Gateway` + `VirtualService` documents with conditional
    branches:
    - Sidecar branch: existing Istio `Gateway` + `VirtualService` (unchanged).
    - Ambient branch: Gateway API `Gateway` (`gatewayClassName: istio`,
      listener `https/443/HTTPS` with `tls.mode: Terminate` and
      `certificateRefs: [{name: worker-{{ .Namespace }}}]`) and an `HTTPRoute`
      attached to it routing to `Service/worker:80`.
    - Keep the cert-manager `Certificate` document in both branches.
11. Append a new conditional document at the end of the worker template
    (ambient only): a Gateway API waypoint `Gateway` named
    `{{ .WaypointName }}` with `gatewayClassName: istio-waypoint` and
    listener `mesh/15008/HBONE`. This is the per-namespace waypoint
    referenced by the `istio.io/use-waypoint` label.

### Phase 3 — Informer template (`cmd/swarmctl/assets/informer.goyaml`)

*Commits*: (3.a) "informer template: ambient namespace label + drop sidecar
pod annotations"; (3.b) "informer template: gate PeerAuthentication on
sidecar + add waypoint Gateway".

12. Namespace labels: same change as step 5.
13. Deployment pod template: drop sidecar annotations as in step 6. The
    `istio.io/use-waypoint` label goes on the `Service informer`, not on
    the pod template (same correction as step 6).
14. `PeerAuthentication informer`: same wrap as step 8.
15. `AuthorizationPolicy informer`: same correction as step 9 — use
    `targetRefs: [{kind: Service, name: informer}]` in ambient mode.
16. Append the same waypoint `Gateway` document (ambient only) for namespace
    `informer`.

### Phase 4 — Examples & docs (light touch)

*Commits*: (4) "swarmctl: add --dataplane-mode ambient examples".

17. Update example strings in `cmd/swarmctl/pkg/swarmctl/swarmctl.go`
    (`GenerateInformerExample`, `GenerateWorkerExample`, `InstallInformerExample`,
    `InstallWorkerExample`) to include `--dataplane-mode ambient` examples.
18. Update `README.md` brief section if it documents install commands.
    *Optional; skip if README has no flag list.*

### Phase 5 — Verification

*No commits unless fixes are needed.* Build with
`go build -o bin/swarmctl ./cmd/swarmctl`.

19. `go build ./...` from `/workspaces/k-swarm` — must compile.
20. Render check (sidecar parity, no diff besides scaffolding):
    `bin/swarmctl manifest generate informer` and
    `bin/swarmctl manifest generate worker 1:1` produce byte-identical output
    to the previous version (modulo intentional changes — diff and review).
21. Render check (ambient):
    - `bin/swarmctl manifest generate informer --dataplane-mode ambient` →
      namespace has `istio.io/dataplane-mode: ambient`, no
      `sidecar.istio.io/*` annotations, no `PeerAuthentication`, deployment
      carries `istio.io/use-waypoint: waypoint`, a waypoint `Gateway` is
      present.
    - `bin/swarmctl manifest generate worker 1:1 --dataplane-mode ambient` →
      same, plus Gateway API `Gateway` + `HTTPRoute` instead of Istio
      `Gateway` + `VirtualService`, no `DestinationRule`.
22. Validate against the kind clusters (see Test environment).
23. Run existing `go test ./...` / `make test` if defined.

## Test environment

- Two kind clusters available: `kind-pasta-1`, `kind-pasta-2`.
- The `install` subcommand accepts `--context` as a regex against kubeconfig
  contexts; `--context 'kind-pasta-.*'` targets both clusters in one
  invocation.
- Container image: workloads pull `ghcr.io/h0tbird/k-swarm:main` (or
  `--image-tag`); no local image build required for testing template changes.

Validation commands:

- Sidecar parity: diff `bin/swarmctl manifest generate {informer,worker 1:1}`
  against the same command from `main` branch — expect no diff.
- Ambient render: same with `--dataplane-mode ambient`; grep for
  `istio.io/dataplane-mode: ambient`, `istio.io/use-waypoint: waypoint`,
  `kind: HTTPRoute`, `gatewayClassName: istio-waypoint`; assert absence of
  `sidecar.istio.io/proxy`, `kind: PeerAuthentication`, `kind: DestinationRule`,
  `kind: VirtualService`.
- Cluster install:
  - `bin/swarmctl manifest install informer --context 'kind-pasta-.*' --dataplane-mode ambient --yes`
  - `bin/swarmctl manifest install worker 1:2 --context 'kind-pasta-.*' --dataplane-mode ambient --yes`
- In-cluster checks per context: namespace labels carry
  `istio.io/dataplane-mode=ambient`; pods are 1/1 (no sidecar);
  waypoint/ingress `Gateway` + `HTTPRoute` exist; no `PeerAuthentication` /
  `DestinationRule` / `VirtualService`; `istioctl ztunnel-config workload`
  lists workloads as HBONE; worker logs show successful informer polls.
- Cleanup:
  `kubectl --context <ctx> delete ns informer service-1 service-2 --ignore-not-found`
  for each context.

## Validation results (kind-pasta-1, kind-pasta-2)

Performed against `kind-pasta-1` and `kind-pasta-2`, both with Istio ambient
(istio-cni + ztunnel + waypoint controller) and Gateway API CRDs installed.

- `bin/swarmctl manifest install informer --context 'kind-pasta-.*' --image-tag main --yes`
  and the equivalent `worker 1:2` invocation apply cleanly. `--image-tag main`
  is required because the default `Version=0.0.0` would resolve to a
  non-existent `ghcr.io/h0tbird/k-swarm:v0.0.0` image.
- `kubectl get pods` in `informer`, `service-1`, `service-2` shows workload
  pods at `1/1 Running` (no sidecar). Per-namespace `waypoint-*` pods are
  also `1/1 Running`.
- `kubectl get gateway` shows `worker` (ingress, class `istio`) and
  `waypoint` (class `istio-waypoint`) with `PROGRAMMED=True`.
- `istioctl ztunnel-config service` shows the `informer` and `worker`
  Services with `WAYPOINT=waypoint` after the Service-label correction.
- Worker logs show successful `polling service list` against
  `http://informer.informer/services` and successful peer
  `sending a request {"service": "worker.service-1:80" ...}` after the
  `targetRefs` AuthorizationPolicy correction.

## Relevant files

- `cmd/swarmctl/assets/worker.goyaml` — namespace label, pod
  annotations/labels, conditional `DestinationRule`/`PeerAuthentication`,
  `Gateway`/`VirtualService` → Gateway API, append waypoint `Gateway`.
- `cmd/swarmctl/assets/informer.goyaml` — namespace label, pod
  annotations/labels, conditional `PeerAuthentication`, append waypoint
  `Gateway`.
- `cmd/swarmctl/cmd/cmd.go` — register `--dataplane-mode`, `--waypoint-name`
  flags + completion + validators on `manifestGenerateCmd` /
  `manifestInstallCmd`; extend `validateFlags`.
- `cmd/swarmctl/pkg/swarmctl/swarmctl.go` — read flags in `GenerateInformer`,
  `GenerateWorker`, `InstallInformer`, `InstallWorker`; add `DataplaneMode` /
  `WaypointName` fields to each template-data struct; refresh example strings.

## Decisions / Scope

- Sidecar mode is **not** removed — chosen via `--dataplane-mode`, default
  `sidecar` for backwards compatibility.
- Drop `PeerAuthentication` and `DestinationRule` only in the ambient branch
  (they don't apply to ztunnel L4 traffic).
- Migrate ingress to Gateway API only in the ambient branch; sidecar branch
  keeps the existing Istio `Gateway` + `VirtualService` to avoid behavior
  change.
- Waypoint is per-namespace (one per `service-N` and one in `informer`),
  `gatewayClassName: istio-waypoint`, name defaults to `waypoint`,
  configurable via `--waypoint-name`.
- `Telemetry` template (`cmd/swarmctl/assets/telemetry.goyaml`) is **out of
  scope** — Telemetry CR works in both modes.
- Controller code under `internal/controller/`, `pkg/worker/`,
  `pkg/informer/` — no changes; the switch is purely deployment-template
  scope.

## Further considerations

1. Waypoint `gatewayClassName`: Istio ambient currently uses
   `istio-waypoint`. Recommendation: hardcode `istio-waypoint` (matches
   upstream defaults). Alternative: expose `--waypoint-class` flag.
   *Recommend hardcoded.*
2. Ingress TLS termination on the worker Gateway-API listener: today the
   Istio `Gateway` uses `tls.mode: SIMPLE` with `credentialName`. Gateway API
   equivalent is `Terminate` with `certificateRefs`. Recommendation: keep the
   cert-manager `Certificate` unchanged and reference the same secret.
3. `--dataplane-mode` defaults to `ambient` (per user preference); pass
   `--dataplane-mode sidecar` to opt back into the sidecar dataplane.
