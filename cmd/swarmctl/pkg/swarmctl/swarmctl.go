package swarmctl

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"bufio"
	stdctx "context"
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"

	// Community
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"

	// Local
	"github.com/h0tbird/k-swarm/cmd/swarmctl/pkg/k8sctx"
	"github.com/h0tbird/k-swarm/cmd/swarmctl/pkg/profiling"
	"github.com/h0tbird/k-swarm/cmd/swarmctl/pkg/util"
)

//-----------------------------------------------------------------------------
// Globals
//-----------------------------------------------------------------------------

var (
	Assets   embed.FS
	Contexts = map[string]*k8sctx.Context{}
)

//-----------------------------------------------------------------------------
// Root
//-----------------------------------------------------------------------------

func Root(cmd *cobra.Command, args []string) error {

	// Return early if the command is a completion command
	if cmd.CalledAs() == "__complete" || strings.Contains(cmd.CommandPath(), "completion") {
		return nil
	}

	// Handle profiling
	if err := profiling.Start(); err != nil {
		return fmt.Errorf("error starting profiling: %w", err)
	}

	// Return
	return nil
}

//-----------------------------------------------------------------------------
// Dump writes every embedded asset template to ~/.swarmctl
//-----------------------------------------------------------------------------

func Dump(cmd *cobra.Command, args []string) error {

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// Create ~/.swarmctl
	if err := os.MkdirAll(util.SwarmDir, 0755); err != nil {
		return fmt.Errorf("error creating ~/.swarmctl: %w", err)
	}

	// List every file under the embedded assets/ directory.
	entries, err := Assets.ReadDir("assets")
	if err != nil {
		return fmt.Errorf("error reading embedded assets directory: %w", err)
	}

	// Write each file to ~/.swarmctl/<name>.
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()

		fileData, err := Assets.ReadFile("assets/" + name)
		if err != nil {
			return fmt.Errorf("error reading %s from embedded FS: %w", name, err)
		}

		if err := os.WriteFile(util.SwarmDir+"/"+name, fileData, 0644); err != nil {
			return fmt.Errorf("error writing ~/.swarmctl/%s: %w", name, err)
		}

		cmd.Printf("Successfully wrote ~/.swarmctl/%s\n", name)
	}

	return nil
}

func DumpExample() string {
	return `
  # Dump every embedded template to ~/.swarmctl
  swarmctl dump

  # Same using the command alias
  swarmctl d
  `
}

//-----------------------------------------------------------------------------
// Install
//-----------------------------------------------------------------------------

func Install(cmd *cobra.Command, args []string) error {

	// Get the flags
	ctxRegex, _ := cmd.Flags().GetString("context")
	assumeYes, _ := cmd.Flags().GetBool("yes")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Run the root PersistentPreRunE
	if err := cmd.Root().PersistentPreRunE(cmd, args); err != nil {
		return err
	}

	// Get the contexts that match the regex
	matches, err := k8sctx.Filter(ctxRegex)
	if err != nil {
		return err
	}

	// In dry-run, skip client init and confirmation.
	// Use nil entries so loops still run.
	if dryRun {
		for _, match := range matches {
			Contexts[match] = nil
		}
		return nil
	}

	// Print
	cmd.Println("\nMatched contexts:")

	// For every match
	for _, match := range matches {

		// Print the match
		cmd.Printf("  - %s\n", match)

		// Create the context
		c, err := k8sctx.New(match)
		if err != nil {
			return err
		}

		// Store the config
		Contexts[match] = c
	}

	// A chance to cancel
	if !assumeYes {

		// Ask the user
		cmd.Print("\nProceed? (y/N) ")
		reader := bufio.NewReader(os.Stdin)

		// Read the answer
		answer, err := util.YesOrNo(cmd, reader)
		if err != nil {
			return fmt.Errorf("error reading user input: %w", err)
		}

		// Return early if the answer is no
		if answer == "" || answer == "n" || answer == "no" {
			cmd.SetErrPrefix("aborted:")
			return errors.New("by user")
		}
	}

	// Return
	return nil
}

//-----------------------------------------------------------------------------
// InstallInformer
//-----------------------------------------------------------------------------

