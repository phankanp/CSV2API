package server

import (
	"fmt"
	"log"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	Router *mux.Router
	DB     *gorm.DB
}

func (a *Server) Initialize(config *Config) {
	var err error

	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
		config.DB.Host, config.DB.Port, config.DB.User, config.DB.DBname, config.DB.Password)

	a.DB, err = gorm.Open(postgres.Open(connectionString), &gorm.Config{})

	if err != nil {
		fmt.Printf("Failed to connect to Database")
		log.Fatal("Error:", err)
	} else {
		fmt.Printf("Successfully connected to Database")
	}

	a.Router = mux.NewRouter()
}
