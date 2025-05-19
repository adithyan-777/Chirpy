package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

// type handler struct{}

func healthzHandlder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8") // normal header
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
func (cfg *apiConfig) merticsReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}

func validationHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode((&params))
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(params.Body) > 140 {
		w.Header().Set("Content-Type", "application/json") // normal header
		type lengthErrorMsg struct {
			Error string `json:"error"`
		}
		errorMsg := lengthErrorMsg{Error: "Chirp is too long"}
		jsonBytes, err := json.Marshal(errorMsg)
		if err != nil {
			log.Printf("Error marshalling error response: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonBytes)
	} else {
		type validResponse struct {
			Valid        bool   `json:"valid"`
			Cleaned_body string `json:"cleaned_body"`
		}
		words := strings.Split(params.Body, " ")
		badwords := []string{"kerfuffle", "sharbert", "fornax"}
		badwordsMap := make(map[string]bool)

		var result []string
		for _, word := range badwords {
			badwordsMap[word] = true
		}
		for _, word := range words {

			if badwordsMap[strings.ToLower(word)] {
				result = append(result, "****")
			} else {
				result = append(result, word)
			}
		}
		final := strings.Join(result, " ")
		successMsg := validResponse{Valid: true, Cleaned_body: final}
		jsonBytes, err := json.Marshal(successMsg)
		if err != nil {
			log.Printf("Error marshalling success response: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json") // normal header
		w.WriteHeader(http.StatusOK)
		w.Write(jsonBytes)
	}
}

func responseWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json") // normal header
	w.WriteHeader(code)
	w.Write([]byte("OK"))
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json") // normal header
	w.WriteHeader(code)
	w.Write([]byte("OK"))
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	apiCfg := &apiConfig{}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", healthzHandlder)
	mux.HandleFunc("POST /api/validate_chirp", validationHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.merticsReset)

	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServer))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