func InstallInformer(cmd *cobra.Command, args []string) error {

	// Get the flags
	replicas, _ := cmd.Flags().GetInt("replicas")
	nodeSelector, _ := cmd.Flags().GetString("node-selector")
	imageTag, _ := cmd.Flags().GetString("image-tag")
	istioRevision, _ := cmd.Flags().GetString("istio-revision")
	dataplaneMode, _ := cmd.Flags().GetString("dataplane-mode")
	waypointName, _ := cmd.Flags().GetString("waypoint-name")
	ingressMode, _ := cmd.Flags().GetString("ingress-mode")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// Parse the mode-specific template
	tmpl, err := util.ParseTemplate(Assets, "informer-"+dataplaneMode)
	if err != nil {
		return err
	}

	// Loop through all contexts
	for name, context := range Contexts {

		// Print the context (skipped in dry-run mode to keep stdout pure YAML)
		if !dryRun {
			fmt.Printf("\n%s\n", name)
		}

		// Render the template
		docs, err := util.RenderTemplate(tmpl, struct {
			Replicas      int
			NodeSelector  string
			Version       string
			ImageTag      string
			IstioRevision string
			DataplaneMode string
			WaypointName  string
			IngressMode   string
		}{
			Replicas:      replicas,
			NodeSelector:  nodeSelector,
			Version:       cmd.Root().Version,
			ImageTag:      imageTag,
			IstioRevision: istioRevision,
			DataplaneMode: dataplaneMode,
			WaypointName:  waypointName,
			IngressMode:   ingressMode,
		})
		if err != nil {
			return err
		}

		// Loop through all yaml documents
		for _, doc := range docs {
			if dryRun {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "---\n%s\n", strings.TrimSpace(doc)); err != nil {
					return err
				}
				continue
			}
			if err := context.ApplyYaml(doc); err != nil {
				fmt.Printf("\nError: %s\n", err)
			}
		}
	}

	// Return
	return nil
}

func InstallInformerExample() string {
	return `
  # Install the informer to the current context
  swarmctl informer

  # Same using the command alias
  swarmctl i

  # Install the informer to a specific context
  swarmctl i --context my-context

  # Install the informer to all contexts that match a regex
  swarmctl i --context 'my-.*'

  # Install the informer to all contexts that match a regex and set the replicas
  swarmctl i --context 'my-.*' --replicas 3

  # Install the informer to all contexts that match a regex and set the node selector
  swarmctl i --context 'my-.*' --node-selector '{key1: value1, key2: value2}'

  # Install the informer to all contexts that match a regex and set the Istio revision
  swarmctl i --context 'my-.*' --istio-revision 1-21-1

  # Install the informer to all contexts that match a regex in Istio ambient mode
  swarmctl i --context 'my-.*' --dataplane-mode ambient

  # Expose the informer Service via the shared istio-system/istio-nsgw gateway.
  swarmctl i --context 'kind-pasta-.*' --dataplane-mode ambient --ingress-mode shared

  # Expose the informer Service via a dedicated Gateway API Gateway+HTTPRoute.
  swarmctl i --context 'kind-pasta-.*' --dataplane-mode ambient --ingress-mode dedicated

  # Render the informer manifests to stdout without applying them or contacting the cluster.
  swarmctl i --dry-run | kubectl diff -f -
  `
}

//-----------------------------------------------------------------------------
// InstallInformerTelemetry
//-----------------------------------------------------------------------------

func InstallInformerTelemetry(cmd *cobra.Command, args []string) error {

	// Get the flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// Parse the template
	tmpl, err := util.ParseTemplate(Assets, "telemetry")
	if err != nil {
		return err
	}

	// Loop through all contexts
	for name, context := range Contexts {

		// Print the context (skipped in dry-run mode to keep stdout pure YAML)
		if !dryRun {
			fmt.Printf("\n%s\n", name)
		}

		// Render the template
		docs, err := util.RenderTemplate(tmpl, struct {
			OnOff     string
			Namespace string
		}{
			OnOff:     args[0],
			Namespace: "informer",
		})
		if err != nil {
			return err
		}

		// Loop through all yaml documents
		for _, doc := range docs {
			if dryRun {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "---\n%s\n", strings.TrimSpace(doc)); err != nil {
					return err
				}
				continue
			}
			if err := context.ApplyYaml(doc); err != nil {
				fmt.Printf("\nError: %s\n", err)
			}
		}
	}

	// Return
	return nil
}

func InstallInformerTelemetryExample() string {
	return `
  # Switch on the informer's telemetry
  swarmctl informer telemetry on

  # Same using command aliases
  swarmctl i t on
  `
}

//-----------------------------------------------------------------------------
// InstallWorker
//-----------------------------------------------------------------------------

