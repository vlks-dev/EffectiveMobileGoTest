package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBMaxIdle  time.Duration
	DBMaxConn  time.Duration
	APIBaseURL string
	ServerPort string
}

func LoadConfig() *Config {
	// Загружаем .env файл
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	dbmaxidle, err := strconv.Atoi(os.Getenv("DB_MAX_IDLE"))
	if err != nil {
		panic(err)
	}
	dbmaxconn, err := strconv.Atoi(os.Getenv("DB_MAX_CONN"))
	if err != nil {
		panic(err)
	}

	return &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBMaxIdle:  time.Duration(dbmaxidle),
		DBMaxConn:  time.Duration(dbmaxconn),
		APIBaseURL: os.Getenv("API_BASE_URL"),
		ServerPort: os.Getenv("SERVER_PORT"),
	}
}
