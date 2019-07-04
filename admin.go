package main

import (
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/net/idna"
)

func savedomain(response http.ResponseWriter, request *http.Request, db Database) {
	decoder := json.NewDecoder(request.Body)
	domain := Domain{}

	err := decoder.Decode(&domain)
	if err != nil {
		log.Println(err)
		http.Error(response, `{"error": "Unable to parse json"}`, http.StatusBadRequest)
		return
	}

	domain.Domain, err = idna.ToASCII(domain.Domain)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = db.Save(domain)
	if err != nil {
		log.Println(err)
		return
	}
}

func deletedomain(response http.ResponseWriter, request *http.Request, db Database) {
	decoder := json.NewDecoder(request.Body)
	domain := Domain{}

	err := decoder.Decode(&domain)
	if err != nil {
		log.Println(err)
		http.Error(response, `{"error": "Unable to parse json"}`, http.StatusBadRequest)
		return
	}

	_, err = db.Delete(domain)
	if err != nil {
		log.Println(err)
		return
	}
}

func loaddomains(response http.ResponseWriter, request *http.Request, db Database) {
	destination, _ := db.GetAll()

	jsonData, _ := json.Marshal(destination)
	response.Write(jsonData)
}
