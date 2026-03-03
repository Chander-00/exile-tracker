package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	POEAPIBaseUrl          string
	FetchIntervalInMinutes int64
	DBPath                 string
	POBRoot                string
	LuajitPath             string
	SSHPort                string
	SSHHostKeyPath         string
	SSHAdminKey            string // authorized_keys format public key
	APIKey                 string
}

var Envs = initConfig()

func initConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Couldnt load env file")
	}
	return Config{
		Port:                   getEnv("PORT", ":3000"),
		POEAPIBaseUrl:          getEnv("POE_API_BASE_URL", "https://api.example.com"),
		FetchIntervalInMinutes: getEnvAsInt("FETCH_INTERVAL_IN_MINUTES", 30),
		DBPath:                 getEnv("DB_PATH", "./data.db"),
		POBRoot:                getEnv("POB_ROOT", "/home/alexander/dev/goofing/PathOfBuilding"),
		LuajitPath:             getEnv("LUAJIT_PATH", ""),
		SSHPort:                getEnv("SSH_PORT", ":2222"),
		SSHHostKeyPath:         getEnv("SSH_HOST_KEY_PATH", ".ssh/exile_tracker_ed25519"),
		SSHAdminKey:            getEnv("SSH_ADMIN_KEY", ""),
		APIKey:                 getEnv("API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fallback
		}
		return i
	}
	return fallback
}
