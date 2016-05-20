package libvirt

import (
	"fmt"
	"path/filepath"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step creates the virtual disk that will be used as the
// hard drive for the virtual machine.
type stepCreateDisk struct{}

func (s *stepCreateDisk) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	path := filepath.Join(config.OutputDir, config.DiskName+".img")
	size := fmt.Sprintf("%dM", config.DiskSize)

	switch config.DomainType {
	case "kvm":
		ui.Say("Creating hard drive...")
		_, _, err := qemuImg("create", "-f", config.DiskType, path, size)
		if err != nil {
			err := fmt.Errorf("Error creating hard drive: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	case "lxc":
		ui.Say("Populate output dir...")
	}

	return multistep.ActionContinue
}

func (s *stepCreateDisk) Cleanup(state multistep.StateBag) {}
