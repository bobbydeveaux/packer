package vcloud

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepNATSSH struct {
}

func (s *stepNATSSH) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(config)

	if !config.NATSSH {
		return multistep.ActionContinue
	}

	client := state.Get("client").(*VCloudClient)
	ui := state.Get("ui").(packer.Ui)
	id := state.Get("vapp_id").(string)

	ui.Say("Setting up NAT for SSH in the Edge GW...")

	// 81.91.13.X DNAT Gw ext IP:81.92.13.200 origin port 2223<free port> <vm ip> <ssh port> TCP (enabled)

	// FW src:external src_port:any dest:81.91.13.200 dest_port:2223 TCP allow

	// Create the vApp based on configuration
	ip, src_port, err := client.DNAT(dest_port)

	if err != nil {
		err := fmt.Errorf("Error setting up DNAT in Edge GW: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("natssh_ip", ip)
	state.Put("natssh_src_port", ip)

	return multistep.ActionContinue
}

func (s *stepNATSSH) Cleanup(state multistep.StateBag) {
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
