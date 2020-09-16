package controller

import "github.com/phankanp/csv-to-json/middleware"

func (server *Server) InitializeRoutes() {
	server.Router.HandleFunc("/login", server.Login).Methods("POST")
	server.Router.HandleFunc("/register", server.Register).Methods("POST")
	server.Router.HandleFunc("/upload", server.UploadHandler).Methods("POST")
	server.Router.HandleFunc("/documents/{username}/{id}", middleware.MiddlewareAuth(server.GetDocument)).Methods("GET")
	server.Router.HandleFunc("/documents/{username}/{id}", middleware.MiddlewareAuth(server.DeleteDocument)).Methods("DELETE")

}
