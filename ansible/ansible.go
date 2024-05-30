package ansible

import "github.com/DrC0ns0le/bind-api/rdb"

// Run deploy config playbook
func DeployConfig() error {

	// TODO: implement ansible playbook

	return nil
}

// Generate inventory ini from database config_key=servers, which is a comma seperated lists of ip addresses
func GenerateInventory(_bd *rdb.BindData) error {

	ips, err := _bd.Config.Get("servers")
	if err != nil {
		return err
	}

	// convert ips to array slice by splitting comma
	


	return nil
}
