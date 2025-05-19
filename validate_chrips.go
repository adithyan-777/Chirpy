package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func validationHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type validResponse struct {
		Cleaned_body string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChripLength = 140
	if len(params.Body) > maxChripLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	} else {
		badWordsMap := map[string]struct{}{
			"kerfuffle": {},
			"sharbert":  {},
			"fornax":    {},
		}
		cleaned := getCleanedBody(params.Body, badWordsMap)

		respondWithJSON(w, http.StatusOK, validResponse{
			Cleaned_body: cleaned,
		})
	}
}

func getCleanedBody(body string, badwords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		lowerredWord := strings.ToLower(word)
		if _, ok := badwords[lowerredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}
