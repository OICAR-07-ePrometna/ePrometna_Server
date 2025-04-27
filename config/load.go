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
	loadData()
	return nil
}

func loadData() {
	conf := &AppConfiguration{}
	conf.IsDevelopment = isDevEnvironment()
	conf.DbConnection = loadString("DB_CONN")
	conf.AccessKey = loadString("ACCESS_KEY")
	conf.RefreshKey = loadString("REFRESH_KEY")
	conf.Port = loadInt("PORT")

	AppConfig = conf
}

func loadInt(name string) int {
	rez := os.Getenv(name)
	if rez == "" {
		fmt.Printf("Env variable %s is empty\n", name)
	}
	num, err := strconv.Atoi(rez)
	if err != nil {
		fmt.Printf("Failed to parse int %s, will use default (0)\n", rez)
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

func isDevEnvironment() bool {
	name := "APP_ENV"
	rez := os.Getenv(name)
	if rez == "" {
		fmt.Printf("Env variable %s is empty\n", name)
	}
	return rez == "development"
}
