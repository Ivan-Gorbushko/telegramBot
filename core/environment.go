package core

import (
	"errors"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

const ENV_PROD = "prod"
const ENV_DEV = "dev"

var IsWorking bool = false

var Config Environment

type Environment struct {
	Env string
	BotToken string
	PingTimeout int
	MongodbUri string
	LardiApiUrl string
	LardiSecretKey string
	ScanTimeout int
	InitialTime int64
	ContactId string
	FilterPageUrl string
}

// init is invoked before main()
func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	Config.BotToken = GetEnvData("bot_token", "")
	Config.Env = GetEnvData("env", "dev")
	Config.PingTimeout, _ = strconv.Atoi(GetEnvData("ping_timeout", "1500"))
	Config.MongodbUri = GetEnvData("mongodb_uri", "mongodb://localhost:27017")
	Config.LardiApiUrl = GetEnvData("lardi_api_url", "")
	Config.LardiSecretKey = GetEnvData("lardi_secret_key", "")
	Config.ScanTimeout, _ = strconv.Atoi(GetEnvData("scan_timeout", "60"))
	Config.InitialTime, _ = strconv.ParseInt(GetEnvData("initial_time", "0"),10,64)
	Config.ContactId = GetEnvData("contact_id", "")
	Config.FilterPageUrl = GetEnvData("filter_page_url", "")
}

// Simple helper function to read an environment or return a default value
func GetEnvData(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func (c *Environment) IsProd() bool {
	if Config.Env == ENV_PROD {
		return true
	} else if Config.Env == ENV_DEV {
		return false
	} else {
		//err := errors.InternalError.Error("Env unknown")
		err := errors.New("Unknown environment")
		panic (err)
	}
}
