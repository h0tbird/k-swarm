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
	manifestInstallCmd.AddGroup(&cobra.Group{ID: "install", Title: "Install subcommands:"})
	manifestGenerateCmd.AddGroup(&cobra.Group{ID: "generate", Title: "Generate subcommands:"})

	// Add commands
	rootCmd.AddCommand(manifestCmd, manifestInstallInformerCmd, manifestInstallWorkerCmd)
	manifestCmd.AddCommand(manifestDumpCmd, manifestGenerateCmd, manifestInstallCmd)
	manifestGenerateCmd.AddCommand(manifestGenerateInformerCmd, manifestGenerateWorkerCmd)
	manifestGenerateInformerCmd.AddCommand(manifestGenerateInformerTelemetryCmd)
	manifestGenerateWorkerCmd.AddCommand(manifestGenerateWorkerTelemetryCmd)
	manifestInstallCmd.AddCommand(manifestInstallInformerCmd, manifestInstallWorkerCmd)
	manifestInstallInformerCmd.AddCommand(manifestInstallInformerTelemetryCmd)
	manifestInstallWorkerCmd.AddCommand(manifestInstallWorkerTelemetryCmd)

	// Profiling flags
	rootCmd.PersistentFlags().BoolVar(&profiling.CPUProfile, "cpu-profile", false, "write cpu profile to file")
	rootCmd.PersistentFlags().BoolVar(&profiling.MemProfile, "mem-profile", false, "write memory profile to file")
	rootCmd.PersistentFlags().BoolVar(&profiling.Tracing, "tracing", false, "write trace to file")
	rootCmd.PersistentFlags().StringVar(&profiling.CPUProfileFile, "cpu-profile-file", "cpu.prof", "file for CPU profiling output")
	rootCmd.PersistentFlags().StringVar(&profiling.MemProfileFile, "mem-profile-file", "mem.prof", "file for memory profiling output")
	rootCmd.PersistentFlags().StringVar(&profiling.TracingFile, "tracing-file", "trace.out", "file for tracing output")

	//---------------------
	// manifest dump flags
	//---------------------

	// --stdout flag
	manifestDumpCmd.Flags().Bool("stdout", false, "Output to stdout")

	//-------------------------
	// manifest generate flags
	//-------------------------

	// --context flag
	manifestGenerateCmd.PersistentFlags().String("context", "", "regex to match the context name.")
	if err := manifestGenerateCmd.RegisterFlagCompletionFunc("context", contextCompletion); err != nil {
		panic(err)
	}

	// --replicas flag
	manifestGenerateCmd.PersistentFlags().Int("replicas", 1, "Number of replicas to deploy.")
	if err := manifestGenerateCmd.RegisterFlagCompletionFunc("replicas", replicasCompletion); err != nil {
		panic(err)
	}

	// --node-selector flag
	manifestGenerateCmd.PersistentFlags().String("node-selector", "", "Node selector to use for deployment.")
	if err := manifestGenerateCmd.RegisterFlagCompletionFunc("node-selector", nodeSelectorCompletion); err != nil {
		panic(err)
	}

	// --image-tag flag
	manifestGenerateCmd.PersistentFlags().String("image-tag", "", "Image tag to use for deployment.")
	if err := manifestGenerateCmd.RegisterFlagCompletionFunc("image-tag", imageTagCompletion); err != nil {
		panic(err)
	}

	// --istio-revision flag
	manifestGenerateCmd.PersistentFlags().String("istio-revision", "", "Istio revision label to use for the namespace.")
	if err := manifestGenerateCmd.RegisterFlagCompletionFunc("istio-revision", istioRevisionCompletion); err != nil {
		panic(err)
	}

	// --cluster-domain flag
	manifestGenerateCmd.PersistentFlags().String("cluster-domain", "", "Cluster domain suffix (default: cluster.local for generate, auto-detect for install).")
	if err := manifestGenerateCmd.RegisterFlagCompletionFunc("cluster-domain", clusterDomainCompletion); err != nil {
		panic(err)
	}

	// --dataplane-mode flag
	manifestGenerateCmd.PersistentFlags().String("dataplane-mode", "", "Istio dataplane mode: sidecar or ambient (required).")
	if err := manifestGenerateCmd.RegisterFlagCompletionFunc("dataplane-mode", dataplaneModeCompletion); err != nil {
		panic(err)
	}
	if err := manifestGenerateCmd.MarkPersistentFlagRequired("dataplane-mode"); err != nil {
		panic(err)
	}

	// --waypoint-name flag
	manifestGenerateCmd.PersistentFlags().String("waypoint-name", "waypoint", "Name of the per-namespace ambient waypoint Gateway.")
	if err := manifestGenerateCmd.RegisterFlagCompletionFunc("waypoint-name", waypointNameCompletion); err != nil {
		panic(err)
	}

	// --ingress-mode flag
	manifestGenerateCmd.PersistentFlags().String("ingress-mode", "none", "Ingress mode: 'none', 'shared' (classic Istio Gateway/VirtualService selecting istio: nsgw) or 'dedicated' (per-service Gateway API Gateway/HTTPRoute).")
	if err := manifestGenerateCmd.RegisterFlagCompletionFunc("ingress-mode", ingressModeCompletion); err != nil {
		panic(err)
	}

	// --multi-cluster flag
	manifestGenerateCmd.PersistentFlags().Bool("multi-cluster", false, "Enable cross-cluster failover for ambient mode: labels the worker and waypoint Services with istio.io/global=true and emits a DestinationRule with locality failover by topology.istio.io/cluster.")

	// --log-responses flag
	manifestGenerateCmd.PersistentFlags().Bool("log-responses", false, "If set, the worker logs the raw JSON response bodies received from the informer's /services endpoint and from peer workers' /data endpoint.")

	//------------------------
	// manifest install flags
	//------------------------

	// --yes flag
	manifestInstallCmd.PersistentFlags().Bool("yes", false, "Automatically confirm all prompts with 'yes'.")

	// --context flag
	manifestInstallCmd.PersistentFlags().String("context", "", "regex to match the context name.")
	if err := manifestInstallCmd.RegisterFlagCompletionFunc("context", contextCompletion); err != nil {
		panic(err)
	}

	// --replicas flag
	manifestInstallCmd.PersistentFlags().Int("replicas", 1, "Number of replicas to deploy.")
	if err := manifestInstallCmd.RegisterFlagCompletionFunc("replicas", replicasCompletion); err != nil {
		panic(err)
	}

	// --node-selector flag
	manifestInstallCmd.PersistentFlags().String("node-selector", "", "Node selector to use for deployment.")
	if err := manifestInstallCmd.RegisterFlagCompletionFunc("node-selector", nodeSelectorCompletion); err != nil {
		panic(err)
	}

	// --image-tag flag
	manifestInstallCmd.PersistentFlags().String("image-tag", "", "Image tag to use for deployment.")
	if err := manifestInstallCmd.RegisterFlagCompletionFunc("image-tag", imageTagCompletion); err != nil {
		panic(err)
	}

	// --istio-revision flag
	manifestInstallCmd.PersistentFlags().String("istio-revision", "", "Istio revision label to use for the namespace.")
	if err := manifestInstallCmd.RegisterFlagCompletionFunc("istio-revision", istioRevisionCompletion); err != nil {
		panic(err)
	}

	// --cluster-domain flag
	manifestInstallCmd.PersistentFlags().String("cluster-domain", "", "Cluster domain suffix (default: cluster.local for generate, auto-detect for install).")
	if err := manifestInstallCmd.RegisterFlagCompletionFunc("cluster-domain", clusterDomainCompletion); err != nil {
		panic(err)
	}

	// --dataplane-mode flag
	manifestInstallCmd.PersistentFlags().String("dataplane-mode", "", "Istio dataplane mode: sidecar or ambient (required).")
	if err := manifestInstallCmd.RegisterFlagCompletionFunc("dataplane-mode", dataplaneModeCompletion); err != nil {
		panic(err)
	}
	if err := manifestInstallCmd.MarkPersistentFlagRequired("dataplane-mode"); err != nil {
		panic(err)
	}

	// --waypoint-name flag
	manifestInstallCmd.PersistentFlags().String("waypoint-name", "waypoint", "Name of the per-namespace ambient waypoint Gateway.")
	if err := manifestInstallCmd.RegisterFlagCompletionFunc("waypoint-name", waypointNameCompletion); err != nil {
		panic(err)
	}

	// --ingress-mode flag
	manifestInstallCmd.PersistentFlags().String("ingress-mode", "none", "Ingress mode: 'none', 'shared' (classic Istio Gateway/VirtualService selecting istio: nsgw) or 'dedicated' (per-service Gateway API Gateway/HTTPRoute).")
	if err := manifestInstallCmd.RegisterFlagCompletionFunc("ingress-mode", ingressModeCompletion); err != nil {
		panic(err)
	}

	// --multi-cluster flag
	manifestInstallCmd.PersistentFlags().Bool("multi-cluster", false, "Enable cross-cluster failover for ambient mode: labels the worker and waypoint Services with istio.io/global=true and emits a DestinationRule with locality failover by topology.istio.io/cluster.")

	// --log-responses flag
	manifestInstallCmd.PersistentFlags().Bool("log-responses", false, "If set, the worker logs the raw JSON response bodies received from the informer's /services endpoint and from peer workers' /data endpoint.")
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

