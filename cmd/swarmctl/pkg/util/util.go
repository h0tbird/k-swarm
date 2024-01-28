package util

import (

	// Stdlib
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"runtime/trace"
	"strings"

	// Community
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/pointer"
)

//-----------------------------------------------------------------------------
// Typedefs
//-----------------------------------------------------------------------------

type Context struct {
	Name   string
	Config *rest.Config
	DynCli *dynamic.DynamicClient
	DisCli *discovery.DiscoveryClient
	MapGV  map[string]*metav1.APIResourceList
}

//-----------------------------------------------------------------------------
// Globals
//-----------------------------------------------------------------------------

var (
	HomeDir  string
	SwarmDir string
)

type Client struct {
	dyn *dynamic.DynamicClient
	dis *discovery.DiscoveryClient
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	var err error

	// Get the user's home directory
	HomeDir, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	// Set the swarm directory
	SwarmDir = filepath.Join(HomeDir, ".swarmctl")
}

//-----------------------------------------------------------------------------
// ListKubeContexts
//-----------------------------------------------------------------------------

func ListKubeContexts() ([]string, error) {

	// Load the kubeconfig file
	config, err := clientcmd.LoadFromFile(filepath.Join(HomeDir, ".kube", "config"))
	if err != nil {
		return nil, err
	}

	// Iterate over the contexts
	var contexts []string
	for context := range config.Contexts {
		contexts = append(contexts, context)
	}

	// Return the contexts.
	return contexts, nil
}

//-------------------------------------------------------------------------
// FilterKubeContexts
//-------------------------------------------------------------------------

func FilterKubeContexts(regex string) ([]string, error) {

	// Variables
	var matchingContexts []string

	// Return empty list if regex is empty
	if regex == "" {
		return matchingContexts, nil
	}

	// List the contexts
	contexts, err := ListKubeContexts()
	if err != nil {
		return nil, err
	}

	// Iterate over the contexts
	for _, context := range contexts {

		// If the context matches the regex
		if match, err := regexp.MatchString(regex, context); match && err == nil {
			matchingContexts = append(matchingContexts, context)
		} else if err != nil {
			return nil, err
		}
	}

	// Return the matching contexts.
	return matchingContexts, nil
}

//-----------------------------------------------------------------------------
// ParseTemplate
//-----------------------------------------------------------------------------

func ParseTemplate(assets embed.FS, component string) (*template.Template, error) {

	// Variables
	var tmpl *template.Template
	var err error

	// Check for the ~/.swarmctl/<component>.goyaml file
	_, err = os.Stat(SwarmDir + "/" + component + ".goyaml")

	// Use the embedded template
	if os.IsNotExist(err) {
		tmpl, err = template.ParseFS(assets, "assets/"+component+".goyaml")
		if err != nil {
			return nil, err
		}
	} else if err == nil {
		tmpl, err = template.ParseFiles(SwarmDir + "/" + component + ".goyaml")
		if err != nil {
			return nil, err
		}
	}

	// Return
	return tmpl, nil
}

//-----------------------------------------------------------------------------
// RenderTemplate
//-----------------------------------------------------------------------------

func RenderTemplate(tmpl *template.Template, data any) ([]string, error) {

	// Render the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	// Split the YAML into docs
	var docs []string
	for _, doc := range strings.Split(buf.String(), "---") {
		if strings.TrimSpace(doc) != "" {
			docs = append(docs, doc)
		}
	}

	// Return
	return docs, nil
}

//-----------------------------------------------------------------------------
// ApplyYaml
//-----------------------------------------------------------------------------

func ApplyYaml(myCtx *Context, doc string) error {

	// Start a trace region
	defer trace.StartRegion(context.TODO(), "ApplyYaml").End()

	// Decode the YAML to an unstructured object
	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode([]byte(doc), nil, obj)
	if err != nil {
		return err
	}

	// Set the group version
	var groupVersion string
	if gvk.Group == "" {
		groupVersion = gvk.Version
	} else {
		groupVersion = gvk.Group + "/" + gvk.Version
	}

	// Get the resource list
	resourceList, err := myCtx.DisCli.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return fmt.Errorf("unable to get server resources for group version %s: %v", groupVersion, err)
	}

	// Find the correct resource
	var resource *metav1.APIResource
	for _, r := range resourceList.APIResources {
		if r.Kind == gvk.Kind {
			resource = &r
			break
		}
	}

	// Return an error if the resource was not found
	if resource == nil {
		return fmt.Errorf("resource type not found")
	}

	// Create the GVR
	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: resource.Name,
	}

	// Set the namespace
	namespace := obj.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	// Cluster-scoped resources
	if !resource.Namespaced {
		foo := myCtx.DynCli.Resource(gvr)
		if _, err = foo.Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, []byte(doc), metav1.PatchOptions{FieldManager: "swarmctl-manager", Force: pointer.Bool(true)}); err != nil {
			return fmt.Errorf("failed to create resource %s with GVR %v: %w", obj.GetName(), gvr, err)
		}
		fmt.Printf("  - %s/%s serverside-applied\n", resource.Kind, obj.GetName())
	}

	// Namespaced resources
	if resource.Namespaced {
		foo := myCtx.DynCli.Resource(gvr).Namespace(namespace)
		if _, err = foo.Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, []byte(doc), metav1.PatchOptions{FieldManager: "swarmctl-manager", Force: pointer.Bool(true)}); err != nil {
			return fmt.Errorf("failed to apply resource %s with GVR %v: %w", obj.GetName(), gvr, err)
		}
		fmt.Printf("  - %s/%s serverside-applied\n", resource.Kind, obj.GetName())
	}

	// Return
	return nil
}
