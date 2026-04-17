package serde

import (
	"errors"
	"log"
	"net/url"
	"time"
	"urlshort/internal/service"
)

type YasuSerializerInf interface {
	YasuDeserializer(req service.ShortenRequest) error
	YasuSerializer() error
}

type YasuSerializer struct {
}

func (s *YasuSerializer) YasuSerializer() error {
	return nil
}

func (s *YasuSerializer) YasuDeserializer(req service.ShortenRequest) error {
	return validator(req)
}

func validator(req service.ShortenRequest) error {
	// TODO: Implement validation logic for the ShortenRequest struct
	// 1. Validate that the URL field is a well-formed URL
	// 2. Validate that the ExpiresAt field, if provided, is a future date
	// 3. Validate that the CustomCode field, if provided, meets any specific requirements (e.g., length, allowed characters, etc.)
	log.Printf("Validating ShortenRequest: URL=%s, ExpiresAt=%v, CustomCode=%s", req.URL, req.ExpiresAt, req.CustomCode)

	_, err := url.ParseRequestURI(req.URL)
	if err != nil {
		return err
	}
	if req.ExpiresAt != (time.Time{}) && req.ExpiresAt.Before(time.Now()) {
		return errors.New("expiration time must be in the future")
	}
	if req.ExpiresAt.Equal((time.Time{})) {
		req.ExpiresAt = time.Now().UTC().Add(7 * 24 * time.Hour) // default expiration time of 7 days
	}
	if req.CustomCode != "" && !isValidCustomCode(req.CustomCode) {
		return errors.New("custom code must be between 4 and 10 characters")
	}
	return nil
}

func isValidCustomCode(s string) bool {
	return len(s) >= 4 && len(s) <= 10 // example validation: custom code must be between 4 and 10 characters
}
