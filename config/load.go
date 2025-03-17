package config

import "fmt"

// LoadConfig loads in program configuration should be a first thing called in the program
func LoadConfig() error {
	fmt.Println("Loading configuration")
	fmt.Println("Configuration loaded in successfully")
	return nil
}
