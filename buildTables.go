package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

func populateIndex(db *sql.DB, currentURL string, cleanRes ExtractResult, dlInC chan string) {

	stopWordsMap := buildStopWordsMap()
	populateImagesIndex(stopWordsMap, db, currentURL, cleanRes)
	// Begin a single transaction for all insertions
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		// recover() can stop a panic from aborting the program and let it continue with execution instead.
		if err := recover(); err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
	}()

	// Insert into the Urls table
	_, err = tx.Exec("INSERT OR IGNORE INTO Urls (url_name, word_count, title) VALUES (?, 0, ?)", cleanRes.url, strings.ReplaceAll(cleanRes.title, "â€“", "-"))
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	// Get the url_ID for the current URL
	var url_ID int
	err = tx.QueryRow("SELECT url_ID FROM Urls WHERE url_name = ?", cleanRes.url).Scan(&url_ID)
	if err != nil {
		tx.Rollback()
		fmt.Println("error in inserting to urls")
		log.Fatal(err)
	}

	// Map of terms to their term_id from the Words table
	termIDMap := make(map[string]int)

	// Iterate through cleaned words
	for i := 0; i < len(cleanRes.words); i++ {
		word := cleanRes.words[i]
		stemmedWord := stemWords(word)

		// Skip stop words
		if _, isStop := stopWordsMap[stemmedWord]; !isStop {
			// Insert into the Words table
			_, err := tx.Exec("INSERT OR IGNORE INTO Words (word) VALUES (?)", stemmedWord)
			if err != nil {
				tx.Rollback()
				log.Fatal(err)
			}

			// Get the Term_ID for the current word
			var termID int
			err = tx.QueryRow("SELECT Term_ID FROM Words WHERE word = ?", stemmedWord).Scan(&termID)
			if err != nil {
				tx.Rollback()
				log.Fatal(err)
			}

			termIDMap[stemmedWord] = termID

			// Update the URLs table to increment the total number of words in the document
			_, err = tx.Exec("UPDATE Urls SET word_count = word_count + 1 WHERE url_name = ?", cleanRes.url)
			if err != nil {
				tx.Rollback()
				log.Fatal(err)
			}

			// Insert into the Hits table
			_, err = tx.Exec("INSERT OR IGNORE INTO Hits (Term_ID, url_ID, Term_Count) VALUES (?, ?, 0)", termID, url_ID)
			if err != nil {
				tx.Rollback()
				log.Fatal(err)
			}

			// Update the Hits table to increment the total number of times a particular word appeared in the document
			_, err = tx.Exec("UPDATE Hits SET Term_Count = Term_Count + 1 WHERE Term_ID = ? AND url_ID = ?", termID, url_ID)
			if err != nil {
				tx.Rollback()
				log.Fatal(err)
			}
		}
	}

	// Function that inserts into the bigramhits table
	populateBigrams(stopWordsMap, cleanRes, url_ID, termIDMap, tx)

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

// Function that inserts into the bigramhits table
func populateBigrams(stopWordsMap map[string]struct{}, cleanRes ExtractResult, url_ID int, termIDMap map[string]int, tx *sql.Tx) {
	// Loop through all words to make the bigram hits table
	for j := 0; j < len(cleanRes.words)-1; j++ {
		stemmedWord1 := stemWords(cleanRes.words[j])
		stemmedWord2 := stemWords(cleanRes.words[j+1])

		// Skip bigramhits table insertion if either word is a stop word
		if _, isStop1 := stopWordsMap[stemmedWord1]; isStop1 {
			continue
		}

		if _, isStop2 := stopWordsMap[stemmedWord2]; isStop2 {
			continue
		}

		// Extract the term_ids of the 'i'th and 'i+1'th term
		term1ID := termIDMap[stemmedWord1]
		term2ID := termIDMap[stemmedWord2]

		_, err := tx.Exec("INSERT OR IGNORE INTO BigramHits (Term1_ID, Term2_ID, url_ID, Term_Count) VALUES (?, ?, ?, 0)",
			term1ID, term2ID, url_ID)
		if err != nil {
			log.Fatal(err)
		}

		// Update the term counts if already exists in table
		_, err = tx.Exec("UPDATE BigramHits SET Term_Count = Term_Count + 1 WHERE Term1_ID = ? AND Term2_ID = ? AND url_ID = ?", term1ID, term2ID, url_ID)
		if err != nil {
			log.Fatal(err)
		}
	}
}
