package vcloud

import (
	"fmt"
	"log"
)

type Artifact struct {
	// The name of the template
	templateName string

	// The UID of the template
	templateUid uint

	// The base URL of the vCloud
	baseUrl string

	// The Organization where the template exists
	org string

	// The Catalog where the template is stored
	catalog string

	// The client for making API calls
	client VCloudClient
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Files() []string {
	// No files with vCloud
	return nil
}

func (a *Artifact) Id() string {
	// mimicing the aws builder
	return fmt.Sprintf("%s:%s:%s:%s", a.baseUrl, a.org, a.catalog, a.templateName)
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A vApp template was created: '%v' in catalog '%v', org: '%v'", a.templateName, a.catalog, a.org)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying vApp template: %d (%s)", a.templateUid, a.templateName)
	return a.client.DestroyTemplate(a.templateUid)
}
