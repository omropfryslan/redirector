package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func redirector(response http.ResponseWriter, request *http.Request, db Database) {
	destination, append, code, err := db.Get(request.Host)

	if err != nil {
		http.Error(response, `{"error": "No such URL"}`, http.StatusNotFound)
		return
	}

	realDestination := destination
	if append {
		realDestination += request.URL.Path
	}

	http.Redirect(response, request, realDestination, code)
}

func main() {
	var (
		frontProxy = false
		port       = "1337"
	)

	if os.Getenv("BASE_URL") == "" {
		log.Fatal("BASE_URL environment variable must be set")
	}
	if os.Getenv("DB_PATH") == "" {
		log.Fatal("DB_PATH environment variable must be set")
	}

	if os.Getenv("FRONT_PROXY") != "" {
		_frontproxy := os.Getenv("FRONT_PROXY")
		frontProxy, _ = strconv.ParseBool(_frontproxy)
	}

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	port = ":" + port

	db := sqlite{Path: path.Join(os.Getenv("DB_PATH"), "db.sqlite")}
	db.Init()

	baseURL := os.Getenv("BASE_URL")
	r := mux.NewRouter()

	if frontProxy {
		r.Use(handlers.ProxyHeaders)
	}

	r.Host(baseURL).Path("/api/load").HandlerFunc(
		func(response http.ResponseWriter, request *http.Request) {
			loaddomains(response, request, db)
		})

	r.Host(baseURL).Path("/api/save").HandlerFunc(
		func(response http.ResponseWriter, request *http.Request) {
			savedomain(response, request, db)
		}).Methods("POST")

	r.Host(baseURL).Path("/api/delete").HandlerFunc(
		func(response http.ResponseWriter, request *http.Request) {
			deletedomain(response, request, db)
		}).Methods("POST")

	r.Host(baseURL).Handler(http.FileServer(http.Dir("public")))

	r.PathPrefix("/").HandlerFunc(
		func(response http.ResponseWriter, request *http.Request) {
			redirector(response, request, db)
		})

	log.Println("Starting server on port " + port)
	log.Fatal(http.ListenAndServe(port, handlers.LoggingHandler(os.Stdout, r)))
}
