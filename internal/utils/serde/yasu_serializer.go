package serde

import (
	"log"
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
	return createYasuValidator(req)
}

func createYasuValidator(req service.ShortenRequest) error {
	// TODO: Implement validation logic for the ShortenRequest struct
	// 1. Validate that the URL field is a well-formed URL
	// 2. Validate that the ExpiresAt field, if provided, is a future date
	// 3. Validate that the CustomCode field, if provided, meets any specific requirements (e.g., length, allowed characters, etc.)
	log.Printf("Validating ShortenRequest: URL=%s, ExpiresAt=%v, CustomCode=%s", req.URL, req.ExpiresAt, req.CustomCode)
	return nil
}
