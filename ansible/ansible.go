package ansible

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/DrC0ns0le/bind-api/rdb"
)

const playbook = "./ansible/deploy_config.yaml"
const inventory = "./ansible/inventory.ini"

// Run deploy config playbook
func DeployConfig(ctx context.Context) (string, error) {

	// check if playbook exists
	_, err := os.Stat(playbook)
	if os.IsNotExist(err) {
		return "", errors.New("playbook not found")
	}

	// check if inventory.ini exists
	_, err = os.Stat(inventory)
	if os.IsNotExist(err) {
		return "", errors.New("inventory not found")
	}

	// run playbook

	output, err := exec.CommandContext(ctx, "ansible-playbook", "-i", inventory, playbook).Output()
	if err != nil {
		return string(output), fmt.Errorf("failed to run ansible playbook: %w", err)
	}

	// set config deploy_status to deployed
	if err := (&rdb.Config{ConfigKey: "config_status", ConfigValue: "awaiting_deployment", Staging: false}).Update(ctx, "deployed"); err != nil {
		return "", fmt.Errorf("failed to update deploy_status: %w", err)
	}

	return string(output), nil
}

// Generate inventory ini from database config_key=servers, which is a comma seperated lists of ip addresses
func GenerateInventory() error {

	// hosts := []string{"10.1.1.109", "10.2.1.2", "10.3.1.25"}

	// TODO: implement ansible playbook

	return nil
}
