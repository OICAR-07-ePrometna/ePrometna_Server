package config

const (
	_CONFIG_FILE                = "ePrometna.json"
	LOG_FILE                    = "ePrometna.log"
	LOG_FILE_MAX_SIZE           = 2
	LOG_FILE_MAX_AGE            = 30
	LOG_FILE_MAX_BACKUPS        = 0
	LOG_FOLDER           string = "./log"
	TMP_FOLDER           string = "./tmp"
)

// AppConfig is struct that contains basic app configuration variables
var AppConfig *AppConfiguration = nil

type AppConfiguration struct {
	// IsDevelopment describes the environment
	IsDevelopment bool
	Env           environment
	Port          int
	DbConnection  string
	AccessKey     string
	RefreshKey    string
}

type environment = string

const (
	Dev  = "dev"
	Test = "test"
	Prod = "prod"
)
