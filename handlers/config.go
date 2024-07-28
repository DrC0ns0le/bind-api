package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/DrC0ns0le/bind-api/rdb"
)

type Config struct {
	ID          int    `json:"-"`
	ConfigKey   string `json:"key"`
	ConfigValue string `json:"value"`
	CreatedAt   uint64 `json:"created_at"`
	ModifiedAt  uint64 `json:"modified_at"`
	DeletedAt   uint64 `json:"deleted_at,omitempty"`
	Staging     bool   `json:"staging"`

	ConfigOld string `json:"old_value,omitempty"`
}

func GetConfigsHandler(w http.ResponseWriter, r *http.Request) {
	configs, err := new(rdb.Config).Get()
	if err != nil {
		responseBody := responseBody{
			Code:    1,
			Message: "Unable to retrieve configs",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseBody)
		return
	}

	var configsList []Config
	for _, config := range configs {
		configsList = append(configsList, Config{
			ConfigKey:   config.ConfigKey,
			ConfigValue: config.ConfigValue,
			CreatedAt:   uint64(config.CreatedAt.Unix()),
			ModifiedAt:  uint64(config.ModifiedAt.Unix()),
			Staging:     config.Staging,
		})
	}

	responseBody := responseBody{
		Code:    0,
		Message: "Configs retrieved successfully",
		Data:    configsList,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}

func GetConfigHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("config_key")
	config, err := (&rdb.Config{ConfigKey: key}).Find()
	if err != nil {
		responseBody := responseBody{
			Code:    1,
			Message: "Unable to retrieve config",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseBody)
		return
	}
	if len(config) == 0 {
		responseBody := responseBody{
			Code:    2,
			Message: "Config key not found",
			Data:    nil,
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseBody)
		return
	}

	var configsList []Config
	for _, config := range config {
		configsList = append(configsList, Config{
			ConfigKey:   config.ConfigKey,
			ConfigValue: config.ConfigValue,
			Staging:     config.Staging,
			CreatedAt:   uint64(config.CreatedAt.Unix()),
			ModifiedAt:  uint64(config.ModifiedAt.Unix()),
		})
	}

	responseBody := responseBody{
		Code:    0,
		Message: "Config retrieved successfully",
		Data:    configsList,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}

func CreateConfigHandler(w http.ResponseWriter, r *http.Request) {
	c := new(Config)
	err := json.NewDecoder(r.Body).Decode(c)
	if err != nil {
		responseBody := responseBody{
			Code:    1,
			Message: "Unable to decode request body",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseBody)
		return
	}

	config := &rdb.Config{ConfigKey: c.ConfigKey, ConfigValue: c.ConfigValue, Staging: c.Staging}
	if err = config.Create(); err != nil {
		responseBody := responseBody{
			Code:    2,
			Message: "Unable to create config",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseBody)
		return
	}

	responseBody := responseBody{
		Code:    0,
		Message: "Config created successfully",
		Data: &Config{
			ConfigKey:   config.ConfigKey,
			ConfigValue: config.ConfigValue,
			Staging:     c.Staging,
			CreatedAt:   uint64(config.CreatedAt.Unix()),
		},
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseBody)
}

func UpdateConfigHandler(w http.ResponseWriter, r *http.Request) {
	c := new(Config)
	err := json.NewDecoder(r.Body).Decode(c)
	if err != nil {
		responseBody := responseBody{
			Code:    1,
			Message: "Unable to decode request body",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseBody)
		return
	}

	config := &rdb.Config{ConfigKey: c.ConfigKey, ConfigValue: c.ConfigValue, Staging: c.Staging}
	if err = config.Update(c.ConfigValue); err != nil {
		responseBody := responseBody{
			Code:    2,
			Message: "Unable to update config",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseBody)
		return
	}

	responseBody := responseBody{
		Code:    0,
		Message: "Config updated successfully",
		Data: &Config{
			ConfigKey:   config.ConfigKey,
			ConfigValue: config.ConfigValue,
			Staging:     c.Staging,
			CreatedAt:   uint64(config.CreatedAt.Unix()),
			ModifiedAt:  uint64(config.ModifiedAt.Unix()),
		},
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}

func DeleteConfigHandler(w http.ResponseWriter, r *http.Request) {
	c := new(Config)
	err := json.NewDecoder(r.Body).Decode(c)
	if err != nil {
		responseBody := responseBody{
			Code:    1,
			Message: "Unable to decode request body",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseBody)
		return
	}

	config := &rdb.Config{ConfigKey: c.ConfigKey, ConfigValue: c.ConfigValue, Staging: c.Staging}
	if err = config.Delete(); err != nil {
		responseBody := responseBody{
			Code:    2,
			Message: "Unable to delete config",
			Data:    err.Error(),
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseBody)
		return
	}

	responseBody := responseBody{
		Code:    0,
		Message: "Config deleted successfully",
		Data:    nil,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseBody)
}
