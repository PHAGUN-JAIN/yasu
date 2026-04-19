package app

import (
	"flag"
	"log"
	"net/http"
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

	handler := mcp.NewStreamableHTTPHandler(
		func(req *http.Request) *mcp.Server {
			server := urlshortmcp.NewMCPServer()
			return server
		},
		nil,
	)

	log.Printf("%s", "MCP HTTP server listening on "+*httpAddr)

	if err := http.ListenAndServe(*httpAddr, handler); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}

// AuthenticateSession Helper to authenticate and store session on initializing
// func AuthenticateSession(sessionStore *authMcp.SessionStore, accessToken, tenantID string) string {
// 	sessionID := authMcp.GenerateSessionID()
// 	sessionStore.Set(sessionID, &authMcp.SessionContext{
// 		AccessToken: accessToken,
// 		TenantID:    tenantID,
// 		CreatedAt:   time.Now(),
// 		LastAccess:  time.Now(),
// 	})
// 	log.Infof("Created session %s for tenant %s", sessionID, tenantID)
// 	return sessionID
// }
