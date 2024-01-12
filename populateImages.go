package main

import (
	"database/sql"
	"log"
	"strings"
)

// Populate image details into the terms, urls and hits table
func populateImagesIndex(stopWordsMap map[string]struct{}, db *sql.DB, currentURL string, cleanRes ExtractResult) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	for _, word := range cleanRes.altWords {
		stemmedWord := stemWords(word)
		if _, isStop := stopWordsMap[stemmedWord]; !isStop {
			// Insert into the words table
			_, err := tx.Exec("INSERT OR IGNORE INTO ImageWords (alt_word) VALUES (?)", stemmedWord)
			if err != nil {
				tx.Rollback()
				log.Fatal(err)
			}
			// Store the current term's ID for the Hits table
			rows, err := tx.Query("SELECT AltWord_ID FROM ImageWords WHERE alt_word = ?", stemmedWord)
			if err != nil {
				tx.Rollback()
				log.Fatal(err)
			}

			var Term_ID int
			for rows.Next() {
				err := rows.Scan(&Term_ID)
				if err != nil {
					tx.Rollback()
					log.Fatal(err)
				}
			}

			// Insert into the urls table
			_, err = tx.Exec("INSERT OR IGNORE INTO ImageUrls (url_name, title, altword_count) VALUES (?, ?, 0)", cleanRes.url, cleanRes.title)
			if err != nil {
				tx.Rollback()
				log.Fatal(err)
			}
			var url_ID int
			// Store the current url's to be inserted into the Hits table
			err = tx.QueryRow("SELECT id FROM ImageUrls WHERE url_name = ?", cleanRes.url).Scan(&url_ID)
			if err != nil {
				tx.Rollback()
				log.Fatal(err)
			}

			// Update the URLs table to increment the total number of words in the document
			_, err = tx.Exec("UPDATE ImageUrls SET altword_count = altword_count + 1 WHERE url_name = ?", cleanRes.url)

			for j, imgAlt := range cleanRes.Images {
				if strings.Contains(strings.ToLower(imgAlt), strings.ToLower(stemmedWord)) {

					_, err = tx.Exec("INSERT OR IGNORE INTO ImageHits (alt_word_id, url_ID, imgSrc, term_count) VALUES (?, ?, ?, 0)",
						Term_ID, url_ID, cleanRes.ImageSrcs[j])
					if err != nil {
						tx.Rollback()
						log.Fatal(err)
					}

					// Update the ImageHits table to increment the total number of times a particular word appeared in the document
					_, err = tx.Exec("UPDATE ImageHits SET term_count = term_count + 1 WHERE alt_word_id = ? AND url_ID = ? AND imgSrc = ?",
						Term_ID, url_ID, cleanRes.ImageSrcs[j])
					if err != nil {
						tx.Rollback()
						log.Fatal(err)
					}
				}
			}
		}

	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

}
