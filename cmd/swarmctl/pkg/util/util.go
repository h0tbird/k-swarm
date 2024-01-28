package util

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"bytes"
	"embed"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	// Community
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
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

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	var err error

	// Get the user's home directory
	if HomeDir, err = os.UserHomeDir(); err != nil {
		panic(err)
	}

	// Set the swarm directory
	SwarmDir = filepath.Join(HomeDir, ".swarmctl")
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