func InstallWorker(cmd *cobra.Command, args []string) error {

	// Get the flags
	replicas, _ := cmd.Flags().GetInt("replicas")
	nodeSelector, _ := cmd.Flags().GetString("node-selector")
	imageTag, _ := cmd.Flags().GetString("image-tag")
	istioRevision, _ := cmd.Flags().GetString("istio-revision")
	clusterDomainFlag, _ := cmd.Flags().GetString("cluster-domain")
	dataplaneMode, _ := cmd.Flags().GetString("dataplane-mode")
	waypointName, _ := cmd.Flags().GetString("waypoint-name")
	ingressMode, _ := cmd.Flags().GetString("ingress-mode")
	multiCluster, _ := cmd.Flags().GetBool("multi-cluster")
	logResponses, _ := cmd.Flags().GetBool("log-responses")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// Parse the range
	start, end, err := util.ParseRange(args[0])
	if err != nil {
		return err
	}

	// Parse the mode-specific template
	tmpl, err := util.ParseTemplate(Assets, "worker-"+dataplaneMode)
	if err != nil {
		return err
	}

	// Loop through all contexts
	for name, context := range Contexts {

		// Print the context (skipped in dry-run mode to keep stdout pure YAML)
		if !dryRun {
			fmt.Printf("\n%s\n", name)
		}

		// Determine cluster domain: flag override or auto-detect from CoreDNS.
		// In dry-run mode there is no live client, so default to cluster.local.
		clusterDomain := clusterDomainFlag
		if clusterDomain == "" {
			if dryRun {
				clusterDomain = "cluster.local"
			} else {
				clusterDomain = context.GetClusterDomain(cmd.Context())
			}
		}

		// Derive cluster name by stripping the kind- prefix (no-op for
		// non-kind contexts).
		clusterName := strings.TrimPrefix(name, "kind-")

		// Loop trough all services
		for i := start; i <= end; i++ {

			if !dryRun {
				fmt.Printf("\n")
			}

			namespace := fmt.Sprintf("%s-n%d", dataplaneMode, i)

			// Render the template
			docs, err := util.RenderTemplate(tmpl, struct {
				Replicas      int
				Namespace     string
				NodeSelector  string
				Version       string
				ImageTag      string
				IstioRevision string
				ClusterDomain string
				ClusterName   string
				DataplaneMode string
				WaypointName  string
				IngressMode   string
				MultiCluster  bool
				LogResponses  bool
			}{
				Replicas:      replicas,
				Namespace:     namespace,
				NodeSelector:  nodeSelector,
				Version:       cmd.Root().Version,
				ImageTag:      imageTag,
				IstioRevision: istioRevision,
				ClusterDomain: clusterDomain,
				ClusterName:   clusterName,
				DataplaneMode: dataplaneMode,
				WaypointName:  waypointName,
				IngressMode:   ingressMode,
				MultiCluster:  multiCluster,
				LogResponses:  logResponses,
			})
			if err != nil {
				return err
			}

			// Loop through all yaml documents
			for _, doc := range docs {
				if dryRun {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "---\n%s\n", strings.TrimSpace(doc)); err != nil {
						return err
					}
					continue
				}
				if err := context.ApplyYaml(doc); err != nil {
					fmt.Printf("\nError: %s\n", err)
				}
			}
		}
	}

	// Return
	return nil
}

func InstallWorkerExample() string {
	return `
  # Install the workers 1 to 1 to the current context
  # (namespaces follow <mode>-n<index>, e.g. sidecar-n1)
  swarmctl worker 1:1 --dataplane-mode sidecar

  # Same using the command alias
  swarmctl w 1:1 --dataplane-mode sidecar

  # Install the workers 1 to 1 to a specific context
  swarmctl w 1:1 --dataplane-mode sidecar --context kind-pasta-1

  # Install the workers 1 to 1 to all contexts that match a regex
  swarmctl w 1:1 --dataplane-mode sidecar --context 'kind-pasta-.*'

  # Install the workers 1 to 1 to all contexts that match a regex and set the replicas
  swarmctl w 1:1 --dataplane-mode sidecar --context 'kind-pasta-.*' --replicas 3

  # Install the workers 1 to 1 to all contexts that match a regex and set the node selector
  swarmctl w 1:1 --dataplane-mode sidecar --context 'kind-pasta-.*' --node-selector '{key1: value1, key2: value2}'

  # Install the workers 1 to 1 to all contexts that match a regex and set the Istio revision
  swarmctl w 1:1 --dataplane-mode sidecar --context 'kind-pasta-.*' --istio-revision 1-21-1

  # Install the workers 1 to 1 to all contexts that match a regex in Istio ambient mode
  swarmctl w 1:1 --dataplane-mode ambient --context 'kind-pizza-.*'

  # Expose the peer Service via the shared istio-system/istio-nsgw gateway.
  swarmctl w 1:1 --dataplane-mode sidecar --context 'kind-pasta-.*' --ingress-mode shared

  # Expose the peer Service via a dedicated Gateway API Gateway+HTTPRoute.
  swarmctl w 1:1 --dataplane-mode ambient --context 'kind-pasta-.*' --ingress-mode dedicated

  # Enable cross-cluster failover for ambient-mode workers: labels the peer
  # and waypoint Services with istio.io/global=true and emits a DestinationRule
  # with locality failover by topology.istio.io/cluster (ambient-only).
  swarmctl w 1:1 --dataplane-mode ambient --context 'kind-pasta-.*' --multi-cluster

  # Render the worker manifests to stdout without applying them or contacting the cluster.
  swarmctl w 1:1 --dataplane-mode ambient --dry-run | kubectl diff -f -
  `
}

