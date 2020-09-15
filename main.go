package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/phankanp/csv-to-json/config"
	"github.com/phankanp/csv-to-json/controller"
)

var s = controller.Server{}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbConfig := config.GetConfig()
	s.Initialize(dbConfig)

	s.Run(":8080")
}
