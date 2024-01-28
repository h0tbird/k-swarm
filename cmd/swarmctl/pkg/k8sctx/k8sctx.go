package k8sctx

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"context"
	"fmt"
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

	// Local
	"github.com/octoroot/swarm/cmd/swarmctl/pkg/util"
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
// New
//-----------------------------------------------------------------------------

func New(name string) (*Context, error) {

	// Create the rest config
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: util.HomeDir + "/.kube/config"},
		&clientcmd.ConfigOverrides{CurrentContext: name},
	).ClientConfig()
	if err != nil {
		return nil, err
	}

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
	resourceList, err := c.DisCli.ServerResourcesForGroupVersion(groupVersion)
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
