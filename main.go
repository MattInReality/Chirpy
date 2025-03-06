package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/MattInReality/Chirpy/internal/auth"
	"github.com/MattInReality/Chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os"
	"sort"
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
		secret:   os.Getenv("JWT_SECRET"),
		apiKey:   os.Getenv("POLKA_KEY"),
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.getMetrics)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetOneChirp)
	mux.HandleFunc("POST /api/login", apiCfg.handlerUserLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevokeRefresh)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerPolkaWebhook)

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
	secret         string
	apiKey         string
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
		Email    string `json:"email"`
		Password string `json:"password"`
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
	hashed, err := auth.HashPassword(data.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong", err)
		return
	}
	user := database.CreateUserParams{
		Email:          data.Email,
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		HashedPassword: hashed,
	}
	newUser, err := cfg.db.CreateUser(r.Context(), user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error saving to db", err)
		return
	}
	type User struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}
	resUser := User{
		ID:          newUser.ID,
		CreatedAt:   newUser.CreatedAt,
		UpdatedAt:   newUser.UpdatedAt,
		Email:       newUser.Email,
		IsChirpyRed: newUser.IsChirpyRed,
	}
	respondWithJson(w, http.StatusCreated, resUser)
}

type chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		return
	}

	type params struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	const maxChirpLength = 140
	defer r.Body.Close()
	d := json.NewDecoder(r.Body)
	p := params{}
	w.Header().Add("Content-Type", "application/json")
	err = d.Decode(&p)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	p.UserID = userID
	if len(p.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	badWords := getBadWords()
	chirpParam := database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      sanitisedChirp(p.Body, badWords),
		UserID:    p.UserID,
	}
	log.Printf("%v", chirpParam)

	newChirp, err := cfg.db.CreateChirp(
		r.Context(),
		chirpParam,
	)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "issue inserting in to database", err)
		return
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

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("author_id")
	sortBy := r.URL.Query().Get("sort")
	var chirps []database.Chirp
	var err error
	if userID == "" {
		chirps, err = cfg.db.GetChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "error getting data from database", err)
			return
		}
	} else {
		uID, err := uuid.Parse(userID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "error getting data from database", err)
			return
		}
		chirps, err = cfg.db.GetChirpsByUserID(r.Context(), uID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "error getting data from database", err)
			return
		}
	}
	if sortBy == "asc" {
		sort.Slice(chirps, func(i, j int) bool {
			return (chirps[i].CreatedAt.Compare(chirps[j].CreatedAt) < 0)
		})
	}
	if sortBy == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return (chirps[i].CreatedAt.Compare(chirps[j].CreatedAt) >= 0)
		})
	}
	theChirps := []chirp{}
	for _, c := range chirps {
		chrp := chirp{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID,
		}
		theChirps = append(theChirps, chrp)
	}
	respondWithJson(w, http.StatusOK, theChirps)
}

func (cfg *apiConfig) handlerGetOneChirp(w http.ResponseWriter, r *http.Request) {
	var chirpID uuid.UUID
	chirpID = uuid.MustParse(r.PathValue("chirpID"))
	c, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound), err)
		return
	}
	chrp := chirp{
		ID:        c.ID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Body:      c.Body,
		UserID:    c.UserID,
	}
	respondWithJson(w, http.StatusOK, chrp)
}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	data := params{}
	d := json.NewDecoder(r.Body)
	d.Decode(&data)
	storedUser, err := cfg.db.GetUserByEmail(r.Context(), data.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "please try again", err)
		return
	}
	if err := auth.CheckPasswordHash(data.Password, storedUser.HashedPassword); err != nil {
		respondWithError(w, http.StatusUnauthorized, "please try again", err)
		return
	}
	token, err := auth.MakeJWT(storedUser.ID, cfg.secret, calculateTimeout(60*60))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "please try again", err)
		return
	}
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "please try again", err)
		return
	}
	now := time.Now()
	_, err = cfg.db.CreateRefreshToken(
		r.Context(),
		database.CreateRefreshTokenParams{
			Token:     refreshToken,
			CreatedAt: now,
			UpdatedAt: now,
			UserID:    storedUser.ID,
			ExpiresAt: now.AddDate(0, 0, 60),
			RevokedAt: sql.NullTime{Valid: false},
		},
	)
	type User struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}
	resUser := User{
		ID:           storedUser.ID,
		CreatedAt:    storedUser.CreatedAt,
		UpdatedAt:    storedUser.UpdatedAt,
		Email:        storedUser.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  storedUser.IsChirpyRed,
	}
	respondWithJson(w, http.StatusOK, resUser)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
	}
	rt, err := cfg.db.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		return
	}
	newToken, err := auth.MakeJWT(rt.ID, cfg.secret, calculateTimeout(60*60))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "please try again", err)
		return
	}
	type rParam struct {
		Token string `json:"token"`
	}
	respondWithJson(w, http.StatusOK, rParam{Token: newToken})
}

func (cfg *apiConfig) handlerRevokeRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		return
	}
	now := time.Now()
	err = cfg.db.RevokeRefreshToken(
		r.Context(),
		database.RevokeRefreshTokenParams{
			Token:     token,
			RevokedAt: sql.NullTime{Time: now, Valid: true},
			UpdatedAt: now,
		})
	w.WriteHeader(204)
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		return
	}
	type params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	p := &params{}
	d := json.NewDecoder(r.Body)
	d.Decode(p)
	user, err := cfg.db.GetUserByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		return
	}
	if userID != user.ID {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		return
	}
	hash, err := auth.HashPassword(p.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), err)
		return
	}
	updated, err := cfg.db.UpdateUser(
		r.Context(),
		database.UpdateUserParams{
			ID:             user.ID,
			Email:          p.Email,
			HashedPassword: hash,
			UpdatedAt:      time.Now(),
		})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), err)
		return
	}
	type response struct {
		ID          uuid.UUID `json:"id"`
		Email       string    `json:"email"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}
	respondWithJson(w, http.StatusOK, response{ID: updated.ID, Email: updated.Email, CreatedAt: updated.CreatedAt, UpdatedAt: updated.UpdatedAt, IsChirpyRed: updated.IsChirpyRed})
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		return
	}
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound), err)
		return
	}
	deleted, err := cfg.db.DeleteChirp(r.Context(), database.DeleteChirpParams{
		ID:     chirpID,
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusForbidden, http.StatusText(http.StatusForbidden), err)
		return
	}
	log.Printf("%v\n", deleted)
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	providedKey, err := auth.GetAPIKey(r.Header)
	if err != nil || providedKey != cfg.apiKey {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), err)
		return
	}
	type params struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	p := &params{}
	d := json.NewDecoder(r.Body)
	err = d.Decode(p)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "issue decoding body", err)
		return
	}
	if p.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	err = cfg.db.UpgradeToRedByID(r.Context(), p.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound), err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func calculateTimeout(seconds int) time.Duration {
	if seconds == 0 || seconds >= 60 {
		return time.Duration(int64(time.Second) * 60 * 60)
	}
	return time.Duration(int64(time.Second) * int64(seconds))
}
