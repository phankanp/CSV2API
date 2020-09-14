package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/phankanp/csv-to-json/server"
)

var s = server.Server{}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbConfig := server.GetConfig()

	s.Initialize(dbConfig)

}
