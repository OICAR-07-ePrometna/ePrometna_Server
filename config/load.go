package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// LoadConfig loads in program configuration should be a first thing called in the program
func LoadConfig() error {
	fmt.Println("Loading configuration")

	if err := godotenv.Load(); err != nil {
		fmt.Printf("Can't load config using real env\n")
		fmt.Printf("Env load err = %+v\n", err)
	}
	if err := loadData(); err != nil {
		fmt.Printf("Error loading config, err = %+v\n", err)
		return err
	}

	fmt.Println("Configuration loaded successfully")
	return nil
}

func loadData() error {
	conf := &AppConfiguration{}
	conf.Env = LoadEnv()
	conf.DbConnection = loadString("DB_CONN")
	conf.AccessKey = loadString("ACCESS_KEY")
	conf.RefreshKey = loadString("REFRESH_KEY")
	conf.Port = loadInt("PORT")

	if conf.AccessKey == "" {
		return fmt.Errorf("ACCESS_KEY environment variable is required")
	}
	if conf.RefreshKey == "" {
		return fmt.Errorf("REFRESH_KEY environment variable is required")
	}

	AppConfig = conf
	return nil
}

func loadInt(name string) int {
	rez := os.Getenv(name)
	if rez == "" {
		fmt.Printf("Env variable %s is empty\n", name)
	}
	num, err := strconv.Atoi(rez)
	if err != nil {
		fmt.Printf("Failed to parse int %s, will use default (8080)\n", rez)
		return 8080
	}

	return num
}

func loadString(name string) string {
	rez := os.Getenv(name)
	if rez == "" {
		fmt.Printf("Env variable %s is empty\n", name)
	}
	return rez
}

func LoadEnv() environment {
	name := "ENV"
	rez := os.Getenv(name)
	if rez == "" {
		fmt.Printf("variable ENV is empty\n")
	}

	switch rez {
	case Dev:
		fmt.Printf("Running in %s environment \n", Dev)
		return Dev

	case Prod:
		fmt.Printf("Running in %s environment \n", Prod)
		return Prod

	case Test:
		fmt.Printf("Running in %s environment \n", Test)
		return Test

	default:
		fmt.Printf("Bad ENV value (%s) must be: (%s,%s,%s) \n", rez, Prod, Dev, Test)
		return Prod
	}
}
