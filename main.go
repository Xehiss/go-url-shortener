package main

import (
    "crypto/md5"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbName      = "urlshortener.db"
	base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

type URL struct {
	Hash      string
	URL       string
	ExpiresAt time.Time
}

// This function will handle the redirection.
func redirectHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := r.URL.Path[3:] // Strip "/r/" from the URL path to get the hash

		var originalURL string
		err := db.QueryRow("SELECT url FROM urls WHERE hash = ?", hash).Scan(&originalURL)

		if err != nil {
			if err == sql.ErrNoRows {
				http.NotFound(w, r)
			} else {
				log.Printf("Database error: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		http.Redirect(w, r, originalURL, http.StatusFound)
	}
}


// / This function will handle creating a short URL.
func createHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlStr := r.URL.Query().Get("url")
		if urlStr == "" {
			http.Error(w, "Missing URL parameter", http.StatusBadRequest)
			return
		}

		// Make sure the URL starts with http:// or https://
		if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
			urlStr = "http://" + urlStr
		}

		// Validate the URL
		if !isValidURL(urlStr) {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		// Generate a hash of the URL
		hash := md5.Sum([]byte(urlStr))

		// Convert the first 8 bytes of the hash into a large integer
		num := uint64(0)
		for i := 0; i < 8; i++ {
			num = num*256 + uint64(hash[i])
		}

		// Convert the number into a base62 string
		hashStr := toBase62(num)

		// Check if the hash already exists in the database
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM urls WHERE hash = ?", hashStr).Scan(&count)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("Failed to query url: %v\n", err)
			return
		}

		// If the hash does not exist, insert it into the database
		if count == 0 {
			// The URL will expire after 24 hours
			expiresAt := time.Now().Add(24 * time.Hour)

			_, err = db.Exec("INSERT INTO urls(hash, url, expires_at) VALUES (?, ?, ?)", hashStr, urlStr, expiresAt)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				log.Printf("Failed to insert url: %v\n", err)
				return
			}
		}

		shortURL := fmt.Sprintf("http://localhost:8080/r/%s", hashStr)
		fmt.Fprintln(w, shortURL)
	}
}

// Convert a number to a base62 string
func toBase62(num uint64) string {
	str := ""
	for num > 0 {
		str = string(base62Chars[num%62]) + str
		num = num / 62
	}
	return str
}

// Validate the URL
func isValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func setupDatabase(db *sql.DB) {
    createTableSQL := `CREATE TABLE IF NOT EXISTS urls (
        "hash" TEXT NOT NULL PRIMARY KEY,		
        "url" TEXT NOT NULL,
        "expires_at" DATETIME
    );`

    statement, err := db.Prepare(createTableSQL) 
    if err != nil {
        log.Fatalf("Failed to prepare database setup statement: %v", err)
    }

    _, err = statement.Exec()
    if err != nil {
        log.Fatalf("Failed to setup database: %v", err)
    }
}


func main() {
	// Database setup
	db, err := sql.Open("sqlite3", "./urls.db")
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

    setupDatabase(db)
	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Handlers
	http.HandleFunc("/create", createHandler(db))
	http.HandleFunc("/r/", redirectHandler(db))

	log.Println("Server listening on port 8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

