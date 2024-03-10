package swarmctl

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	// Community
	"github.com/spf13/cobra"

	// Local
	"github.com/octoroot/k-swarm/cmd/swarmctl/pkg/k8sctx"
	"github.com/octoroot/k-swarm/cmd/swarmctl/pkg/profiling"
	"github.com/octoroot/k-swarm/cmd/swarmctl/pkg/util"
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
// Dump writes the informer and/or worker template to ~/.swarmctl
//-----------------------------------------------------------------------------

func Dump(cmd *cobra.Command, args []string) error {

	// Get the flags
	stdout, _ := cmd.Flags().GetBool("stdout")

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// No args? Default to both
	if len(args) == 0 {
		args = []string{"informer", "worker"}
	}

	// Create ~/.swarmctl
	if !stdout {
		if err := os.MkdirAll(util.SwarmDir, 0755); err != nil {
			return fmt.Errorf("error creating ~/.swarmctl: %w", err)
		}
	}

	// Loop through the components
	for _, component := range args {

		// Open the file from the embedded file system
		fileData, err := Assets.ReadFile(fmt.Sprintf("assets/%s.goyaml", component))
		if err != nil {
			return fmt.Errorf("error reading file from embedded FS: %w", err)
		}

		// Write the content to stdout
		if stdout {
			_, err = io.Copy(os.Stdout, bytes.NewReader(fileData))
			if err != nil {
				return fmt.Errorf("error writing file data to stdout: %w", err)
			}
			continue
		}

		// Write the contents to ~/.swarmctl/<component>.goyaml
		if err := os.WriteFile(util.SwarmDir+"/"+component+".goyaml", fileData, 0644); err != nil {
			return fmt.Errorf("error writing file data to ~/.swarmctl/%s.goyaml: %w", component, err)
		}

		// Print the success message
		cmd.Printf("Successfully wrote ~/.swarmctl/%s.goyaml\n", component)
	}

	return nil
}

func DumpExample() string {
	return `
  # Dump the informer and worker templates to ~/.swarmctl
  swarmctl manifest dump

  # Dump only the informer template to ~/.swarmctl
  swarmctl m d informer

  # Dump the informer and worker templates to stdout
  swarmctl m d --stdout
  `
}

//-----------------------------------------------------------------------------
// GenerateInformer outputs the informer manifest
//-----------------------------------------------------------------------------

func GenerateInformer(cmd *cobra.Command, args []string) error {

	// Get the flags
	replicas, _ := cmd.Flags().GetInt("replicas")
	nodeSelector, _ := cmd.Flags().GetString("node-selector")
	imageTag, _ := cmd.Flags().GetString("image-tag")

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// Parse the template
	tmpl, err := util.ParseTemplate(Assets, "informer")
	if err != nil {
		return err
	}

	// Render the template
	tmpl.Execute(cmd.OutOrStdout(), struct {
		Replicas     int
		NodeSelector string
		Version      string
		ImageTag     string
	}{
		Replicas:     replicas,
		NodeSelector: nodeSelector,
		Version:      cmd.Root().Version,
		ImageTag:     imageTag,
	})

	// Return
	return nil
}

func GenerateInformerExample() string {
	return `
  # Output the generated informer manifest to stdout
  swarmctl manifest generate informer

  # Same using command aliases
  swarmctl m g i

  # Set informer replicas and node selector
  swarmctl m g i --replicas 3 --node-selector '{key1: value1, key2: value2}'
  `
}

//-----------------------------------------------------------------------------
// GenerateInformerTelemetry outputs the informer telemetry manifest
//-----------------------------------------------------------------------------

func GenerateInformerTelemetry(cmd *cobra.Command, args []string) error {

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// Parse the template
	tmpl, err := util.ParseTemplate(Assets, "telemetry")
	if err != nil {
		return err
	}

	// Render the template
	tmpl.Execute(cmd.OutOrStdout(), struct {
		OnOff     string
		Namespace string
	}{
		OnOff:     args[0],
		Namespace: "informer",
	})

	// Return
	return nil
}

func GenerateInformerTelemetryExample() string {
	return `
  # Output the generated informer telemetry manifest to stdout
  swarmctl manifest generate informer telemetry on

  # Same using command aliases
  swarmctl m g i t on
  `
}

//-----------------------------------------------------------------------------
// GenerateWorker outputs the worker manifest
//-----------------------------------------------------------------------------