var manifestCmd = &cobra.Command{
	Use:     "manifest",
	Short:   "Manages manifests.",
	Aliases: []string{"m"},
}

var manifestDumpCmd = &cobra.Command{
	Use:          "dump [informer] [worker]",
	Short:        "Dumps templates to ~/.swarmctl or stdout.",
	SilenceUsage: true,
	Example:      swarmctl.DumpExample(),
	Aliases:      []string{"d"},
	ValidArgs:    []string{"informer", "worker"},
	Args:         cobra.MatchAll(cobra.MaximumNArgs(2), cobra.OnlyValidArgs),
	RunE:         swarmctl.Dump,
}

var manifestGenerateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generates a manifest and outputs it.",
	Aliases: []string{"g"},
}

var manifestGenerateInformerCmd = &cobra.Command{
	Use:          "informer",
	Short:        "Generates the informer's manifests.",
	GroupID:      "generate",
	SilenceUsage: true,
	Example:      swarmctl.GenerateInformerExample(),
	Aliases:      []string{"i"},
	Args:         cobra.ExactArgs(0),
	PreRunE:      validateFlags,
	RunE:         swarmctl.GenerateInformer,
}

var manifestGenerateInformerTelemetryCmd = &cobra.Command{
	Use:          "telemetry (on|off)",
	Short:        "Generates the informer's telemetry manifests.",
	SilenceUsage: true,
	Example:      swarmctl.GenerateInformerTelemetryExample(),
	Aliases:      []string{"t"},
	Args:         cobra.ExactArgs(1),
	ValidArgs:    []string{"on", "off"},
	PreRunE:      validateFlags,
	RunE:         swarmctl.GenerateInformerTelemetry,
}

