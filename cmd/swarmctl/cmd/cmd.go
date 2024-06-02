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
	"github.com/octoroot/k-swarm/cmd/swarmctl/pkg/k8sctx"
	"github.com/octoroot/k-swarm/cmd/swarmctl/pkg/profiling"
	"github.com/octoroot/k-swarm/cmd/swarmctl/pkg/swarmctl"
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

	// manifestDumpCmd flags
	manifestDumpCmd.Flags().Bool("stdout", false, "Output to stdout")

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
// validateFlags
//-----------------------------------------------------------------------------

func validateFlags(cmd *cobra.Command, args []string) error {

	if cmd.Flags().Changed("context") {
		if valid := contextIsValid(); !valid {
			return errors.New("invalid context")
		}
	}

	if cmd.Flags().Changed("replicas") {
		if valid := replicasIsValid(); !valid {
			return errors.New("invalid replicas")
		}
	}

	if cmd.Flags().Changed("node-selector") {
		if valid := nodeSelectorIsValid(); !valid {
			return errors.New("invalid node-selector")
		}
	}

	if cmd.Flags().Changed("image-tag") {
		if valid := imageTagIsValid(); !valid {
			return errors.New("invalid image-tag")
		}
	}

	if cmd.Flags().Changed("istio-revision") {
		if valid := istioRevisionIsValid(); !valid {
			return errors.New("invalid istio-revision")
		}
	}

	// Return
	return nil
}
