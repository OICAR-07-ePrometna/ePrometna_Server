package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadConfig loads in program configuration should be a first thing called in the program
func LoadConfig() error {
	fmt.Println("Loading configuration")

	data, err := os.ReadFile(_CONFIG_FILE)
	if err != nil {
		return err
	}
	var conf AppConfiguration
	err = json.Unmarshal(data, &conf)
	if err != nil {
		return err
	}
	AppConfig = &conf

	fmt.Println("Configuration loaded in successfully")
	if AppConfig.IsDevelopment {
		fmt.Printf("AppConfig: %+v\n", AppConfig)
	}

	return nil
}
