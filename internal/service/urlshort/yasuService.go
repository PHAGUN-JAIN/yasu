package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"
	"urlshort/internal/service/outbox"
	"urlshort/internal/utils"

	"github.com/redis/go-redis/v9"
)

const cacheWriteTimeout = 200 * time.Millisecond

type YasuService struct {
	PgConn *sql.DB
	Cache  *redis.Client
	worker *outbox.Worker
	geo    *utils.Resolver
}

type ShortenRequest struct {
	ExpiresAt  time.Time `json:"expires_at"`            // optional expiration time for the shortened URL
	CustomCode string    `json:"custom_code,omitempty"` // only used when IsCustom is true
	URL        string    `json:"url"`
}

func (s *YasuService) Get(ctx context.Context, code string, header http.Header, source string) (string, error) {
	// 1. Redis Cache Lookup
	cached, err := s.Cache.Get(ctx, "url:"+code).Result()
	if err == nil {
		s.trackClick(code, header, source)
		return cached, nil
	}

	if err != redis.Nil {
		log.Printf("Redis GET error for %s: %v", code, err)
	}

	// 2. Postgres Lookup
	var originalURL string
	var expiresAt time.Time

	err = s.PgConn.QueryRowContext(ctx, // Use QueryRowContext for better timeout handling
		`SELECT original_url, expires_at 
         FROM short_urls 
         WHERE code = $1`,
		code,
	).Scan(&originalURL, &expiresAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		log.Printf("DB error for code %s: %v", code, err)
		return "", ErrInternal
	}

	// 3. Expiration Check
	if !expiresAt.IsZero() && expiresAt.Before(time.Now().UTC()) {
		return "", ErrExpired
	}

	// 4. Update Cache synchronously with a short timeout so it doesn't hold up the request for long
	ttl := time.Until(expiresAt)
	if ttl > 0 || expiresAt.IsZero() {
		if expiresAt.IsZero() {
			ttl = 24 * time.Hour
		} // Default TTL for permanent links
		cacheCtx, cancel := context.WithTimeout(ctx, cacheWriteTimeout)
		defer cancel()
		if err := s.Cache.Set(cacheCtx, "url:"+code, originalURL, ttl).Err(); err != nil {
			log.Printf("Redis SET error for key url:%s: %v", code, err)
		}
	}

	// 5. Track Analytics
	s.trackClick(code, header, source)

	return originalURL, nil
}

// Helper to keep the main logic clean
func (s *YasuService) trackClick(code string, header http.Header, source string) {

	ip := utils.RealIP(header)
	country := s.geo.CountryCode(ip)
	city := s.geo.City(ip)
	asn := s.geo.ASN(ip)

	var referer string
	var device string
	if header != nil {
		referer = header.Get("Referer")
		device = header.Get("User-Agent")
	}
	if referer == "" {
		referer = "direct"
	}
	if device == "" {
		device = "unknown"
	}

	s.worker.Enqueue(outbox.ClickEvent{
		Code:      code,
		Referer:   referer,
		Device:    device,
		Source:    source,
		IP:        ip,
		Country:   country,
		City:      city,
		ASN:       asn,
		ClickedAt: time.Now().UTC(),
	})
}

func (s *YasuService) Shorten(ctx context.Context, req ShortenRequest) (string, error) {
	// TODO: Implement URL shortening logic here
	// 1. Generate a unique short ID for the URL ✅
	// 2. Store the mapping of short ID to original URL in the database ✅
	// 3. publish an event to buffer for asynchronous processing (e.g., analytics, cache warming, etc.)
	// 4. Return the shortened URL to the client ✅

	req.ExpiresAt = time.Now().UTC().Add(7 * 24 * time.Hour)

	var code string
	var err error
	//Generate + retry (handle collision)
	for range 3 {
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

func NewYasuService(pg *sql.DB, cache *redis.Client, worker *outbox.Worker, geoDbPath string) *YasuService {
	return &YasuService{
		PgConn: pg,
		Cache:  cache,
		worker: worker,
		geo:    utils.NewResolver(geoDbPath),
	}
}
