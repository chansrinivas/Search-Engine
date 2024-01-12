package main

import (
	"database/sql"
	"fmt"
	"log"
)

// Calculates TF-IDF and returns map of results
func unigramSearch(db *sql.DB, word string, url string, wildcard bool) ([]SearchResult, map[string]float64) {
	// Query to get the total number of matching documents
	var queryCount string
	var hitsCount float64
	var totalDocs float64
	var queryStmt string
	var rows *sql.Rows
	var results []SearchResult
	resultMap := make(map[string]float64)

	if wildcard {
		queryCount = `
		SELECT COUNT(*) AS record_count
		FROM Words
		LEFT JOIN Hits ON Words.Term_ID = Hits.Term_ID
		LEFT JOIN Urls ON Hits.url_ID = Urls.url_ID
		WHERE Words.word LIKE ?
		`
		err := db.QueryRow(queryCount, (word + "%")).Scan(&hitsCount)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		queryCount = `
		SELECT COUNT(*) AS record_count
		FROM Words
		LEFT JOIN Hits ON Words.Term_ID = Hits.Term_ID
		LEFT JOIN Urls ON Hits.url_ID = Urls.url_ID
		WHERE Words.word = ?
		`
		err := db.QueryRow(queryCount, (word)).Scan(&hitsCount)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Query to get the total number of urls as the total number of docs in the corpus
	queryTotalCount := "SELECT COUNT(*) FROM Urls"
	err := db.QueryRow(queryTotalCount).Scan(&totalDocs)
	if err != nil {
		log.Fatal(err)
	}

	// Wildcard Search
	if wildcard {
		queryStmt = `
            SELECT Words.word, Urls.url_name, Urls.word_count, Urls.title, Hits.Term_Count 
            FROM Words
            LEFT JOIN Hits ON Words.Term_ID = Hits.Term_ID
            LEFT JOIN Urls ON Hits.url_ID = Urls.url_ID
            WHERE Words.word LIKE ?`
		rows, err = db.Query(queryStmt, (word + "%"))
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
	} else {
		// Non wildcard search
		queryStmt = `
            SELECT Words.word, Urls.url_name, Urls.word_count, Urls.title, Hits.Term_Count 
            FROM Words
            LEFT JOIN Hits ON Words.Term_ID = Hits.Term_ID
            LEFT JOIN Urls ON Hits.url_ID = Urls.url_ID
            WHERE Words.word = ?`
		rows, err = db.Query(queryStmt, (word))
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
	}

	for rows.Next() {
		var u, title string
		var wc, tfreq float64
		err = rows.Scan(&word, &u, &wc, &title, &tfreq)
		// Calculate the tfidf scores
		tfidf := calcTFIDF(hitsCount, totalDocs, tfreq, wc)
		resultMap[title] = tfidf
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, SearchResult{URL: u, Title: title, Freq: int(tfreq), TFIDF: tfidf})
	}
	fmt.Println(resultMap)
	return results, resultMap
}
