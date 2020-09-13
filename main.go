package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/indiejustice/redirection-tracking/pkg/client_cookie"
)

type Configuration struct {
	Port       string `json:"port"`
	CookieName string `json:"cookie_name"`
	Debug      bool   `json:"debug"`
}

type PageData struct {
	AcceptWebp bool
}

var config Configuration

var logError *log.Logger
var logInfo *log.Logger
var logDebug *log.Logger
var clientCookie *client_cookie.ClientCookie
var indexTemplate *template.Template

func main() {
	file, err := ioutil.ReadFile("./config/config.json")
	panicError(err)

	json.Unmarshal(file, &config)

	logError = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	logInfo = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	if config.Debug {
		logDebug = log.New(os.Stdout, "DEBUG\t", log.Ldate|log.Ltime)
	} else {
		logDebug = log.New(ioutil.Discard, "", 0)
	}

	clientCookie = &client_cookie.ClientCookie{Name: config.CookieName}

	paths := []string{"landing_page/tmpl/index.html"}
	indexTemplate = template.Must(template.ParseFiles(paths...))

	mux := http.NewServeMux()
	mux.HandleFunc("/", pageHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("landing_page/static/"))))

	logInfo.Println("Landing Page Server running on port:", config.Port)
	panic(http.ListenAndServe(":"+config.Port, mux))
}
func pageHandler(w http.ResponseWriter, r *http.Request) {
	_, w = clientCookie.GetClientID(w, r)

	acceptHeader := r.Header.Get("Accept")

	var pageData PageData

	if strings.Contains(acceptHeader, "image/webp") {
		pageData.AcceptWebp = true
	}

	if r.URL.Path == "/" {
		indexTemplate.Execute(w, pageData)
	} else {
		returnCode404(w, r)

	}
}
func returnCode404(w http.ResponseWriter, r *http.Request) {
	// see http://golang.org/pkg/net/http/#pkg-constants
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Page Not Found"))
}
func logAndExit(message string) {
	logError.Println(message)
	os.Exit(1)
}
func panicError(err error) {
	if err != nil {
		logAndExit(err.Error())
	}
}
