package server

import "os"

type Config struct {
	DB *DBConfig
}

type DBConfig struct {
	User     string
	Password string
	Port     string
	Host     string
	DBname   string
}

func GetConfig() *Config {
	return &Config{
		DB: &DBConfig{
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Port:     os.Getenv("DB_PORT"),
			Host:     os.Getenv("DB_HOST"),
			DBname:   os.Getenv("DB_NAME"),
		},
	}
}