func GenerateWorker(cmd *cobra.Command, args []string) error {

	// Get the flags
	replicas, _ := cmd.Flags().GetInt("replicas")
	nodeSelector, _ := cmd.Flags().GetString("node-selector")
	imageTag, _ := cmd.Flags().GetString("image-tag")

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// Parse the range
	start, end, err := util.ParseRange(args[0])
	if err != nil {
		return err
	}

	// Parse the template
	tmpl, err := util.ParseTemplate(Assets, "worker")
	if err != nil {
		return err
	}

	// Loop from start to end
	for i := start; i <= end; i++ {

		// Render the template
		tmpl.Execute(cmd.OutOrStdout(), struct {
			Replicas     int
			Namespace    string
			NodeSelector string
			Version      string
			ImageTag     string
		}{
			Replicas:     replicas,
			Namespace:    fmt.Sprintf("service-%d", i),
			NodeSelector: nodeSelector,
			Version:      cmd.Root().Version,
			ImageTag:     imageTag,
		})
	}

	// Return
	return nil
}

func GenerateWorkerExample() string {
	return `
  # Output the generated workers 1 to 1 manifests to stdout
  swarmctl manifest generate worker 1:1

  # Same using command aliases
  swarmctl m g w 1:1

  # Set worker replicas and node selector
  swarmctl m g w 1:1 --replicas 3 --node-selector '{key1: value1, key2: value2}'
  `
}

//-----------------------------------------------------------------------------
// GenerateWorkerTelemetry outputs the worker telemetry manifest
//-----------------------------------------------------------------------------

func GenerateWorkerTelemetry(cmd *cobra.Command, args []string) error {

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// Split args[0] into start and end
	parts := strings.Split(args[0], ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid range format. Please use the format start:end")
	}

	// Convert start and end to integers
	start, err1 := strconv.Atoi(parts[0])
	end, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return fmt.Errorf("invalid range. Both start and end should be integers")
	}

	// Parse the template
	tmpl, err := util.ParseTemplate(Assets, "telemetry")
	if err != nil {
		return err
	}

	// Loop from start to end
	for i := start; i <= end; i++ {

		// Render the template
		tmpl.Execute(cmd.OutOrStdout(), struct {
			OnOff     string
			Namespace string
		}{
			OnOff:     args[0],
			Namespace: fmt.Sprintf("service-%d", i),
		})
	}

	// Return
	return nil
}

func GenerateWorkerTelemetryExample() string {
	return `
  # Output the generated worker telemetry manifest to stdout
  swarmctl manifest generate worker 1:1 telemetry on

  # Same using command aliases
  swarmctl m g w 1:1 t on
  `
}

//-----------------------------------------------------------------------------
// Install
//-----------------------------------------------------------------------------