//-----------------------------------------------------------------------------
// InstallWorkerTelemetry
//-----------------------------------------------------------------------------

func InstallWorkerTelemetry(cmd *cobra.Command, args []string) error {

	// Get the flags
	dataplaneMode, _ := cmd.Flags().GetString("dataplane-mode")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// Parse the range
	start, end, err := util.ParseRange(args[0])
	if err != nil {
		return err
	}

	// Parse the template
	tmpl, err := util.ParseTemplate(Assets, "telemetry")
	if err != nil {
		return err
	}

	// Loop through all contexts
	for name, context := range Contexts {

		// Print the context (skipped in dry-run mode to keep stdout pure YAML)
		if !dryRun {
			fmt.Printf("\n%s\n", name)
		}

		// Loop trough all services
		for i := start; i <= end; i++ {

			if !dryRun {
				fmt.Printf("\n")
			}

			// Render the template
			docs, err := util.RenderTemplate(tmpl, struct {
				OnOff     string
				Namespace string
			}{
				OnOff:     args[1],
				Namespace: fmt.Sprintf("%s-n%d", dataplaneMode, i),
			})
			if err != nil {
				return err
			}

			// Loop through all yaml documents
			for _, doc := range docs {
				if dryRun {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "---\n%s\n", strings.TrimSpace(doc)); err != nil {
						return err
					}
					continue
				}
				if err := context.ApplyYaml(doc); err != nil {
					fmt.Printf("\nError: %s\n", err)
				}
			}
		}
	}

	// Return
	return nil
}

func InstallWorkerTelemetryValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return []string{"1:1"}, cobra.ShellCompDirectiveNoFileComp
	case 1:
		return []string{"on", "off"}, cobra.ShellCompDirectiveNoFileComp
	default:
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}
}

func InstallWorkerTelemetryExample() string {
	return `
  # Switch on the worker's telemetry
  swarmctl worker telemetry 1:1 on

  # Same using command aliases
  swarmctl w t 1:1 on
  `
}

//-----------------------------------------------------------------------------
// Delete removes everything swarmctl has installed in the matching contexts.
// Targets are identified by the label app.kubernetes.io/managed-by=swarmctl,
// which is set by every embedded template:
//   - cluster-scoped: ClusterRole, ClusterRoleBinding
//   - namespaced:     Namespace (cascades all child resources)
//-----------------------------------------------------------------------------

const deleteLabelSelector = "app.kubernetes.io/managed-by=swarmctl"

var (
	deleteNsGVR         = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	deleteClusterScoped = []struct {
		Kind string
		GVR  schema.GroupVersionResource
	}{
		{"ClusterRoleBinding", schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}},
		{"ClusterRole", schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}},
	}
)

// deletePlan captures, for a single context, the swarmctl-managed resources
// discovered on the cluster.
type deletePlan struct {
	ctx        *k8sctx.Context
	namespaces []string
	cluster    map[string][]string // kind -> names
}

func (p *deletePlan) empty() bool {
	if len(p.namespaces) > 0 {
		return false
	}
	for _, names := range p.cluster {
		if len(names) > 0 {
			return false
		}
	}
	return true
}

// discoverDeletePlan lists every swarmctl-managed resource in the given context.
func discoverDeletePlan(ctx stdctx.Context, c *k8sctx.Context) (*deletePlan, error) {
	p := &deletePlan{ctx: c, cluster: map[string][]string{}}

	nss, err := c.ListByLabel(ctx, deleteNsGVR, "", deleteLabelSelector)
	if err != nil {
		return nil, err
	}
	p.namespaces = nss

	for _, t := range deleteClusterScoped {
		names, err := c.ListByLabel(ctx, t.GVR, "", deleteLabelSelector)
		if err != nil {
			return nil, err
		}
		p.cluster[t.Kind] = names
	}
	return p, nil
}

