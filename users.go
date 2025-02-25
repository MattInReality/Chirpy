package main

import (
	"encoding/json"
	"github.com/MattInReality/Chirpy/internal/database"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"net/mail"
	"time"
)

func handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Email string `json:"email"`
	}
	data := params{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading body: %v", err)
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("error unmarshalling body: %v", err)
	}
	if _, err := mail.ParseAddress(data.Email); err != nil {
		log.Printf("%v", err)
	}
	_ = &database.CreateUserParams{
		Email:     data.Email,
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

}
