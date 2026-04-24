package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"errors"
	"strings"

	// Community
	"github.com/spf13/cobra"

	// Local
	"github.com/h0tbird/k-swarm/cmd/swarmctl/pkg/k8sctx"
	"github.com/h0tbird/k-swarm/cmd/swarmctl/pkg/profiling"
	"github.com/h0tbird/k-swarm/cmd/swarmctl/pkg/swarmctl"
)

//-----------------------------------------------------------------------------
// Globals
//-----------------------------------------------------------------------------

var version = "0.0.0"

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Command grouping
	rootCmd.AddGroup(&cobra.Group{ID: "install", Title: "Install subcommands:"})

	// Add commands
	rootCmd.AddCommand(dumpCmd, informerCmd, workerCmd)
	informerCmd.AddCommand(informerTelemetryCmd)
	workerCmd.AddCommand(workerTelemetryCmd)

	// Profiling flags
	rootCmd.PersistentFlags().BoolVar(&profiling.CPUProfile, "cpu-profile", false, "write cpu profile to file")
	rootCmd.PersistentFlags().BoolVar(&profiling.MemProfile, "mem-profile", false, "write memory profile to file")
	rootCmd.PersistentFlags().BoolVar(&profiling.Tracing, "tracing", false, "write trace to file")
	rootCmd.PersistentFlags().StringVar(&profiling.CPUProfileFile, "cpu-profile-file", "cpu.prof", "file for CPU profiling output")
	rootCmd.PersistentFlags().StringVar(&profiling.MemProfileFile, "mem-profile-file", "mem.prof", "file for memory profiling output")
	rootCmd.PersistentFlags().StringVar(&profiling.TracingFile, "tracing-file", "trace.out", "file for tracing output")

	//------------
	// dump flags
	//------------

	// --stdout flag
	dumpCmd.Flags().Bool("stdout", false, "Output to stdout")

	//---------------------------
	// informer and worker flags
	//---------------------------

	// Registered separately on informerCmd and workerCmd so they don't leak
	// into `swarmctl --help` or `swarmctl dump --help`. Telemetry subcommands
	// inherit them via their parent.
	for _, c := range []*cobra.Command{informerCmd, workerCmd} {

		// --context flag
		c.PersistentFlags().String("context", "", "regex to match the context name.")
		if err := c.RegisterFlagCompletionFunc("context", contextCompletion); err != nil {
			panic(err)
		}

		// --replicas flag
		c.PersistentFlags().Int("replicas", 1, "Number of replicas to deploy.")
		if err := c.RegisterFlagCompletionFunc("replicas", replicasCompletion); err != nil {
			panic(err)
		}

		// --node-selector flag
		c.PersistentFlags().String("node-selector", "", "Node selector to use for deployment.")
		if err := c.RegisterFlagCompletionFunc("node-selector", nodeSelectorCompletion); err != nil {
			panic(err)
		}

		// --image-tag flag
		c.PersistentFlags().String("image-tag", "", "Image tag to use for deployment.")
		if err := c.RegisterFlagCompletionFunc("image-tag", imageTagCompletion); err != nil {
			panic(err)
		}

		// --istio-revision flag
		c.PersistentFlags().String("istio-revision", "", "Istio revision label to use for the namespace.")
		if err := c.RegisterFlagCompletionFunc("istio-revision", istioRevisionCompletion); err != nil {
			panic(err)
		}

		// --cluster-domain flag
		c.PersistentFlags().String("cluster-domain", "", "Cluster domain suffix (default: auto-detect from CoreDNS, or 'cluster.local' in --dry-run).")
		if err := c.RegisterFlagCompletionFunc("cluster-domain", clusterDomainCompletion); err != nil {
			panic(err)
		}

		// --dataplane-mode flag
		c.PersistentFlags().String("dataplane-mode", "", "Istio dataplane mode: sidecar or ambient (required).")
		if err := c.RegisterFlagCompletionFunc("dataplane-mode", dataplaneModeCompletion); err != nil {
			panic(err)
		}
		if err := c.MarkPersistentFlagRequired("dataplane-mode"); err != nil {
			panic(err)
		}

		// --waypoint-name flag
		c.PersistentFlags().String("waypoint-name", "waypoint", "Name of the per-namespace ambient waypoint Gateway.")
		if err := c.RegisterFlagCompletionFunc("waypoint-name", waypointNameCompletion); err != nil {
			panic(err)
		}

		// --ingress-mode flag
		c.PersistentFlags().String("ingress-mode", "none", "Ingress mode: 'none', 'shared' (classic Istio Gateway/VirtualService selecting istio: nsgw) or 'dedicated' (per-service Gateway API Gateway/HTTPRoute).")
		if err := c.RegisterFlagCompletionFunc("ingress-mode", ingressModeCompletion); err != nil {
			panic(err)
		}

		// --yes flag
		c.PersistentFlags().Bool("yes", false, "Automatically confirm all prompts with 'yes'.")

		// --dry-run flag
		c.PersistentFlags().Bool("dry-run", false, "Render manifests to stdout without applying them or contacting the cluster.")

		// --multi-cluster flag
		c.PersistentFlags().Bool("multi-cluster", false, "Enable cross-cluster failover for ambient mode: labels the worker and waypoint Services with istio.io/global=true and emits a DestinationRule with locality failover by topology.istio.io/cluster.")

		// --log-responses flag
		c.PersistentFlags().Bool("log-responses", false, "If set, the worker logs the raw JSON response bodies received from the informer's /services endpoint and from peer workers' /data endpoint.")
	}
}

