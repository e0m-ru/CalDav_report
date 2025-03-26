package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type yacaldavConfig struct {
	YAUSER,
	CALPWD,
	YACAL,
	DAVPWD string
}

type net struct {
	Timeout time.Duration
}

type logConfig struct {
	File  string
	Level int
}

type reportConfig struct {
}

type Config struct {
	Report    reportConfig
	ProductID string
	AppName   string
	YaAuth    yacaldavConfig
	Net       net
	Logging   logConfig
}

// New returns a new Config struct
func LoadConifg() *Config {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		panic("No .env file found")
	}
	portStr := getEnv("LOGLEVEL", "0")
	loglevel, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Error converting LOGLEVEL to integer: %v", err)
	}
	return &Config{
		Report: reportConfig{},
		YaAuth: yacaldavConfig{
			YAUSER: getEnv("YAUSER", "user@yandex.ru"),
			CALPWD: getEnv("CALPWD", "PA$$w0rD"),
			YACAL:  getEnv("YACAL", "https://caldav.yandex.ru"),
		},
		Net: net{
			Timeout: time.Millisecond * 3000,
		},
		ProductID: getEnv("ProductID", "ProductID"),
		AppName:   getEnv("AppName", "AppName"),
		Logging: logConfig{
			File:  getEnv("LOGFILE", "app.log"),
			Level: loglevel,
		},
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
