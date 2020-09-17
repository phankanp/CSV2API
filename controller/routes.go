package controller

import "github.com/phankanp/csv-to-json/middleware"

func (server *Server) InitializeRoutes() {
	server.Router.HandleFunc("/login", server.Login).Methods("POST")
	server.Router.HandleFunc("/register", server.Register).Methods("POST")
	server.Router.HandleFunc("/upload", server.UploadHandlerConcurrent).Methods("POST")
	server.Router.HandleFunc("/uploadLinear", server.UploadHandler).Methods("POST")
	server.Router.HandleFunc("/{username}/documents", middleware.MiddlewareAuth(server.GetDocuments)).Methods("GET")
	server.Router.HandleFunc("/{username}/documents/{id}", middleware.MiddlewareAuth(server.GetDocument)).Methods("GET")
	server.Router.HandleFunc("/{username}/documents/{id}", middleware.MiddlewareAuth(server.DeleteDocument)).Methods("DELETE")
	server.Router.HandleFunc("/{username}/documents/{docID}/rows", middleware.MiddlewareAuth(server.SearchRows)).Queries("column", "{column}", "data", "{data}").Methods("GET")
	server.Router.HandleFunc("/{username}/documents/{docID}/rows", middleware.MiddlewareAuth(server.GetDocumentRows)).Methods("GET")
	server.Router.HandleFunc("/{username}/documents/{docID}/rows", middleware.MiddlewareAuth(server.CreateDocumentRow)).Methods("POST")
	server.Router.HandleFunc("/{username}/documents/{docID}/rows/{rowID}", middleware.MiddlewareAuth(server.GetDocumentRow)).Methods("GET")
	server.Router.HandleFunc("/{username}/documents/{docID}/rows/{rowID}", middleware.MiddlewareAuth(server.UpdateDocumentRow)).Methods("PUT")
	server.Router.HandleFunc("/{username}/documents/{docID}/rows/{rowID}", middleware.MiddlewareAuth(server.DeleteDocumentRow)).Methods("DELETE")
}