//-----------------------------------------------------------------------------
// This is called by main()
//-----------------------------------------------------------------------------

func Execute() error {
	defer profiling.Stop()
	return rootCmd.Execute()
}

//-----------------------------------------------------------------------------
// All commands
//-----------------------------------------------------------------------------

var rootCmd = &cobra.Command{
	Version:           version,
	Use:               "swarmctl",
	Short:             "swarmctl controls the swarm.",
	PersistentPreRunE: swarmctl.Root,
}

var dumpCmd = &cobra.Command{
	Use:          "dump [informer] [worker]",
	Short:        "Dumps templates to ~/.swarmctl or stdout.",
	SilenceUsage: true,
	Example:      swarmctl.DumpExample(),
	Aliases:      []string{"d"},
	ValidArgs:    []string{"informer", "worker"},
	Args:         cobra.MatchAll(cobra.MaximumNArgs(2), cobra.OnlyValidArgs),
	RunE:         swarmctl.Dump,
}

var informerCmd = &cobra.Command{
	Use:               "informer",
	Short:             "Installs the informer's manifests.",
	GroupID:           "install",
	SilenceUsage:      true,
	Example:           swarmctl.InstallInformerExample(),
	Aliases:           []string{"i"},
	Args:              cobra.ExactArgs(0),
	PersistentPreRunE: swarmctl.Install,
	PreRunE:           validateFlags,
	RunE:              swarmctl.InstallInformer,
}

var informerTelemetryCmd = &cobra.Command{
	Use:          "telemetry (on|off)",
	Short:        "Installs the informer's telemetry manifests.",
	SilenceUsage: true,
	Example:      swarmctl.InstallInformerTelemetryExample(),
	Aliases:      []string{"t"},
	Args:         cobra.ExactArgs(1),
	ValidArgs:    []string{"on", "off"},
	PreRunE:      validateFlags,
	RunE:         swarmctl.InstallInformerTelemetry,
}

var workerCmd = &cobra.Command{
	Use:               "worker <start:end>",
	Short:             "Installs the worker's manifests.",
	GroupID:           "install",
	SilenceUsage:      true,
	Example:           swarmctl.InstallWorkerExample(),
	Aliases:           []string{"w"},
	Args:              cobra.ExactArgs(1),
	ValidArgs:         []string{"1:1"},
	PersistentPreRunE: swarmctl.Install,
	PreRunE:           validateFlags,
	RunE:              swarmctl.InstallWorker,
}

var workerTelemetryCmd = &cobra.Command{
	Use:               "telemetry <start:end> (on|off)",
	Short:             "Installs the worker's telemetry manifests.",
	SilenceUsage:      true,
	Example:           swarmctl.InstallWorkerTelemetryExample(),
	Aliases:           []string{"t"},
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: swarmctl.InstallWorkerTelemetryValidArgs,
	PreRunE:           validateFlags,
	RunE:              swarmctl.InstallWorkerTelemetry,
}

//-----------------------------------------------------------------------------
// context
//-----------------------------------------------------------------------------

// contextCompletion
func contextCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

	// Get the contexts
	contexts, err := k8sctx.List()
	if err != nil {
		panic(err)
	}

	// Filter the contexts
	var completions []string
	for _, context := range contexts {
		if strings.HasPrefix(context, toComplete) {
			completions = append(completions, context)
		}
	}

	// Return the completions
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// contextIsValid
func contextIsValid() bool {
	return true
}

