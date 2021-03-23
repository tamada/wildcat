package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestParseQueryParam(t *testing.T) {
	testdata := []struct {
		giveURL           string
		wontFileListFlag  bool
		wontNoExtractFlag bool
	}{
		{"/wildcat/api/counts", false, false},
		{"/wildcat/api/counts?readAs=", false, false},
		{"/wildcat/api/counts?readAs=file-list", true, false},
		{"/wildcat/api/counts?readAs=no-extract", false, true},
		{"/wildcat/api/counts?readAs=file-list,no-extract", true, true},
		{"/wildcat/api/counts?readAs=no-extract,file-list", true, true},
		{"/wildcat/api/counts?readAs=no-extract&readAs=file-list", true, true},
		{"/wildcat/api/counts?readAs=no-extract,file-list&file-name=hoge", true, true},
		{"/wildcat/api/counts?readAs=file-list,no-extract&file-name=hoge", true, true},
	}
	for _, td := range testdata {
		req := httptest.NewRequest("POST", td.giveURL, nil)
		opts := parseQueryParams(req)
		if opts.FileList != td.wontFileListFlag {
			t.Errorf("%s: parseOptions failed, fileList: wont %v, got %v", td.giveURL, td.wontFileListFlag, opts.FileList)
		}
		if opts.NoExtract != td.wontNoExtractFlag {
			t.Errorf("%s: parseOptions failed, noExtract: wont %v, got %v", td.giveURL, td.wontNoExtractFlag, opts.NoExtract)
		}
	}
}

func TestBasicRequest(t *testing.T) {
	testdata := []struct {
		giveURL         string
		giveContentPath string
		wontStatus      int
		wontSuffix      string
	}{
		{"/wildcat/api/counts", "../../testdata/wc/humpty_dumpty.txt", 200, `"results":[{"filename":"<request>","lines":"4","words":"26","characters":"142","bytes":"142"}]}`},
		{"/wildcat/api/counts?file-name=humpty_dumpty.txt", "../../testdata/wc/humpty_dumpty.txt", 200, `"results":[{"filename":"humpty_dumpty.txt","lines":"4","words":"26","characters":"142","bytes":"142"}]}`},
		{"/wildcat/api/counts", "../../testdata/archives/wc.jar", 200, `"results":[{"filename":"<request>!humpty_dumpty.txt","lines":"4","words":"26","characters":"142","bytes":"142"},{"filename":"<request>!ja/","lines":"0","words":"0","characters":"0","bytes":"0"},{"filename":"<request>!ja/sakura_sakura.txt","lines":"15","words":"26","characters":"118","bytes":"298"},{"filename":"<request>!london_bridge_is_broken_down.txt","lines":"59","words":"260","characters":"1,341","bytes":"1,341"},{"filename":"total","lines":"78","words":"312","characters":"1,601","bytes":"1,781"}]`},
		{"/wildcat/api/counts?file-name=wc.jar", "../../testdata/archives/wc.jar", 200, `"results":[{"filename":"wc.jar!humpty_dumpty.txt","lines":"4","words":"26","characters":"142","bytes":"142"},{"filename":"wc.jar!ja/","lines":"0","words":"0","characters":"0","bytes":"0"},{"filename":"wc.jar!ja/sakura_sakura.txt","lines":"15","words":"26","characters":"118","bytes":"298"},{"filename":"wc.jar!london_bridge_is_broken_down.txt","lines":"59","words":"260","characters":"1,341","bytes":"1,341"},{"filename":"total","lines":"78","words":"312","characters":"1,601","bytes":"1,781"}]`},
		{"/wildcat/api/counts?file-name=wc.jar&readAs=no-extract", "../../testdata/archives/wc.jar", 200, `"results":[{"filename":"wc.jar","lines":"5","words":"62","characters":"1,054","bytes":"1,080"}]`},
	}

	router := createRestAPIServer()
	for _, td := range testdata {
		reader, _ := os.Open(td.giveContentPath)
		defer reader.Close()
		req := httptest.NewRequest("POST", td.giveURL, reader)
		rec := httptest.NewRecorder()
		req.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
		router.ServeHTTP(rec, req)
		if rec.Code != td.wontStatus {
			t.Errorf("%s: status code did not match, wont %d, got %d", td.giveURL, td.wontStatus, rec.Code)
		}
		gotData := rec.Body.String()
		if !strings.Contains(gotData, td.wontSuffix) {
			t.Errorf("%s: response body did not match, wont %s, got %s", td.giveURL, td.wontSuffix, gotData)
		}
	}
}

