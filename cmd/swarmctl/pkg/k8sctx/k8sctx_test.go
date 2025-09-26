package k8sctx

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import "testing"

//-----------------------------------------------------------------------------
// BenchmarkApplyYaml
//-----------------------------------------------------------------------------

func BenchmarkApplyYaml(b *testing.B) {

	// Setup
	c, err := New("kind-foo-1")
	if err != nil {
		b.Fatal(err)
	}

	// YAML
	doc := `
---
apiVersion: v1
kind: Namespace
metadata:
  labels:
    istio-injection: enabled
  name: informer
`

	// Reset timer
	b.ResetTimer()

	// Run
	for i := 0; i < b.N; i++ {
		if err := c.ApplyYaml(doc); err != nil {
			b.Fatal(err)
		}
	}
}
