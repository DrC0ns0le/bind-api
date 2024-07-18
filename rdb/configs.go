package rdb

type Config struct {
	ConfigKey   string
	ConfigValue string
}

func (c *Config) GetAll() ([]Config, error) {
	rows, err := db.Query("SELECT config_key, config_value FROM bind_dns.configs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []Config
	for rows.Next() {
		var config Config
		err := rows.Scan(&config.ConfigKey, &config.ConfigValue)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func (c *Config) Find() error {
	row := db.QueryRow("SELECT config_value FROM bind_dns.configs WHERE config_key = ?", c.ConfigKey)
	err := row.Scan(&c.ConfigValue)
	if err != nil {
		return err
	}
	return nil
}
