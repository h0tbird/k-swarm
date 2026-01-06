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
	"k8s.io/utils/ptr"
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
		if _, err = foo.Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, []byte(doc), metav1.PatchOptions{FieldManager: "swarmctl-manager", Force: ptr.To(true)}); err != nil {
			return fmt.Errorf("failed to create resource %s with GVR %v: %w", obj.GetName(), gvr, err)
		}
		fmt.Printf("  - %s/%s serverside-applied\n", resource.Kind, obj.GetName())
	}

	// Namespaced resources
	if resource.Namespaced {
		foo := c.DynCli.Resource(gvr).Namespace(namespace)
		if _, err = foo.Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, []byte(doc), metav1.PatchOptions{FieldManager: "swarmctl-manager", Force: ptr.To(true)}); err != nil {
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
	contexts := make([]string, 0, len(config.Contexts))
	for ctxName := range config.Contexts {
		contexts = append(contexts, ctxName)
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

	// Get the current context
	if regex == "" {
		config, err := clientcmd.LoadFromFile(filepath.Join(HomeDir, ".kube", "config"))
		if err != nil {
			return nil, err
		}
		return []string{config.CurrentContext}, nil
	}

	// List the contexts
	contexts, err := List()
	if err != nil {
		return nil, err
	}

	// Iterate over the contexts
	for _, ctxName := range contexts {

		// If the context matches the regex
		if match, err := regexp.MatchString(regex, ctxName); match && err == nil {
			matchingContexts = append(matchingContexts, ctxName)
		} else if err != nil {
			return nil, err
		}
	}

	// Return the matching contexts.
	return matchingContexts, nil
}

//-----------------------------------------------------------------------------
// GetClusterDomain reads the cluster domain from the CoreDNS ConfigMap.
// Falls back to "cluster.local" if unable to read or parse.
//-----------------------------------------------------------------------------

func (c *Context) GetClusterDomain(ctx context.Context) string {

	const defaultDomain = "cluster.local"

	// Define the GVR for ConfigMaps
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}

	// Get the coredns ConfigMap from kube-system namespace
	cm, err := c.DynCli.Resource(gvr).Namespace("kube-system").Get(ctx, "coredns", metav1.GetOptions{})
	if err != nil {
		return defaultDomain
	}

	// Get the Corefile data
	data, found, err := unstructured.NestedStringMap(cm.Object, "data")
	if err != nil || !found {
		return defaultDomain
	}

	corefile, ok := data["Corefile"]
	if !ok {
		return defaultDomain
	}

	// Parse the Corefile to find the kubernetes plugin line
	// Example: "kubernetes cluster.local in-addr.arpa ip6.arpa {"
	domain := parseClusterDomainFromCorefile(corefile)
	if domain == "" {
		return defaultDomain
	}

	return domain
}

//-----------------------------------------------------------------------------
// parseClusterDomainFromCorefile extracts the cluster domain from Corefile.
// Looks for pattern: kubernetes <domain> [in-addr.arpa ip6.arpa] {
//-----------------------------------------------------------------------------

func parseClusterDomainFromCorefile(corefile string) string {

	// Regex to match: kubernetes <domain> ...
	// The domain is typically the first argument after "kubernetes"
	re := regexp.MustCompile(`(?m)^\s*kubernetes\s+(\S+)`)
	matches := re.FindStringSubmatch(corefile)

	if len(matches) < 2 {
		return ""
	}

	domain := matches[1]

	// Validate it looks like a domain (contains at least one dot, no special chars except dots)
	if !strings.Contains(domain, ".") {
		return ""
	}

	// Clean up any trailing characters
	domain = strings.TrimSpace(domain)

	return domain
}
