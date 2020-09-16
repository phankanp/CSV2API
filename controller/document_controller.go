package controller

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	"github.com/phankanp/csv-to-json/auth"
	"github.com/phankanp/csv-to-json/model"
	"github.com/phankanp/csv-to-json/response"
	"mime/multipart"
	"net/http"
	"sync"
)

func (server *Server) GetDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	apiKey := r.Context().Value("key").(string)
	username := vars["username"]
	docID := vars["id"]

	user := &model.User{}
	retrievedUser, err := user.AuthenticateUser(server.DB, username)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusUnauthorized)
		return
	}

	ok := auth.CheckPasswordHash(retrievedUser.AuthKey, apiKey)

	if !ok {
		err = errors.New("invalid api key")
		response.ErrorResponse(w, err, err.Error(), http.StatusUnauthorized)
		return
	}

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusUnauthorized)
		return
	}

	document := &model.Document{}

	d, err := document.GetDocumentByID(server.DB, uuid.Parse(docID))

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusInternalServerError)
		return
	}

	response.JsonResponse(w, http.StatusOK, d)
}

func (server *Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	sessionToken, err := auth.GetSessionToken(r)

	if err != nil {
		if err == http.ErrNoCookie {
			response.ErrorResponse(w, err, err.Error(), http.StatusUnauthorized)
			return
		}
		response.ErrorResponse(w, err, err.Error(), http.StatusBadRequest)
		return
	}

	userEmail, err := auth.GetUserEmailFromSessionToken(server.Cache, sessionToken)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusInternalServerError)
		return
	}

	if userEmail == "" {
		response.ErrorResponse(w, err, err.Error(), http.StatusUnauthorized)
		return
	}

	user := &model.User{}

	authenticatedUser, err := user.GetUserByEmail(server.DB, userEmail)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusInternalServerError)
		return
	}

	err = r.ParseMultipartForm(200000)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	formdata := r.MultipartForm

	files := formdata.File["multiplefiles"]
	titles := formdata.Value["title"]

	documents := make([]*model.Document, 0)

	resCh := make(chan *model.Document)
	errCh := make(chan error)
	doneCh := make(chan struct{})

	wg := sync.WaitGroup{}

	for i, _ := range files {
		file := files[i]
		fname := titles[i]
		wg.Add(1)

		go func(file *multipart.FileHeader, fname string, server *Server, authenticatedUser *model.User) {
			defer wg.Done()

			f, err := file.Open()

			if err != nil {
				errCh <- fmt.Errorf("cannot open file: %s", err)
				return
			}

			defer f.Close()

			doc := model.Document{}

			data, err := doc.CreateDocument(f, fname, server.DB, authenticatedUser)

			if err != nil {
				errCh <- err
			}

			resCh <- data
		}(file, fname, server, authenticatedUser)
	}

	go func() {
		wg.Wait()
		doneCh <- struct{}{}
	}()

	for {
		select {
		case err := <-errCh:
			response.ErrorResponse(w, err, err.Error(), http.StatusInternalServerError)
			return
		case data := <-resCh:
			documents = append(documents, data)
		case <-doneCh:
			println("test")
			response.JsonResponse(w, http.StatusOK, documents)
			return
		}
	}
}

func (server *Server) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	apiKey := r.Context().Value("key").(string)
	username := vars["username"]
	docID := vars["id"]

	user := &model.User{}
	retrievedUser, err := user.AuthenticateUser(server.DB, username)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusUnauthorized)
		return
	}

	ok := auth.CheckPasswordHash(retrievedUser.AuthKey, apiKey)

	if !ok {
		err = errors.New("invalid api key")
		response.ErrorResponse(w, err, err.Error(), http.StatusUnauthorized)
		return
	}

	document := &model.Document{}
	retrievedDocument, err := document.GetDocumentByID(server.DB, uuid.Parse(docID))

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusUnauthorized)
		return
	}

	if !uuid.Equal(retrievedDocument.UserID, retrievedUser.ID) {
		response.ErrorResponse(w, err, err.Error(), http.StatusUnauthorized)
		return
	}

	_, err = document.DeleteDocument(server.DB, uuid.Parse(docID))

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusBadRequest)
		return
	}

	response.JsonResponse(w, http.StatusOK, "")
}
