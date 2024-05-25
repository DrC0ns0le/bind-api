package handlers

type Config struct {
	ID          int    `json:"-"`
	ConfigKey   string `json:"key"`
	ConfigValue string `json:"value"`
}
