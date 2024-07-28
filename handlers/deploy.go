package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DrC0ns0le/bind-api/ansible"
	"github.com/DrC0ns0le/bind-api/rdb"
)

func GetDeployHandler(w http.ResponseWriter, r *http.Request) {
	status, err := (&rdb.Config{ConfigKey: "config_status"}).Find()
	if err != nil {
		responseBody := responseBody{
			Code:    1,
			Message: "Unable to retrieve deploy status",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseBody)
		return
	}
	if len(status) == 0 {
		responseBody := responseBody{
			Code:    2,
			Message: "Deploy status not found",
			Data:    nil,
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseBody)
		return
	}
	switch status[0].ConfigValue {
	case "awaiting_deployment":
		responseBody := responseBody{
			Code:    0,
			Message: "Awaiting deployment",
			Data:    true,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responseBody)
	default:
		responseBody := responseBody{
			Code:    0,
			Message: "Deployment complete",
			Data:    false,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responseBody)
	}

}
func DeployHandler(w http.ResponseWriter, r *http.Request) {
	output, err := ansible.DeployConfig()
	if err != nil {
		responseBody := responseBody{
			Code:    1,
			Message: "Unable to deploy changes",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseBody)
		return
	}

	responseBody := responseBody{
		Code:    0,
		Message: "Successfully deployed changes",
		Data:    output,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}