func TestFileList(t *testing.T) {
	testdata := []struct {
		giveURL    string
		wontStatus int
		wontSuffix string
	}{
		// {"/wildcat/api/counts", 200, `"results":[{"filename":"<request>","lines":"1","words":"2","characters":"140","bytes":"140"}]`},
		{"/wildcat/api/counts?readAs=file-list", 200, `"results":[{"filename":"https://github.com/tamada/wildcat/raw/main/testdata/archives/wc.jar!humpty_dumpty.txt","lines":"4","words":"26","characters":"142","bytes":"142"},{"filename":"https://github.com/tamada/wildcat/raw/main/testdata/archives/wc.jar!ja/","lines":"0","words":"0","characters":"0","bytes":"0"},{"filename":"https://github.com/tamada/wildcat/raw/main/testdata/archives/wc.jar!ja/sakura_sakura.txt","lines":"15","words":"26","characters":"118","bytes":"298"},{"filename":"https://github.com/tamada/wildcat/raw/main/testdata/archives/wc.jar!london_bridge_is_broken_down.txt","lines":"59","words":"260","characters":"1,341","bytes":"1,341"},{"filename":"https://github.com/tamada/wildcat/raw/main/testdata/wc/humpty_dumpty.txt","lines":"4","words":"26","characters":"142","bytes":"142"},{"filename":"total","lines":"82","words":"338","characters":"1,743","bytes":"1,923"}]`},
		// {"/wildcat/api/counts?readAs=no-extract,file-list", 200, `"results":[{"filename":"https://github.com/tamada/wildcat/raw/main/testdata/archives/wc.jar","lines":"5","words":"62","characters":"1,054","bytes":"1,080"},{"filename":"https://github.com/tamada/wildcat/raw/main/testdata/wc/humpty_dumpty.txt","lines":"4","words":"26","characters":"142","bytes":"142"},{"filename":"total","lines":"9","words":"88","characters":"1,196","bytes":"1,222"}]`},
	}
	content := `https://github.com/tamada/wildcat/raw/main/testdata/archives/wc.jar
https://github.com/tamada/wildcat/raw/main/testdata/wc/humpty_dumpty.txt`
	router := createRestAPIServer()
	for _, td := range testdata {
		req := httptest.NewRequest("POST", td.giveURL, strings.NewReader(content))
		rec := httptest.NewRecorder()
		req.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
		router.ServeHTTP(rec, req)
		if rec.Code != td.wontStatus {
			t.Errorf("%s: status code did not match, wont %d, got %d", td.giveURL, td.wontStatus, rec.Code)
		}
		gotData := rec.Body.String()
		if !strings.Contains(gotData, td.wontSuffix) {
			t.Errorf("%s: response body did not match,\nwont %s,\ngot  %s", td.giveURL, td.wontSuffix, gotData)
		}
	}
}

func addPart(writer *multipart.Writer, name string, fileName string) {
	dest, _ := writer.CreateFormFile("file", name)
	file, _ := os.Open(fileName)
	defer file.Close()
	io.Copy(dest, file)
}

func TestMultipart(t *testing.T) {
	testdata := []struct {
		giveURL    string
		wontStatus int
		wontSuffix string
	}{
		{"/wildcat/api/counts", 200, `"results":[{"filename":"humpty_dumpty.txt","lines":"4","words":"26","characters":"142","bytes":"142"},{"filename":"wc.jar!humpty_dumpty.txt","lines":"4","words":"26","characters":"142","bytes":"142"},{"filename":"wc.jar!ja/","lines":"0","words":"0","characters":"0","bytes":"0"},{"filename":"wc.jar!ja/sakura_sakura.txt","lines":"15","words":"26","characters":"118","bytes":"298"},{"filename":"wc.jar!london_bridge_is_broken_down.txt","lines":"59","words":"260","characters":"1,341","bytes":"1,341"},{"filename":"total","lines":"82","words":"338","characters":"1,743","bytes":"1,923"}]`},
		{"/wildcat/api/counts?readAs=no-extract", 200, `"results":[{"filename":"humpty_dumpty.txt","lines":"4","words":"26","characters":"142","bytes":"142"},{"filename":"wc.jar","lines":"5","words":"62","characters":"1,054","bytes":"1,080"},{"filename":"total","lines":"9","words":"88","characters":"1,196","bytes":"1,222"}]`},
	}
	router := createRestAPIServer()
	content := bytes.NewBuffer([]byte{})
	writer := multipart.NewWriter(content)
	addPart(writer, "humpty_dumpty.txt", "../../testdata/wc/humpty_dumpty.txt")
	addPart(writer, "wc.jar", "../../testdata/archives/wc.jar")
	writer.Close()
	for _, td := range testdata {
		req := httptest.NewRequest("POST", td.giveURL, bytes.NewReader(content.Bytes()))
		rec := httptest.NewRecorder()
		req.Header["Content-Type"] = []string{writer.FormDataContentType()}
		router.ServeHTTP(rec, req)
		if rec.Code != td.wontStatus {
			t.Errorf("%s: status code did not match, wont %d, got %d", td.giveURL, td.wontStatus, rec.Code)
		}
		gotData := rec.Body.String()
		if !strings.Contains(gotData, td.wontSuffix) {
			t.Errorf("%s: response body did not match,\nwont %s,\ngot  %s", td.giveURL, td.wontSuffix, gotData)
		}
	}
}
