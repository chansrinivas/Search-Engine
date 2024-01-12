package main

import (
	"database/sql"
	"log"
)

// Calculate tfidf scores or bigram hits
func BigramSearches(db *sql.DB, firstTerm string, secondTerm string, url string) ([]SearchResult, map[string]float64) {
	var hitsCount float64
	var totalDocs float64
	var rows *sql.Rows
	var results []SearchResult
	resultMap := make(map[string]float64)

	// Count of number of hits
	queryCount := `SELECT COUNT(*)
    FROM BigramHits BH
    JOIN Words W1 ON BH.Term1_ID = W1.Term_ID
    JOIN Words W2 ON BH.Term2_ID = W2.Term_ID
    JOIN Urls U ON BH.url_ID = U.url_ID
    WHERE W1.word = ? AND W2.word = ?;`

	err := db.QueryRow(queryCount, firstTerm, secondTerm).Scan(&hitsCount)
	if err != nil {
		log.Fatal(err)
	}

	// Count of number of urls
	queryTotalCount := "SELECT COUNT(*) FROM Urls"
	err = db.QueryRow(queryTotalCount).Scan(&totalDocs)
	if err != nil {
		log.Fatal(err)
	}

	// Hits
	queryStmt := `
    SELECT W1.word AS word1, W2.word AS word2, U.url_name, U.word_count, U.title, BH.Term_Count
    FROM BigramHits BH
    JOIN Words W1 ON BH.Term1_ID = W1.Term_ID
    JOIN Words W2 ON BH.Term2_ID = W2.Term_ID
    JOIN Urls U ON BH.url_ID = U.url_ID
    WHERE W1.word = ? AND W2.word = ?;
	`
	rows, err = db.Query(queryStmt, firstTerm, secondTerm)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var word1, word2, urlName, title string
		var wordCount, termCount float64

		err := rows.Scan(&word1, &word2, &urlName, &wordCount, &title, &termCount)
		if err != nil {
			log.Fatal(err)
		}

		// Calculate term frequency
		tfidf := calcTFIDF(hitsCount, totalDocs, termCount, wordCount)
		resultMap[title] = tfidf
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, SearchResult{URL: urlName, Title: title, Freq: int(termCount), TFIDF: tfidf})
	}
	return results, resultMap
}

func calcTFIDF(hitsCount float64, totalDocs float64, termCount float64, wordCount float64) float64 {
	df := hitsCount / totalDocs
	idf := 1 / df
	tf := (termCount) / (wordCount)
	tfidf := tf * idf
	return tfidf
}
