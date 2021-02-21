package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tamada/wildcat"
)

type multipartEntry struct {
	header *multipart.FileHeader
}

func (me *multipartEntry) Name() string {
	return me.header.Filename
}

func (me *multipartEntry) Open() (io.ReadCloser, error) {
	return me.header.Open()
}

func parseQueryParams(req *http.Request) *wildcat.ReadOptions {
	values := req.URL.Query()
	opts := &wildcat.ReadOptions{}
	params := strings.Join(values["readAs"], ",")
	if strings.Contains(params, "file-list") {
		opts.FileList = true
	}
	if strings.Contains(params, "no-extract") {
		opts.NoExtract = true
	}
	return opts
}

type myEntry struct {
	name   string
	reader io.ReadCloser
}

func (me *myEntry) Name() string {
	return me.name
}

func (me *myEntry) Open() (io.ReadCloser, error) {
	return me.reader, nil
}

func respond(rs *wildcat.ResultSet, err error, res http.ResponseWriter) {
	updateHeader(res)
	if err != nil {
		respondImpl(res, 400, []byte(fmt.Sprintf(`{"message":"%s"}`, err.Error())))
		return
	}
	buffer := bytes.NewBuffer([]byte{})
	printer := wildcat.NewPrinter(buffer, "json")
	rs.Print(printer)
	respondImpl(res, 200, buffer.Bytes())
}

func respondImpl(res http.ResponseWriter, statusCode int, message []byte) {
	res.WriteHeader(statusCode)
	res.Write(message)
}

func countsBody(res http.ResponseWriter, req *http.Request, opts *wildcat.ReadOptions) (*wildcat.ResultSet, error) {
	ec := wildcat.NewErrorCenter()
	fileName := req.URL.Query().Get("file-name")
	if fileName == "" {
		fileName = "<request>"
	}
	ds := wildcat.NewDataSink(wildcat.DefaultGenerator, ec)
	opts.HandleArg(&myEntry{name: fileName, reader: req.Body}, ds, nil)
	return ds.ResultSet(), nil
}

func counts(res http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	defer req.Body.Close()
	handlers := []struct {
		contentType string
		execFunc    func(http.ResponseWriter, *http.Request, *wildcat.ReadOptions) (*wildcat.ResultSet, error)
	}{
		{"application/x-www-form-urlencoded", countsBody},
		{"multipart/form-data", countsMultipartBody},
	}
	opts := parseQueryParams(req)
	for _, handler := range handlers {
		if strings.HasPrefix(contentType, handler.contentType) {
			rs, err := handler.execFunc(res, req, opts)
			respond(rs, err, res)
			break
		}
	}
}

func countsMultipartBody(res http.ResponseWriter, req *http.Request, reader *wildcat.ReadOptions) (*wildcat.ResultSet, error) {
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	entries := []wildcat.Entry{}
	for _, headers := range req.MultipartForm.File {
		for _, header := range headers {
			entries = append(entries, &multipartEntry{header: header})
		}
	}
	argf := wildcat.Argf{Entries: entries, Options: reader}
	ec := wildcat.NewErrorCenter()
	rs := argf.CountAll(wildcat.DefaultGenerator, ec)
	return rs, nil
}

func wrapHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

func updateHeader(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func optionsHandler(res http.ResponseWriter, req *http.Request) {
	updateHeader(res)
	res.Header().Set("Access-Control-Request-Method", "POST")
	res.Header().Set("Access-Control-Allow-Headers", "origin, accept, X-PINGOTHER, Content-Type")
	res.WriteHeader(200)
	res.Write([]byte{})
}

func createRestAPIServer() *mux.Router {
	router := mux.NewRouter()
	subRouter := router.PathPrefix("/wildcat/api/").Subrouter()
	subRouter.HandleFunc("/counts", counts).Methods("POST")
	subRouter.HandleFunc("/counts", optionsHandler).Methods("OPTIONS")
	router.PathPrefix("/wildcat/").Handler(fileServer())
	return router
}

func fileServer() http.Handler {
	dirs := []string{
		"docs",
		"/opts/wildcat/docs",
		"/usr/local/opts/wildcat/docs",
	}
	for _, dir := range dirs {
		if wildcat.ExistDir(dir) {
			return wrapHandler(http.StripPrefix("/wildcat/", http.FileServer(http.Dir(dir))))
		}
	}
	return nil
}
func (server *serverOptions) start(router *mux.Router) int {
	fmt.Printf("Listen server at port %d\n", server.port)
	fmt.Printf("start shutdown: %s\n", http.ListenAndServe(fmt.Sprintf(":%d", server.port), router))
	return 0
}

func (server *serverOptions) launchServer() int {
	router := createRestAPIServer()
	return server.start(router)
}