//-----------------------------------------------------------------------------
// replicas
//-----------------------------------------------------------------------------

// replicasCompletion
func replicasCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"1"}, cobra.ShellCompDirectiveNoFileComp
}

// replicasIsValid
func replicasIsValid() bool {
	return true
}

//-----------------------------------------------------------------------------
// nodeSelector
//-----------------------------------------------------------------------------

// nodeSelectorCompletion
func nodeSelectorCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"{key1: value1, key2: value2}"}, cobra.ShellCompDirectiveNoFileComp
}

// nodeSelectorIsValid
func nodeSelectorIsValid() bool {
	return true
}

//-----------------------------------------------------------------------------
// imageTag
//-----------------------------------------------------------------------------

// imageTagCompletion
func imageTagCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"main", "pr-xx"}, cobra.ShellCompDirectiveNoFileComp
}

// imageTagIsValid
func imageTagIsValid() bool {
	return true
}

//-----------------------------------------------------------------------------
// istioRevision
//-----------------------------------------------------------------------------

// istioRevisionCompletion
func istioRevisionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"1-19-x", "1-20-x", "1-21-x"}, cobra.ShellCompDirectiveNoFileComp
}

// istioRevisionIsValid
func istioRevisionIsValid() bool {
	return true
}

//-----------------------------------------------------------------------------
// clusterDomain
//-----------------------------------------------------------------------------

// clusterDomainCompletion
func clusterDomainCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"cluster.local"}, cobra.ShellCompDirectiveNoFileComp
}

// clusterDomainIsValid
func clusterDomainIsValid() bool {
	return true
}

//-----------------------------------------------------------------------------
// dataplaneMode
//-----------------------------------------------------------------------------

// dataplaneModeCompletion
func dataplaneModeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"ambient", "sidecar"}, cobra.ShellCompDirectiveNoFileComp
}

// dataplaneModeIsValid
func dataplaneModeIsValid(value string) bool {
	return value == "sidecar" || value == "ambient"
}

//-----------------------------------------------------------------------------
// waypointName
//-----------------------------------------------------------------------------

// waypointNameCompletion
func waypointNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"waypoint"}, cobra.ShellCompDirectiveNoFileComp
}

// waypointNameIsValid
func waypointNameIsValid(value string) bool {
	return value != ""
}

//-----------------------------------------------------------------------------
// ingressMode
//-----------------------------------------------------------------------------

// ingressModeCompletion
func ingressModeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"none", "shared", "dedicated"}, cobra.ShellCompDirectiveNoFileComp
}

// ingressModeIsValid
func ingressModeIsValid(value string) bool {
	switch value {
	case "none", "shared", "dedicated":
		return true
	}
	return false
}

//-----------------------------------------------------------------------------
// validateFlags
//-----------------------------------------------------------------------------

func validateFlags(cmd *cobra.Command, args []string) error {

	if cmd.Flags().Changed("context") {
		if !contextIsValid() {
			return errors.New("invalid context")
		}
	}

	if cmd.Flags().Changed("replicas") {
		if !replicasIsValid() {
			return errors.New("invalid replicas")
		}
	}

	if cmd.Flags().Changed("node-selector") {
		if !nodeSelectorIsValid() {
			return errors.New("invalid node-selector")
		}
	}

	if cmd.Flags().Changed("image-tag") {
		if !imageTagIsValid() {
			return errors.New("invalid image-tag")
		}
	}

	if cmd.Flags().Changed("istio-revision") {
		if !istioRevisionIsValid() {
			return errors.New("invalid istio-revision")
		}
	}

	if cmd.Flags().Changed("cluster-domain") {
		if !clusterDomainIsValid() {
			return errors.New("invalid cluster-domain")
		}
	}

	if cmd.Flags().Changed("dataplane-mode") {
		value, _ := cmd.Flags().GetString("dataplane-mode")
		if !dataplaneModeIsValid(value) {
			return errors.New("invalid dataplane-mode (must be 'sidecar' or 'ambient')")
		}
	}

	if cmd.Flags().Changed("waypoint-name") {
		value, _ := cmd.Flags().GetString("waypoint-name")
		if !waypointNameIsValid(value) {
			return errors.New("invalid waypoint-name")
		}
	}

	if cmd.Flags().Changed("ingress-mode") {
		value, _ := cmd.Flags().GetString("ingress-mode")
		if !ingressModeIsValid(value) {
			return errors.New("invalid ingress-mode (must be 'none', 'shared' or 'dedicated')")
		}
	}

	// Return
	return nil
}
