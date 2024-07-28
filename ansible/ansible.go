package ansible

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

const playbook = "deploy_config.yml"
const inventory = "inventory.ini"

// Run deploy config playbook
func DeployConfig() error {

	// check if playbook exists
	_, err := os.Stat(playbook)
	if os.IsNotExist(err) {
		return errors.New("playbook not found")
	}

	// check if inventory.ini exists
	_, err = os.Stat(inventory)
	if os.IsNotExist(err) {
		return errors.New("inventory not found")
	}

	// run playbook
	if err = exec.Command("ansible-playbook", "-i", inventory, playbook).Run(); err != nil {
		return fmt.Errorf("failed to run ansible playbook: %w", err)
	}

	return nil
}

// Generate inventory ini from database config_key=servers, which is a comma seperated lists of ip addresses
func GenerateInventory() error {

	// hosts := []string{"10.1.1.109", "10.2.1.2", "10.3.1.25"}

	// TODO: implement ansible playbook

	return nil
}
