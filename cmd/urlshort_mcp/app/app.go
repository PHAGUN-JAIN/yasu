package app

import (
	router "urlshort/internal/http"
)

func RunApplication() {

	router.RunYasuMcpServer("8001")
}
