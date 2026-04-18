package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type YasuService struct {
	PgConn *sql.DB
	Cache  *redis.Client
}

type ShortenRequest struct {
	ExpiresAt  time.Time `json:"expires_at"`            // optional expiration time for the shortened URL
	CustomCode string    `json:"custom_code,omitempty"` // only used when IsCustom is true
	URL        string    `json:"url"`
}

func (s *YasuService) Get(ctx context.Context, code string) (string, error) {
	// TODO: Implement logic to retrieve the original URL based on the short ID
	// 1. Extract the short ID from the request URL ✅
	// 2. Query the database to find the corresponding original URL ✅
	// 3. Handle cases where the short ID does not exist or has expired ✅
	// 4. publish an event to buffer for asynchronous processing (e.g., analytics, cache warming, etc.)
	// 5. Return the original URL to the client

	cached, err := s.Cache.Get(ctx, "url:"+code).Result()
	if err == nil {
		return cached, nil
	}
	if err != redis.Nil {
		log.Printf("Redis GET error: %v", err)
	}

	var originalURL string
	var expiresAt time.Time

	err = s.PgConn.QueryRow(
		`SELECT original_url, expires_at 
         FROM short_urls 
         WHERE code = $1`,
		code,
	).Scan(&originalURL, &expiresAt)
	if err != nil {
		log.Printf("Error querying database for code %s: %v", code, err)
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", ErrInternal // wrap the error with additional context
	}

	if expiresAt.Before(time.Now().UTC()) {
		return "", ErrExpired
	}

	if ttl := time.Until(expiresAt); ttl > 0 {
		if err := s.Cache.Set(ctx, "url:"+code, originalURL, ttl).Err(); err != nil {
			log.Printf("Redis SET error: %v", err)
		}
	}

	return originalURL, nil
}

func (s *YasuService) Shorten(ctx context.Context, req ShortenRequest) (string, error) {
	// TODO: Implement URL shortening logic here
	// 1. Generate a unique short ID for the URL ✅
	// 2. Store the mapping of short ID to original URL in the database ✅
	// 3. publish an event to buffer for asynchronous processing (e.g., analytics, cache warming, etc.)
	// 4. Return the shortened URL to the client ✅

	code, err := generateCode()
	if err != nil {
		return "", err
	}

	req.ExpiresAt = time.Now().UTC().Add(7 * 24 * time.Hour)

	//Generate + retry (handle collision)
	for i := 0; i < 3; i++ {
		code, err = generateCode()
		if err != nil {
			log.Printf("Code generation error: %v", err)
			return "", ErrInternal
		}

		// TODO: move this to store layer
		res, err := s.PgConn.Exec(
			`INSERT INTO short_urls (code, original_url, expires_at)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (code) DO NOTHING`,
			code, req.URL, req.ExpiresAt,
		)
		if err != nil {
			log.Printf("DB insert error: %v", err)
			return "", ErrInternal
		}

		rows, _ := res.RowsAffected()
		if rows > 0 {
			break // success
		}

		code = "" // retry
	}

	if code == "" {
		return "", ErrInternal
	}

	if ttl := time.Until(req.ExpiresAt); ttl > 0 {
		if err := s.Cache.Set(ctx, "url:"+code, req.URL, ttl).Err(); err != nil {
			log.Printf("Redis SET error: %v", err)
		}
	}


	// enqueue analytics event — non-blocking, never slows down redirect
    // s.Worker.Enqueue(outbox.ClickEvent{
    //     Code:      code,
    //     Referer:   r.Header.Get("Referer"),   // pass from handler
    //     ClickedAt: time.Now().UTC(),
    //     // Country + Device added in Phase 4 (GeoIP)
    // })

	return code, nil
}

func (s *YasuService) Delete() string {
	// TODO: Implement logic to delete a shortened URL based on the short ID
	// 1. Extract the short ID from the request URL
	// 2. Delete the corresponding entry from the database
	// 3. Handle cases where the short ID does not exist
	// 4. Optionally, publish an event to buffer for asynchronous processing (e.g., analytics, cache warming, etc.)
	// 5. Return a success message to the client
	return "Hello from YasuService! I am DeleteYasu."
}

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateCode() (string, error) {
	bytes := make([]byte, 7)

	_, err := rand.Read(bytes) // Fill with cryptographically random bytes from the OS (crypto/rand not math/rand). Each byte is a random number 0-255.
	if err != nil {
		log.Printf("Error generating random bytes: %v", err)
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = alphabet[b%byte(len(alphabet))]
	}

	return string(bytes), nil
}