func Install(cmd *cobra.Command, args []string) error {

	// Get the flags
	ctxRegex, _ := cmd.Flags().GetString("context")

	// Run the root PersistentPreRunE
	if err := cmd.Root().PersistentPreRunE(cmd, args); err != nil {
		return err
	}

	// Get the contexts that match the regex
	matches, err := k8sctx.Filter(ctxRegex)
	if err != nil {
		return err
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
	cmd.Print("\nProceed? (y/N) ")
	var answer string
	if _, err := fmt.Scanln(&answer); err != nil {
		return err
	}
	if answer != "y" {
		cmd.SetErrPrefix("aborted:")
		return errors.New("by user")
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

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// Read the CRDs
	crds, err := Assets.ReadFile("assets/crds.yaml")
	if err != nil {
		return err
	}

	// Parse the template
	tmpl, err := util.ParseTemplate(Assets, "informer")
	if err != nil {
		return err
	}

	// Loop through all contexts
	for name, context := range Contexts {

		// Print the context
		fmt.Printf("\n%s\n\n", name)

		// Loop through all CRDs
		for _, doc := range util.SplitYAML(bytes.NewBuffer(crds)) {
			if err := context.ApplyYaml(doc); err != nil {
				fmt.Printf("\nError: %s\n", err)
			}
		}

		// Render the template
		docs, err := util.RenderTemplate(tmpl, struct {
			Replicas     int
			NodeSelector string
			Version      string
			ImageTag     string
		}{
			Replicas:     replicas,
			NodeSelector: nodeSelector,
			Version:      cmd.Root().Version,
			ImageTag:     imageTag,
		})
		if err != nil {
			return err
		}

		// Loop through all yaml documents
		for _, doc := range docs {
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
  swarmctl manifest install informer

  # Same using command aliases
  swarmctl m i i

  # Same using a shoret command chain
  swarmctl informer

  # Same using a short command chain with aliases
  swarmctl i

  # Install the informer to a specific context
  swarmctl i --context my-context

  # Install the informer to all contexts that match a regex
  swarmctl i --context 'my-.*'

  # Install the informer to all contexts that match a regex and set the replicas
  swarmctl i --context 'my-.*' --replicas 3

  # Install the informer to all contexts that match a regex and set the node selector
  swarmctl i --context 'my-.*' --node-selector '{key1: value1, key2: value2}'
  `
}

//-----------------------------------------------------------------------------
// InstallInformerTelemetry
//-----------------------------------------------------------------------------

func InstallInformerTelemetry(cmd *cobra.Command, args []string) error {

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// Read the CRDs
	crds, err := Assets.ReadFile("assets/crds.yaml")
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

		// Print the context
		fmt.Printf("\n%s\n\n", name)

		// Loop through all CRDs
		for _, doc := range util.SplitYAML(bytes.NewBuffer(crds)) {
			if err := context.ApplyYaml(doc); err != nil {
				fmt.Printf("\nError: %s\n", err)
			}
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

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// Parse the range
	start, end, err := util.ParseRange(args[0])
	if err != nil {
		return err
	}

	// Read the CRDs
	crds, err := Assets.ReadFile("assets/crds.yaml")
	if err != nil {
		return err
	}

	// Parse the template
	tmpl, err := util.ParseTemplate(Assets, "worker")
	if err != nil {
		return err
	}

	// Loop through all contexts
	for name, context := range Contexts {

		// Print the context
		fmt.Printf("\n%s\n\n", name)

		// Loop through all CRDs
		for _, doc := range util.SplitYAML(bytes.NewBuffer(crds)) {
			if err := context.ApplyYaml(doc); err != nil {
				fmt.Printf("\nError: %s\n", err)
			}
		}

		// Loop trough all services
		for i := start; i <= end; i++ {

			fmt.Printf("\n")

			// Render the template
			docs, err := util.RenderTemplate(tmpl, struct {
				Replicas     int
				Namespace    string
				NodeSelector string
				Version      string
				ImageTag     string
			}{
				Replicas:     replicas,
				Namespace:    fmt.Sprintf("service-%d", i),
				NodeSelector: nodeSelector,
				Version:      cmd.Root().Version,
				ImageTag:     imageTag,
			})
			if err != nil {
				return err
			}

			// Loop through all yaml documents
			for _, doc := range docs {
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
  swarmctl manifest install worker 1:1

  # Same using command aliases
  swarmctl m i w 1:1

  # Same using a shoret command chain
  swarmctl worker 1:1

  # Same using a short command chain with aliases
  swarmctl w 1:1

  # Install the workers 1 to 1 to a specific context
  swarmctl w 1:1 --context my-context

  # Install the workers 1 to 1 to all contexts that match a regex
  swarmctl w 1:1 --context 'my-.*'

  # Install the workers 1 to 1 to all contexts that match a regex and set the replicas
  swarmctl w 1:1 --context 'my-.*' --replicas 3

  # Install the workers 1 to 1 to all contexts that match a regex and set the node selector
  swarmctl w 1:1 --context 'my-.*' --node-selector '{key1: value1, key2: value2}'
  `
}

//-----------------------------------------------------------------------------
// InstallWorkerTelemetry
//-----------------------------------------------------------------------------

func InstallWorkerTelemetry(cmd *cobra.Command, args []string) error {

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// Parse the range
	start, end, err := util.ParseRange(args[0])
	if err != nil {
		return err
	}

	// Read the CRDs
	crds, err := Assets.ReadFile("assets/crds.yaml")
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

		// Print the context
		fmt.Printf("\n%s\n\n", name)

		// Loop through all CRDs
		for _, doc := range util.SplitYAML(bytes.NewBuffer(crds)) {
			if err := context.ApplyYaml(doc); err != nil {
				fmt.Printf("\nError: %s\n", err)
			}
		}

		// Loop trough all services
		for i := start; i <= end; i++ {

			fmt.Printf("\n")

			// Render the template
			docs, err := util.RenderTemplate(tmpl, struct {
				OnOff     string
				Namespace string
			}{
				OnOff:     args[1],
				Namespace: fmt.Sprintf("service-%d", i),
			})
			if err != nil {
				return err
			}

			// Loop through all yaml documents
			for _, doc := range docs {
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
