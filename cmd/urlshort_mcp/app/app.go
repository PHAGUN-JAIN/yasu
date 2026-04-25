package app

import (
	"context"
	"flag"
	"log"
	"net/http"
	"urlshort/cmd/initialize"
	urlshortmcp "urlshort/internal/mcp"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var httpAddr = flag.String("http", ":8001", "HTTP address to listen on")

func RunApplication() {

	// stdio transport need to figure out http streamable transport
	// t := &mcp.LoggingTransport{Transport: &mcp.StdioTransport{}, Writer: os.Stderr}
	//if err := mcpServer.Run(context.Background(), t); err != nil {
	//	log.Panic("Error initializing MCP server", log.ErrorField(err))
	//}
	clients := initialize.GetClients()

	server, worker, drainer := urlshortmcp.NewMCPServer(clients)

	handler := mcp.NewStreamableHTTPHandler(
		func(req *http.Request) *mcp.Server {
			return server
		},
		nil,
	)

	go worker.Start(context.Background())
	go drainer.Start(context.Background())

	log.Printf("%s", "MCP HTTP server listening on "+*httpAddr)

	if err := http.ListenAndServe(*httpAddr, handler); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
