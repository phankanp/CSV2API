package controller

import (
	"fmt"
	"github.com/phankanp/csv-to-json/model"
	"github.com/phankanp/csv-to-json/response"
	"mime/multipart"
	"net/http"
	"sync"
)

func (server *Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(200000)
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

		go func(file *multipart.FileHeader, fname string, server *Server) {
			defer wg.Done()

			f, err := file.Open()

			if err != nil {
				errCh <- fmt.Errorf("cannot open file: %s", err)
				return
			}

			defer f.Close()

			doc := model.Document{}

			data, err := doc.CreateDocument(f, fname, server.DB)

			if err != nil {
				errCh <- err
			}

			resCh <- data
		}(file, fname, server)
	}

	go func() {
		wg.Wait()
		doneCh <- struct{}{}
	}()

	for {
		select {
		case err := <-errCh:
			response.ErrorResponse(w, err, err.Error(), http.StatusInternalServerError)
		case data := <-resCh:
			documents = append(documents, data)
		case <-doneCh:
			println("test")
			response.JsonResponse(w, http.StatusOK, documents)
			return
		}
	}
}
