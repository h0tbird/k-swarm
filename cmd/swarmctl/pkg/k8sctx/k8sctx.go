package k8sctx

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/trace"

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

var HomeDir string

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	var err error

	// Get the user's home directory
	if HomeDir, err = os.UserHomeDir(); err != nil {
		panic(err)
	}
}

//-----------------------------------------------------------------------------
// New
//-----------------------------------------------------------------------------

func New(name string) (*Context, error) {

	// Create the rest config
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: HomeDir + "/.kube/config"},
		&clientcmd.ConfigOverrides{CurrentContext: name},
	).ClientConfig()
	if err != nil {
		return nil, err
	}

	// Set the QPS and Burst
	config.QPS = 100
	config.Burst = 200

	// Create the dynamic client
	dynCli, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Create the discovery client
	disCli, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	// Return the context
	return &Context{
		Name:   name,
		Config: config,
		DynCli: dynCli,
		DisCli: disCli,
		MapGV:  map[string]*metav1.APIResourceList{},
	}, nil
}

//-----------------------------------------------------------------------------
// ApplyYaml
//-----------------------------------------------------------------------------

func (c *Context) ApplyYaml(doc string) error {

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
	resourceList, ok := c.MapGV[groupVersion]

	// If the key exists
	if !ok {
		resourceList, err = c.DisCli.ServerResourcesForGroupVersion(groupVersion)
		if err != nil {
			return fmt.Errorf("unable to get server resources for group version %s: %v", groupVersion, err)
		}
		c.MapGV[groupVersion] = resourceList
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
		foo := c.DynCli.Resource(gvr)
		if _, err = foo.Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, []byte(doc), metav1.PatchOptions{FieldManager: "swarmctl-manager", Force: pointer.Bool(true)}); err != nil {
			return fmt.Errorf("failed to create resource %s with GVR %v: %w", obj.GetName(), gvr, err)
		}
		fmt.Printf("  - %s/%s serverside-applied\n", resource.Kind, obj.GetName())
	}

	// Namespaced resources
	if resource.Namespaced {
		foo := c.DynCli.Resource(gvr).Namespace(namespace)
		if _, err = foo.Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, []byte(doc), metav1.PatchOptions{FieldManager: "swarmctl-manager", Force: pointer.Bool(true)}); err != nil {
			return fmt.Errorf("failed to apply resource %s with GVR %v: %w", obj.GetName(), gvr, err)
		}
		fmt.Printf("  - %s/%s serverside-applied\n", resource.Kind, obj.GetName())
	}

	// Return
	return nil
}

//-----------------------------------------------------------------------------
// List
//-----------------------------------------------------------------------------

func List() ([]string, error) {

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
// Filter
//-------------------------------------------------------------------------

func Filter(regex string) ([]string, error) {

	// Variables
	var matchingContexts []string

	// Return empty list if regex is empty
	if regex == "" {
		return matchingContexts, nil
	}

	// List the contexts
	contexts, err := List()
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
