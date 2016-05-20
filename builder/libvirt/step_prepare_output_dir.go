package libvirt

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepPrepareOutputDir struct{}

func (stepPrepareOutputDir) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)
	isoPath := state.Get("iso_path").(string)

	if _, err := os.Stat(config.OutputDir); err == nil && config.PackerForce {
		ui.Say("Deleting previous output directory...")
		os.RemoveAll(config.OutputDir)
	}

	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	args := []string{"-C", config.OutputDir}
	switch config.DomainType {
	case "lxc":
		parts := strings.Split(isoPath, ".")
		compressor := parts[len(parts)-1:][0]
		switch compressor {
		case "gz":
			args = append(args, "-z")
		case "xz":
			args = append(args, "-J")
		case "bz2", "bzip2":
			args = append(args, "-j")
		}
		args = append(args, "-x", "-f")
		args = append(args, isoPath)

		cmd := exec.Command("tar", args...)
		buf, err := cmd.CombinedOutput()
		if err != nil {
			err := fmt.Errorf("Error extracting: %s %s", err, buf)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (stepPrepareOutputDir) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if cancelled || halted {
		config := state.Get("config").(*config)
		ui := state.Get("ui").(packer.Ui)

		ui.Say("Deleting output directory...")
		for i := 0; i < 5; i++ {
			err := os.RemoveAll(config.OutputDir)
			if err == nil {
				break
			}

			log.Printf("Error removing output dir: %s", err)
			time.Sleep(2 * time.Second)
		}
	}
}
