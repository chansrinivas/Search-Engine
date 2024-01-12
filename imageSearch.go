package main

import (
	"database/sql"
	"fmt"
	"log"
)

func imageSearch(db *sql.DB, word string, url string) ([]SearchResult, map[string]float64) {
	var hitsCount float64
	var totalDocs float64
	var rows *sql.Rows
	var results []SearchResult
	resultMap := make(map[string]float64)

	// Query to calculate the total number of hits
	queryCount := `
		SELECT COUNT(*) AS record_count
		FROM ImageWords
		LEFT JOIN ImageHits ON ImageWords.AltWord_ID = ImageHits.alt_word_id
		LEFT JOIN ImageUrls ON ImageHits.url_ID = ImageUrls.id
		WHERE ImageWords.alt_word = ?
	`
	err := db.QueryRow(queryCount, (word)).Scan(&hitsCount)
	if err != nil {
		log.Fatal(err)
	}

	// Total number of documents
	queryTotalCount := "SELECT COUNT(*) FROM ImageUrls"
	err = db.QueryRow(queryTotalCount).Scan(&totalDocs)
	if err != nil {
		log.Fatal(err)
	}

	// Query for the hits
	queryStmt := `
		SELECT ImageWords.alt_word, ImageUrls.url_name, ImageUrls.altword_count, ImageUrls.title, ImageHits.imgSrc, ImageHits.term_count 
		FROM ImageWords
		LEFT JOIN ImageHits ON ImageWords.AltWord_ID = ImageHits.alt_word_id
		LEFT JOIN ImageUrls ON ImageHits.url_ID = ImageUrls.id
		WHERE ImageWords.alt_word = ?
	`
	rows, err = db.Query(queryStmt, (word))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var u, title, imgSrc string
		var wc, tfreq float64
		err = rows.Scan(&word, &u, &wc, &title, &imgSrc, &tfreq)
		// Calculate the tfidf scores for the image results and append them to the resultMap
		tfidf := calcTFIDF(hitsCount, totalDocs, tfreq, wc)
		resultMap[title] = tfidf
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, SearchResult{URL: u, Title: title, Freq: int(tfreq), TFIDF: tfidf, ImgSrc: imgSrc})
	}
	fmt.Println(results)
	return results, resultMap
}
