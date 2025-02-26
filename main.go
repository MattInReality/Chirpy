package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/MattInReality/Chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os"
	"sync/atomic"
	"time"
)

import _ "github.com/lib/pq"

func main() {
	const filepathRoot = "."
	const port = "8080"
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("could not connect to db")
	}
	queries := database.New(db)

	apiCfg := &apiConfig{
		db:       queries,
		platform: os.Getenv("PLATFORM"),
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.getMetrics)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("/admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("/api/chirps", apiCfg.handlerCreateChirp)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Servinbg files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())

}

func handlerReadiness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("<html>\n  <body>\n    <h1>Welcome, Chirpy Admin</h1>\n    <p>Chirpy has been visited %d times!</p>\n  </body>\n</html>", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits.Swap(0)
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, http.StatusText(http.StatusForbidden), nil)
		return
	}
	err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "issue deleting resource", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Email string `json:"email"`
	}
	data := params{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error reading data", err)
		return
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error unmarshalling data", err)
		return
	}
	if _, err := mail.ParseAddress(data.Email); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid email", err)
		return
	}
	user := database.CreateUserParams{
		Email:     data.Email,
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	newUser, err := cfg.db.CreateUser(r.Context(), user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error saving to db", err)
		return
	}
	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}
	resUser := User{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	}
	respondWithJson(w, http.StatusCreated, resUser)
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
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
	if len(p.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	badWords := getBadWords()
	sanitisedChirp := sanitisedChirp(p.Body, badWords)

	newChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Body: sanitisedChirp, UserID: p.UserID})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "issue inserting in to database", err)
	}
	type chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}
	respondWithJson(w, http.StatusCreated, chirp{ID: newChirp.ID, CreatedAt: newChirp.CreatedAt, UpdatedAt: newChirp.UpdatedAt, Body: newChirp.Body, UserID: newChirp.UserID})
}
