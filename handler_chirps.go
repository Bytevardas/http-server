package main

import (
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
)

var badWords = []string{"kerfuffle", "sharbert", "fornax"}

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
	}

	defer req.Body.Close()
	data, err := io.ReadAll(req.Body)
	if err != nil {
		respondWithError(w, 500, "could not read the body")
		return
	}

	params := requestBody{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		respondWithError(w, 500, "could not unmarshal the body")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "chirp is longer the 140 chars")
		return
	}

	filteredSentence := filterBadWords(params.Body)
	type responseBody struct {
		Valid       bool   `json:"valid"`
		CleanedBody string `json:"cleaned_body,omitempty"`
	}
	respondWithJSON(w, 200, responseBody{Valid: true, CleanedBody: filteredSentence})
}

func filterBadWords(sentence string) string {
	words := strings.Split(sentence, " ")
	filtered := make([]string, 0, len(words))
	for _, word := range words {
		if slices.Contains(badWords, strings.ToLower(word)) {
			filtered = append(filtered, "****")
			continue
		}
		filtered = append(filtered, word)
	}
	return strings.Join(filtered, " ")
}
