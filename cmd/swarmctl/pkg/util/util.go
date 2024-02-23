package util

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"bytes"
	"embed"
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
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

var SwarmDir string

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Get the user's home directory
	HomeDir, err := os.UserHomeDir()
	if err != nil {
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

//-----------------------------------------------------------------------------
// ParseRange
//-----------------------------------------------------------------------------

func ParseRange(arg string) (int, int, error) {

	// Split arg into start and end
	parts := strings.Split(arg, ":")
	if len(parts) != 2 {
		return 0, 0, errors.New("invalid range format. Please use the format start:end")
	}

	// Convert start and end to integers
	start, err1 := strconv.Atoi(parts[0])
	end, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, errors.New("invalid range. Both start and end should be integers")
	}

	// Return
	return start, end, nil
}
