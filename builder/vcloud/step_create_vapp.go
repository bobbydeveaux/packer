package vcloud

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepCreateVApp struct {
	vAppId string
}

func (s *stepCreateVApp) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*VCloudClient)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)
	//sshKeyId := state.Get("ssh_key_id").(uint)

	ui.Say("Creating vApp...")

	// Create the vApp based on configuration
	// TODO lots of more config here!
	vAppId, err := client.CreateVApp(c.TemplateName, c.SourceCatalog, c.SourceTemplate)

	if err != nil {
		err := fmt.Errorf("Error creating vApp: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this in cleanup
	s.vAppId = vAppId

	// Store the droplet id for later
	state.Put("vapp_id", vAppId)

	return multistep.ActionContinue
}

func (s *stepCreateVApp) Cleanup(state multistep.StateBag) {
	// If the vApp id isn't there, we probably never created it
	if s.vAppId == "" {
		return
	}

	client := state.Get("client").(*VCloudClient)
	ui := state.Get("ui").(packer.Ui)

	// Destroy the vApp we just created
	ui.Say("Destroying vApp...")

	err := client.DestroyVApp(s.vAppId)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying vApp: %v", s.vAppId))
	}

}
