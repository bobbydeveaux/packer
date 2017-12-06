package vcloud

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepStop struct {
}

func (s *stepStop) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*VCloudClient)
	ui := state.Get("ui").(packer.Ui)
	id := state.Get("vapp_id").(string)

	ui.Say("Stopping vApp...")

	err := client.PowerOffVApp(id)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error stopping vApp: %v", id))
	}

	state.Put("started", false)

	return multistep.ActionContinue
}

func (s *stepStop) Cleanup(state multistep.StateBag) {
}
