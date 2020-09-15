package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/phankanp/csv-to-json/config"
	"github.com/phankanp/csv-to-json/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	Router *mux.Router
	DB     *gorm.DB
	Cache  redis.Conn
}

func (server *Server) Initialize(config *config.Config) {
	var err error

	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
		config.DB.Host, config.DB.Port, config.DB.User, config.DB.DBname, config.DB.Password)

	server.DB, err = gorm.Open(postgres.Open(connectionString), &gorm.Config{})

	if err != nil {
		fmt.Printf("Failed to connect to Database")
		log.Fatal("Error:", err)
	} else {
		fmt.Println("Successfully connected to Database")
	}

	conn, err := redis.DialURL("redis://localhost")
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Successfully connected to redis")
	}

	server.Cache = conn
	server.DB.AutoMigrate(&model.User{})
	server.Router = mux.NewRouter()
	server.InitializeRoutes()

}

func (server *Server) Run(addr string) {
	fmt.Println("Listening to port 8080")
	log.Fatal(http.ListenAndServe(addr, server.Router))
}
