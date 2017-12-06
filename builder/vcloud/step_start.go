package vcloud

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepStart struct {
}

func (s *stepStart) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*VCloudClient)
	ui := state.Get("ui").(packer.Ui)
	id := state.Get("vapp_id").(string)

	ui.Say("Starting vApp...")

	// Create the vApp based on configuration
	err := client.StartVApp(id)

	if err != nil {
		err := fmt.Errorf("Error creating vApp: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("started", true)

	return multistep.ActionContinue
}

func (s *stepStart) Cleanup(state multistep.StateBag) {
	id := state.Get("vapp_id").(string)
	if id == "" {
		return
	}

	// TODO: Fix!
	//	client := state.Get("client").(*VCloudClient)
	//	ui := state.Get("ui").(packer.Ui)
	//
	//	// Destroy the vApp we just created
	//	ui.Say("Ensure vApp Stopped...")
	//
	//	err := client.PowerOffVApp(id)
	//	if err != nil {
	//		ui.Error(fmt.Sprintf(
	//			"Error stopping vApp: %v", id))
	//	}
}
