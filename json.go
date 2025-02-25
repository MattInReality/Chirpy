package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithError(w http.ResponseWriter, code int, message string, err error) {
	if err != nil {
		log.Println(err)
	}

	if code > 499 {
		log.Printf("responding with 5XX error: %s", message)
	}

	type responseError struct {
		Error string `json:"error"`
	}

	respondWithJson(w, code, responseError{Error: message})
}

func respondWithJson(w http.ResponseWriter, code int, body interface{}) {
	data, err := json.Marshal(body)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}
