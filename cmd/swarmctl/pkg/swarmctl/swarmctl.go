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
	version, _ := cmd.Flags().GetString("version")
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
		Version:      version,
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
// GenerateWorker outputs the worker manifest
//-----------------------------------------------------------------------------

func GenerateWorker(cmd *cobra.Command, args []string) error {

	// Get the flags
	replicas, _ := cmd.Flags().GetInt("replicas")
	nodeSelector, _ := cmd.Flags().GetString("node-selector")
	version, _ := cmd.Flags().GetString("version")
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
			Version:      version,
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
		return errors.New("aborted")
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
	version, _ := cmd.Flags().GetString("version")
	imageTag, _ := cmd.Flags().GetString("image-tag")

	// Set the error prefix
	cmd.SetErrPrefix("\nError:")

	// Parse the template
	tmpl, err := util.ParseTemplate(Assets, "informer")
	if err != nil {
		return err
	}

	// Loop through all contexts
	for name, context := range Contexts {

		// Print the context
		fmt.Printf("\n%s\n\n", name)

		// Render the template
		docs, err := util.RenderTemplate(tmpl, struct {
			Replicas     int
			NodeSelector string
			Version      string
			ImageTag     string
		}{
			Replicas:     replicas,
			NodeSelector: nodeSelector,
			Version:      version,
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
// InstallWorker
//-----------------------------------------------------------------------------

func InstallWorker(cmd *cobra.Command, args []string) error {

	// Get the flags
	replicas, _ := cmd.Flags().GetInt("replicas")
	nodeSelector, _ := cmd.Flags().GetString("node-selector")
	version, _ := cmd.Flags().GetString("version")
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

	// Loop through all contexts
	for name, context := range Contexts {

		// Print the context
		fmt.Printf("\n%s\n", name)

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
				Version:      version,
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
