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