// printDeletePreview prints the discovered targets for a single context.
func printDeletePreview(cmd *cobra.Command, name string, p *deletePlan) {
	cmd.Printf("  - %s\n", name)
	for _, ns := range p.namespaces {
		cmd.Printf("      Namespace/%s\n", ns)
	}
	for _, t := range deleteClusterScoped {
		for _, n := range p.cluster[t.Kind] {
			cmd.Printf("      %s/%s\n", t.Kind, n)
		}
	}
}

// executeDeletePlan deletes cluster-scoped resources first (bindings before
// roles) then namespaces. Per-resource errors are printed and execution
// continues, mirroring the ApplyYaml loop in the install commands.
func executeDeletePlan(ctx stdctx.Context, p *deletePlan) {
	for _, t := range deleteClusterScoped {
		for _, n := range p.cluster[t.Kind] {
			if err := p.ctx.DeleteResource(ctx, t.GVR, t.Kind, "", n); err != nil {
				fmt.Printf("\nError: %s\n", err)
			}
		}
	}
	for _, ns := range p.namespaces {
		if err := p.ctx.DeleteResource(ctx, deleteNsGVR, "Namespace", "", ns); err != nil {
			fmt.Printf("\nError: %s\n", err)
		}
	}
}

// printDeleteDryRun renders the static deletion plan without contacting any cluster.
func printDeleteDryRun(cmd *cobra.Command, matches []string) {
	cmd.Println("\nMatched contexts:")
	for _, name := range matches {
		cmd.Printf("  - %s\n", name)
	}
	cmd.Printf("\nWould delete in each context (label %q):\n", deleteLabelSelector)
	for _, t := range deleteClusterScoped {
		cmd.Printf("  - %s (cluster-scoped)\n", t.Kind)
	}
	cmd.Println("  - Namespace (cascades all child resources)")
}

// confirmDelete prompts the user to confirm a destructive delete.
func confirmDelete(cmd *cobra.Command) error {
	cmd.Print("\nProceed with deletion? (y/N) ")
	reader := bufio.NewReader(os.Stdin)
	answer, err := util.YesOrNo(cmd, reader)
	if err != nil {
		return fmt.Errorf("error reading user input: %w", err)
	}
	if answer == "" || answer == "n" || answer == "no" {
		cmd.SetErrPrefix("aborted:")
		return errors.New("by user")
	}
	return nil
}

func Delete(cmd *cobra.Command, args []string) error {

	// Get the flags
	ctxRegex, _ := cmd.Flags().GetString("context")
	assumeYes, _ := cmd.Flags().GetBool("yes")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// Run the root PersistentPreRunE (profiling, etc.)
	if err := cmd.Root().PersistentPreRunE(cmd, args); err != nil {
		return err
	}

	// Get the contexts that match the regex
	matches, err := k8sctx.Filter(ctxRegex)
	if err != nil {
		return err
	}

	// Dry-run: print the static plan per matched context, no live client
	if dryRun {
		printDeleteDryRun(cmd, matches)
		return nil
	}

	// Build live contexts and discover what's actually there
	plans := make(map[string]*deletePlan, len(matches))
	cmd.Println("\nMatched contexts:")
	for _, name := range matches {
		c, err := k8sctx.New(name)
		if err != nil {
			return err
		}
		p, err := discoverDeletePlan(cmd.Context(), c)
		if err != nil {
			return err
		}
		plans[name] = p
		printDeletePreview(cmd, name, p)
	}

	// Bail out early if nothing to do
	nothing := true
	for _, p := range plans {
		if !p.empty() {
			nothing = false
			break
		}
	}
	if nothing {
		cmd.Println("\nNothing to delete.")
		return nil
	}

	// A chance to cancel
	if !assumeYes {
		if err := confirmDelete(cmd); err != nil {
			return err
		}
	}

	// Perform deletes per context
	for name, p := range plans {
		fmt.Printf("\n%s\n", name)
		executeDeletePlan(cmd.Context(), p)
	}

	return nil
}

func DeleteExample() string {
	return `
  # Delete everything swarmctl has installed in the current context
  swarmctl delete

  # Same using the command alias
  swarmctl rm

  # Delete everything in all contexts that match a regex
  swarmctl delete --context 'kind-pasta-.*'

  # Skip the confirmation prompt
  swarmctl delete --context 'kind-pasta-.*' --yes

  # Show what would be deleted without contacting the cluster
  swarmctl delete --context 'kind-pasta-.*' --dry-run
  `
}
