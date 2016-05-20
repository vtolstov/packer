package libvirt

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepCreateNetwork struct{}

func (stepCreateNetwork) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	netTemplate := `
<network>
  <name>{{ .VMName }}</name>
  <forward mode='nat'/>
<!--  <bridge name='packer-{{ .VMName }}' stp='on' delay='0' /> -->
  <ip address='10.0.2.2' netmask='255.255.255.0'>
    <dhcp>
      <range start='10.0.2.15' end='10.0.2.254' />
    </dhcp>
  </ip>
</network>
`

	type netInfo struct {
		VMName string
	}

	ni := &netInfo{
		VMName: config.VMName,
	}

	netContents, err := config.tpl.Process(netTemplate, ni)
	if err != nil {
		err := fmt.Errorf("Error procesing network template: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	f, err := ioutil.TempFile("", "packer-")
	if err != nil {
		err := fmt.Errorf("Error creating network: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	defer f.Close()
	defer os.Remove(f.Name())

	_, err = f.Write([]byte(netContents))
	if err != nil {
		err := fmt.Errorf("Error creating network: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	f.Sync()
	_, _, err = virsh("-c", config.URI, "net-create", f.Name())
	if err != nil {
		err := fmt.Errorf("Error creating network: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (stepCreateNetwork) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	_, _, err := virsh("-c", config.URI, "net-info", config.VMName)
	if err == nil {
		_, _, err := virsh("-c", config.URI, "net-destroy", config.VMName)
		if err != nil {
			ui.Error(fmt.Sprintf("Error destroying network: %s", err))
		}
	}
}
