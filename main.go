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

		if len(request.URL.RawQuery) > 0 {
			realDestination += "?" + request.URL.RawQuery
		}
	}

	http.Redirect(response, request, realDestination, code)
}

// Set the URL.Host variable since handlers.ProxyHeaders does not.
// https://github.com/gorilla/handlers/pull/96
func setURLHost(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		xForwardedHost := http.CanonicalHeaderKey("X-Forwarded-Host")

		if forwardedHost := r.Header.Get(xForwardedHost); forwardedHost != "" {
			if r.URL.Host == "" {
				r.URL.Host = forwardedHost
			}
		}

		// Call the next handler in the chain.
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
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

		r.Use(setURLHost)
	}

	r.Path("/api/load").HandlerFunc(
		func(response http.ResponseWriter, request *http.Request) {
			loaddomains(response, request, db)
		}).
		Host(baseURL)

	r.Path("/api/save").HandlerFunc(
		func(response http.ResponseWriter, request *http.Request) {
			savedomain(response, request, db)
		}).
		Host(baseURL).
		Methods("POST")

	r.Path("/api/delete").HandlerFunc(
		func(response http.ResponseWriter, request *http.Request) {
			deletedomain(response, request, db)
		}).
		Host(baseURL).
		Methods("POST")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("public"))).
		Host(baseURL)

	r.PathPrefix("/").HandlerFunc(
		func(response http.ResponseWriter, request *http.Request) {
			redirector(response, request, db)
		})

	logHandler := handlers.CombinedLoggingHandler(os.Stdout, r)
	if frontProxy {
		logHandler = handlers.CombinedLoggingHandler(os.Stdout, setURLHost(handlers.ProxyHeaders(r)))
	}

	log.Println("Starting server on port " + port)
	log.Fatal(http.ListenAndServe(port, logHandler))
}
