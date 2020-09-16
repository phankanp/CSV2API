package controller

func (server *Server) InitializeRoutes() {
	server.Router.HandleFunc("/login", server.Login).Methods("POST")
	server.Router.HandleFunc("/register", server.Register).Methods("POST")
	server.Router.HandleFunc("/upload", server.UploadHandler).Methods("POST")
}
