package app

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		worker.Start(ctx)
	}()
	go func() {
		defer wg.Done()
		drainer.Start(ctx)
	}()

	httpServer := &http.Server{
		Addr:    *httpAddr,
		Handler: handler,
	}

	go func() {
		<-ctx.Done()
		log.Printf("shutdown signal received; stopping HTTP server...")
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	log.Printf("%s", "MCP HTTP server listening on "+*httpAddr)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error starting HTTP server: %v", err)
	}

	wg.Wait()
	log.Printf("shutdown complete")
}
