package config

const (
	LOG_FILE                    = "poscln.log"
	LOG_FILE_MAX_SIZE           = 2
	LOG_FILE_MAX_AGE            = 30
	LOG_FILE_MAX_BACKUPS        = 0
	LOG_FOLDER           string = "./log"
	TMP_FOLDER           string = "./tmp"
)

var IsDevelopment = true
