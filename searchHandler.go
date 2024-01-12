package main

import (
	"database/sql"
	"log"
	"net/http"
	"sort"
	"strings"
	"text/template"
)

type SearchResult struct {
	URL    string
	Freq   int
	Title  string
	TFIDF  float64
	ImgSrc string
}

type SearchResultSlice []SearchResult

func (s SearchResultSlice) Len() int {
	return len(s)
}

// Descending order sort
func (s SearchResultSlice) Less(i, j int) bool {

	if s[i].TFIDF == s[j].TFIDF {
		return s[i].URL > s[j].URL
	}
	return s[i].TFIDF > s[j].TFIDF
}

func (s SearchResultSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func SQLSearchEngineHandler(db *sql.DB, url string, resultTemplate *template.Template) map[string]float64 {
	var resultMapSearch map[string]float64
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		// Get the search term from the URL query parameter
		searchTerm := r.URL.Query().Get("term")
		if searchTerm == "" {
			http.Error(w, "Please enter a search term", http.StatusBadRequest)
			return
		}

		wildcard := r.URL.Query().Get("wildcard") == "wildcard"
		image := r.URL.Query().Get("image") == "image"
		searchTerm = strings.ToLower(searchTerm)
		stem := stemWords(searchTerm)

		var results []SearchResult
		var resultMap map[string]float64
		if strings.Contains(searchTerm, " ") {
			// The search term is a bigram, split into two words
			terms := strings.Split(searchTerm, " ")
			if len(terms) == 2 {
				firstTerm := stemWords(terms[0])
				secondTerm := stemWords(terms[1])

				// Call the BigramSearches function
				results, resultMap = BigramSearches(db, firstTerm, secondTerm, url)
				resultMapSearch = resultMap
			} else {
				// Handle invalid bigram format
				http.Error(w, "Invalid bigram format", http.StatusBadRequest)
				return
			}
		} else {
			if image {
				// Image search
				results, resultMap = imageSearch(db, stem, url)
				resultMapSearch = resultMap
			} else {
				// Unigram search
				results, resultMap = unigramSearch(db, stem, url, wildcard)
				resultMapSearch = resultMap
			}
		}

		resultMapSearch = resultMap
		sort.Sort(SearchResultSlice(results))

		// Data structure for the template
		data := struct {
			Results []SearchResult
		}{
			Results: results,
		}

		w.Header().Set("Content-Type", "text/html")

		// Execute the results template
		err := resultTemplate.Execute(w, data)
		if err != nil {
			log.Println("Error executing image template:", err)
		}

	})
	return resultMapSearch
}
