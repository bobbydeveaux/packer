package vcloud

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepLogin struct{}

func (s *stepLogin) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*VCloudClient)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Login to get token...")

	err := client.Login()

	if err != nil {
		err := fmt.Errorf("Error login: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("logged_in", true)
	state.Put("client", client)

	return multistep.ActionContinue
}

func (s *stepLogin) Cleanup(state multistep.StateBag) {

	logged_in := state.Get("logged_in").(bool)
	if !logged_in {
		return
	}

	client := state.Get("client").(*VCloudClient)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Logout reverting token...")

	err := client.Logout()
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error logout: %s", err))
	}

}
