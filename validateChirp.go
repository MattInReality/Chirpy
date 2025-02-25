package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func getBadWords() map[string]int {
	return map[string]int{
		"kerfuffle": 0,
		"sharbert":  0,
		"fornax":    0,
	}
}

type Chirp struct {
	Text     string `json:"body"`
	BadWords bool
}

func sanitisedChirp(chirpText string, badWords map[string]int) string {
	splitText := strings.Split(chirpText, " ")
	for i, w := range splitText {
		l := strings.ToLower(w)
		if _, ok := badWords[l]; ok {
			splitText[i] = "****"
		}
	}
	return strings.Join(splitText, " ")
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Chirp string `json:"body"`
	}
	const maxChirpLength = 140
	defer r.Body.Close()
	d := json.NewDecoder(r.Body)
	p := params{}
	w.Header().Add("Content-Type", "application/json")
	err := d.Decode(&p)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	if len(p.Chirp) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	badWords := getBadWords()
	sanitisedChirp := sanitisedChirp(p.Chirp, badWords)
	type badWordResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}
	respondWithJson(w, http.StatusOK, badWordResponse{sanitisedChirp})
}
