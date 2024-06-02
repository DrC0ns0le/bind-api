package ansible

import (
	"strings"

	"github.com/DrC0ns0le/bind-api/rdb"
)

// Run deploy config playbook
func DeployConfig() error {

	// TODO: implement ansible playbook

	return nil
}

// Generate inventory ini from database config_key=servers, which is a comma seperated lists of ip addresses
func GenerateInventory(_bd *rdb.BindData) error {

	ips, err := _bd.Configs.Get("servers")
	if err != nil {
		return err
	}

	// convert ips to array slice by splitting commas
	IPS := append(IPS, strings.Split(ips, ",")...)

	// generate inventory ini file

	return nil
}
