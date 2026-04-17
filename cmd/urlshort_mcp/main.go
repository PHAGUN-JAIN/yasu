package urlshortmcp

import (
	"log"
	"urlshort/cmd/urlshort_mcp/app"
)

func main() {
	log.Println("Starting URL Shortener MCP Application...")
	app.RunApplication()
}
