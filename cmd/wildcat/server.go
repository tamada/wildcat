package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/tamada/wildcat/errors"

	"github.com/gorilla/mux"
	"github.com/tamada/wildcat"

	"github.com/tamada/wildcat/logger"
)

type multipartEntry struct {
	header *multipart.FileHeader
	index  *wildcat.Order
	reader io.ReadSeekCloser
}

func (me *multipartEntry) Name() string {
	return me.header.Filename
}

func (me *multipartEntry) Open() (io.ReadCloser, error) {
	if me.reader != nil {
		me.reader.Seek(0, 0)
		return me.reader, nil
	}
	reader, err := me.header.Open()
	if err == nil {
		me.reader = wildcat.NewReadSeekCloser(reader)
	}
	return me.reader, nil
}

func (me *multipartEntry) Index() *wildcat.Order {
	return me.index
}

func (me *multipartEntry) Reindex(newIndex int) {
	// me.index = newIndex
}

func (me *multipartEntry) Count(generator wildcat.Generator) *wildcat.Either {
	return wildcat.CountDefault(me, generator())
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
	order  *wildcat.Order
	reader io.ReadSeekCloser
}

func (me *myEntry) Name() string {
	return me.name
}

func (me *myEntry) Open() (io.ReadCloser, error) {
	me.reader.Seek(0, 0)
	return me.reader, nil
}

func (me *myEntry) Index() *wildcat.Order {
	if me.order == nil {
		me.order = wildcat.NewOrder()
	}
	return me.order
}

func (me *myEntry) Reindex(newIndex int) {
}

func (me *myEntry) Count(generator wildcat.Generator) *wildcat.Either {
	return wildcat.CountDefault(me, generator())
}

func createResultJSON(rs *wildcat.ResultSet, sizer wildcat.Sizer) []byte {
	buffer := bytes.NewBuffer([]byte{})
	printer := wildcat.NewPrinter(buffer, "json", sizer)
	rs.Print(printer)
	return buffer.Bytes()
}

func isError(err error) bool {
	center, ok := err.(*errors.Center)
	return err != nil && (ok && !center.IsEmpty())
}

func respond(rs *wildcat.ResultSet, err error, res http.ResponseWriter, sizer wildcat.Sizer) {
	updateHeader(res)
	if isError(err) {
		respondImpl(res, 400, []byte(fmt.Sprintf(`{"message":"%s"}`, err.Error())))
	} else {
		respondImpl(res, 200, createResultJSON(rs, sizer))
	}
}

func respondImpl(res http.ResponseWriter, statusCode int, message []byte) {
	res.WriteHeader(statusCode)
	res.Write(message)
}

func readAsTargetList(targets *wildcat.Targets, entry wildcat.Entry, opts *wildcat.ReadOptions) *wildcat.Targets {
	newOpts := *opts
	newOpts.FileList = false
	reader, err := entry.Open()
	config := wildcat.NewConfig(wildcat.NewNoIgnore(), &newOpts, errors.New())
	if err == nil {
		targets.ReadFileListFromReader(reader, wildcat.NewOrder(), config)
	}
	return targets
}

func createTargets(req *http.Request, name string, opts *wildcat.ReadOptions) *wildcat.Targets {
	targets := &wildcat.Targets{}
	var entry wildcat.Entry = &myEntry{name: name, reader: wildcat.NewReadSeekCloser(req.Body)}
	appendTargetItem(targets, entry, opts)
	return targets
}

func countsBody(res http.ResponseWriter, req *http.Request, opts *wildcat.ReadOptions) (*wildcat.ResultSet, error) {
	fileName := req.URL.Query().Get("file-name")
	if fileName == "" {
		fileName = "<request>"
	}
	targets := createTargets(req, fileName, opts)
	return targets.CountAll(wildcat.DefaultGenerator)
}

func counts(res http.ResponseWriter, req *http.Request) {
	logger.Infof("counts: %s\n", req.URL)
	contentType := req.Header.Get("Content-Type")
	defer req.Body.Close()
	handlers := []struct {
		contentType string
		execFunc    func(http.ResponseWriter, *http.Request, *wildcat.ReadOptions) (*wildcat.ResultSet, error)
	}{
		{"multipart/form-data", countsMultipartBody},
		{"*", countsBody},
	}
	opts := parseQueryParams(req)
	sizer := wildcat.BuildSizer(false)
	for _, handler := range handlers {
		if handler.contentType == "*" || strings.HasPrefix(contentType, handler.contentType) {
			rs, err := handler.execFunc(res, req, opts)
			respond(rs, err, res, sizer)
			break
		}
	}
}

func appendTargetItem(targets *wildcat.Targets, entry wildcat.Entry, opts *wildcat.ReadOptions) {
	if opts.FileList {
		reader, err := entry.Open()
		if err == nil {
			defer reader.Close()
			readAsTargetList(targets, entry, opts)
		}
	} else {
		if !opts.NoExtract {
			convertedEntry, _ := wildcat.ConvertToArchiveEntry(entry)
			entry = convertedEntry
		}
		targets.Push(entry)
	}
}

func countsMultipartBody(res http.ResponseWriter, req *http.Request, opts *wildcat.ReadOptions) (*wildcat.ResultSet, error) {
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		return nil, fmt.Errorf("ParseMultpartForm: %w", err)
	}
	targets := &wildcat.Targets{}
	index := wildcat.NewOrder()
	for _, headers := range req.MultipartForm.File {
		for _, header := range headers {
			appendTargetItem(targets, &multipartEntry{header: header, index: index}, opts)
			index = index.Next()
		}
	}
	return targets.CountAll(wildcat.DefaultGenerator)
}

func wrapHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("url: %s\n", r.URL)
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

func registerHandlers(router *mux.Router) {
	router.HandleFunc("/counts", counts).Methods("POST")
	router.HandleFunc("/counts", optionsHandler).Methods("OPTIONS")
}

func createRestAPIServer() *mux.Router {
	router := mux.NewRouter()
	registerHandlers(router.PathPrefix("/wildcat/api/").Subrouter())
	router.PathPrefix("/wildcat").Handler(fileServer())
	return router
}

func fileServer() http.Handler {
	dirs := []string{
		"docs/public",
		"/opt/wildcat/docs",
		"/usr/local/opts/wildcat/docs",
	}
	for _, dir := range dirs {
		if wildcat.ExistDir(dir) {
			logger.Infof("ready to serve on %s", dir)
			return wrapHandler(http.StripPrefix("/wildcat/", http.FileServer(http.Dir(dir))))
		}
	}
	return nil
}
func (server *serverOptions) start(router *mux.Router) int {
	logger.Infof("Listen server at port %d", server.port)
	logger.Infof("start shutdown: %s\n", http.ListenAndServe(fmt.Sprintf(":%d", server.port), router))
	return 0
}

func (server *serverOptions) launchServer() int {
	logger.SetLevel(logger.INFO)
	router := createRestAPIServer()
	return server.start(router)
}