var manifestGenerateWorkerCmd = &cobra.Command{
	Use:          "worker <start:end>",
	Short:        "Generates the worker's manifests.",
	GroupID:      "generate",
	SilenceUsage: true,
	Example:      swarmctl.GenerateWorkerExample(),
	Aliases:      []string{"w"},
	Args:         cobra.RangeArgs(1, 3),
	PreRunE:      validateFlags,
	RunE:         swarmctl.GenerateWorker,
}

var manifestGenerateWorkerTelemetryCmd = &cobra.Command{
	Use:          "telemetry (on|off)",
	Short:        "Generates the worker's telemetry manifests.",
	SilenceUsage: true,
	Example:      swarmctl.GenerateWorkerTelemetryExample(),
	Aliases:      []string{"t"},
	Args:         cobra.ExactArgs(1),
	ValidArgs:    []string{"on", "off"},
	PreRunE:      validateFlags,
	RunE:         swarmctl.GenerateWorkerTelemetry,
}

var manifestInstallCmd = &cobra.Command{
	Use:               "install",
	Short:             "Generates manifests and applies them.",
	Aliases:           []string{"i"},
	PersistentPreRunE: swarmctl.Install,
}

var manifestInstallInformerCmd = &cobra.Command{
	Use:          "informer",
	Short:        "Installs the informer's manifests.",
	GroupID:      "install",
	SilenceUsage: true,
	Example:      swarmctl.InstallInformerExample(),
	Aliases:      []string{"i"},
	Args:         cobra.ExactArgs(0),
	PreRunE:      validateFlags,
	RunE:         swarmctl.InstallInformer,
}

var manifestInstallInformerTelemetryCmd = &cobra.Command{
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

var manifestInstallWorkerCmd = &cobra.Command{
	Use:          "worker <start:end>",
	Short:        "Installs the worker's manifests.",
	GroupID:      "install",
	SilenceUsage: true,
	Example:      swarmctl.InstallWorkerExample(),
	Aliases:      []string{"w"},
	Args:         cobra.ExactArgs(1),
	ValidArgs:    []string{"1:1"},
	PreRunE:      validateFlags,
	RunE:         swarmctl.InstallWorker,
}

var manifestInstallWorkerTelemetryCmd = &cobra.Command{
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

	// --multi-cluster requires --dataplane-mode=ambient
	if cmd.Flags().Changed("multi-cluster") {
		multiCluster, _ := cmd.Flags().GetBool("multi-cluster")
		dataplaneMode, _ := cmd.Flags().GetString("dataplane-mode")
		if multiCluster && dataplaneMode != "ambient" {
			return errors.New("--multi-cluster requires --dataplane-mode=ambient")
		}
	}

	// Return
	return nil
}
