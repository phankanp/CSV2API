package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"runtime"
	"sync"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	"github.com/phankanp/csv-to-json/auth"
	"github.com/phankanp/csv-to-json/helper"
	"github.com/phankanp/csv-to-json/model"
	"github.com/phankanp/csv-to-json/response"
)

func (server *Server) GetDocuments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	apiKey := r.Context().Value("key").(string)
	username := vars["username"]

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

	d, err := document.GetDocuments(server.DB, user.ID)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusInternalServerError)
		return
	}

	response.JsonResponse(w, http.StatusOK, d)
}

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

	document := &model.Document{}

	d, err := document.GetDocumentByID(server.DB, uuid.Parse(docID))

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusInternalServerError)
		return
	}

	response.JsonResponse(w, http.StatusOK, d)
}

func (server *Server) UploadHandlerConcurrent(w http.ResponseWriter, r *http.Request) {
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
	filesCh := make(chan map[string]*multipart.FileHeader)

	wg := sync.WaitGroup{}
	go func() {
		defer close(filesCh)
		for i, _ := range files {
			m := make(map[string]*multipart.FileHeader)
			m[titles[i]] = files[i]
			filesCh <- m
		}
	}()
	for i := 0; i < runtime.NumCPU(); i++ {
		var file *multipart.FileHeader
		var fname string

		wg.Add(1)

		go func(file *multipart.FileHeader, fname string, server *Server, authenticatedUser *model.User) {
			defer wg.Done()
			for m := range filesCh {
				for key, val := range m {
					file := val
					fname := key

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
				}
			}
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

	for i, _ := range files {
		file := files[i]
		fname := titles[i]

		f, err := file.Open()
		defer f.Close()

		if err != nil {
			return
		}

		doc := model.Document{}

		data, err := doc.CreateDocument(f, fname, server.DB, authenticatedUser)

		documents = append(documents, data)

	}
	response.JsonResponse(w, http.StatusOK, documents)
}

func (server *Server) GetDocumentRows(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	apiKey := r.Context().Value("key").(string)
	username := vars["username"]
	docID := vars["docID"]

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

	row := &model.Row{}

	rows, err := row.GetAllRowsByDocument(server.DB, uuid.Parse(docID))

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusInternalServerError)
		return
	}

	response.JsonResponse(w, http.StatusOK, rows)
}

func (server *Server) CreateDocumentRow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	apiKey := r.Context().Value("key").(string)
	username := vars["username"]
	docID := vars["docID"]

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

	newRow := model.Row{}
	rowData := model.JSONB{}
	err = json.NewDecoder(r.Body).Decode(&rowData)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	headers, err := document.GetDocumentHeaders(server.DB)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ok = helper.CompareHeaders(rowData, headers)

	if !ok {
		err := errors.New("data keys do not match csv headers")
		response.ErrorResponse(w, err, err.Error(), http.StatusBadRequest)
		return
	}

	createdRow, err := newRow.CreateRow(server.DB, uuid.Parse(docID), rowData)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusInternalServerError)
		return
	}

	response.JsonResponse(w, http.StatusOK, createdRow)
}

func (server *Server) GetDocumentRow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	apiKey := r.Context().Value("key").(string)
	username := vars["username"]
	docID := vars["docID"]
	rowID, err := helper.IntFromString(vars["rowID"])

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusInternalServerError)
	}

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

	row := &model.Row{}

	retrievedRow, err := row.GetRowByID(server.DB, uuid.Parse(docID), uint(rowID))

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusInternalServerError)
		return
	}

	response.JsonResponse(w, http.StatusOK, retrievedRow)
}

func (server *Server) UpdateDocumentRow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	apiKey := r.Context().Value("key").(string)
	username := vars["username"]
	docID := vars["docID"]
	rowID, err := helper.IntFromString(vars["rowID"])

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

	row := &model.Row{}

	retrievedRow, err := row.GetRowByID(server.DB, uuid.Parse(docID), uint(rowID))

	updateRow := model.Row{}
	rowData := model.JSONB{}
	err = json.NewDecoder(r.Body).Decode(&rowData)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	headers, err := document.GetDocumentHeaders(server.DB)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ok = helper.CompareHeaders(rowData, headers)

	if !ok {
		err := errors.New("data keys do not match csv headers")
		response.ErrorResponse(w, err, err.Error(), http.StatusBadRequest)
		return
	}

	updateRow.ID = retrievedRow.ID

	updatedRow, err := updateRow.UpdateRow(server.DB, rowData)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusInternalServerError)
		return
	}

	response.JsonResponse(w, http.StatusOK, updatedRow)
}

func (server *Server) DeleteDocumentRow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	apiKey := r.Context().Value("key").(string)
	username := vars["username"]
	docID := vars["docID"]
	rowID, err := helper.IntFromString(vars["rowID"])

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

	row := &model.Row{}

	_, err = row.DeleteRow(server.DB, uuid.Parse(docID), uint(rowID))

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusBadRequest)
		return
	}

	response.JsonResponse(w, http.StatusOK, "")
}

func (server *Server) SearchRows(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	apiKey := r.Context().Value("key").(string)
	username := vars["username"]
	docID := vars["docID"]
	headerInput := vars["column"]
	dataInput := vars["data"]

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

	row := &model.Row{}
	println("****************test3************")
	rows, err := row.SearchRows(server.DB, uuid.Parse(docID), headerInput, dataInput)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusInternalServerError)
		return
	}

	response.JsonResponse(w, http.StatusOK, rows)
}
