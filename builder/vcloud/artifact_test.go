package vcloud

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestArtifact_Impl(t *testing.T) {
	var raw interface{}
	raw = &Artifact{}
	if _, ok := raw.(packer.Artifact); !ok {
		t.Fatalf("Artifact should be artifact")
	}
}

func TestArtifactString(t *testing.T) {
	dummyClient := VCloudClient{"", "", "", ""}
	a := &Artifact{"packer-foobar", 42, "https://vcloud.example.com/api", "Org A", "Catalog", dummyClient}
	expected := "A vApp template was created: 'packer-foobar' in catalog 'Catalog', org: 'Org A'"

	if a.String() != expected {
		t.Fatalf("artifact string should match: %v", expected)
	}
}
