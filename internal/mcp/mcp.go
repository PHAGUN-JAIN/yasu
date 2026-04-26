package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"urlshort/cmd/initialize"
	ob "urlshort/internal/service/outbox"
	svc "urlshort/internal/service/urlshort"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type YasuMCPHandler struct {
	svc   *svc.YasuService
	stats *svc.AnalyticsService
}

// typed input structs — SDK auto-infers JSON schema from these
type ShortenInput struct {
	URL string `json:"url" jsonschema:"the URL to shorten"`
}

type ShortenOutput struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

type ResolveInput struct {
	Code string `json:"code" jsonschema:"the short code to resolve"`
}

type ResolveOutput struct {
	OriginalURL string `json:"original_url"`
}

func loggingMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return func(
		ctx context.Context,
		method string,
		req mcp.Request,
	) (mcp.Result, error) {
		logger.Info("MCP method started",
			"method", method,
			"session_id", req.GetSession().ID(),
			"has_params", req.GetParams() != nil,
		)
		// Log more for tool calls.
		if ctr, ok := req.(*mcp.CallToolRequest); ok {
			logger.Info("Calling tool",
				"name", ctr.Params.Name,
				"args", ctr.Params.Arguments)
		}

		start := time.Now()
		result, err := next(ctx, method, req)
		duration := time.Since(start)
		if err != nil {
			logger.Error("MCP method failed",
				"method", method,
				"session_id", req.GetSession().ID(),
				"duration_ms", duration.Milliseconds(),
				"err", err,
			)
		} else {
			logger.Info("MCP method completed",
				"method", method,
				"session_id", req.GetSession().ID(),
				"duration_ms", duration.Milliseconds(),
				"has_result", result != nil,
			)
			// Log more for tool results.
			if ctr, ok := result.(*mcp.CallToolResult); ok {
				logger.Info("tool result",
					"isError", ctr.IsError,
					"structuredContent", ctr.StructuredContent)
			}
		}
		return result, err
	}
}

func addTools(server *mcp.Server, h *YasuMCPHandler) {
	pingTool := &mcp.Tool{
		Name:        "ping",
		Description: "Ping the server",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "Message to send",
				},
			},
		},
	}
	server.AddTool(pingTool, h.pingingToolHandler)

	// Tool 1 — shorten a URL
	shortenTool := &mcp.Tool{
		Name:        "shorten_url",
		Description: "Shortens a long URL and returns a short code",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "The long URL to shorten",
				},
			},
			"required": []string{"url"},
		},
	}
	server.AddTool(shortenTool, h.shortenHandler)

	// Tool 2 — resolve a short code
	resolveTool := &mcp.Tool{
		Name:        "resolve_url",
		Description: "Resolves a short code back to the original URL",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"code": map[string]interface{}{
					"type":        "string",
					"description": "The short code to resolve",
				},
			},
			"required": []string{"code"},
		},
	}
	server.AddTool(resolveTool, h.resolveHandler)

	analyticsTool := &mcp.Tool{
		Name:        "get_analytics",
		Description: "Gets analytics data for a given short code",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"code": map[string]interface{}{
					"type":        "string",
					"description": "The short code to get analytics for",
				},
			},
			"required": []string{"code"},
		},
	}
	server.AddTool(analyticsTool, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		log.Printf("Got analytics request: %v", req.Params)

		var input struct {
			Code string `json:"code"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		if input.Code == "" {
			return nil, fmt.Errorf("code is required")
		}

		analyticsData, err := h.stats.Get(ctx, input.Code)
		if err != nil {
			return nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("analytics for code %s: %v", input.Code, analyticsData)},
			},
			StructuredContent: analyticsData,
		}, nil
	})

}

func (h *YasuMCPHandler) pingingToolHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Printf("Got ping request: %v", req.Params)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "pong", Meta: map[string]any{"ping": "pong"}},
		},
	}, nil
}

func (h *YasuMCPHandler) shortenHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Printf("Got shorten request: %v", req.Params)

	var input ShortenInput
	if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	if input.URL == "" {
		return nil, fmt.Errorf("url is required")
	}

	// Here you would add your URL shortening logic, e.g. generate a code and store the mapping
	code := ""

	code, err := h.svc.Shorten(ctx, svc.ShortenRequest{
		URL:       input.URL,
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("shortened URL with code: %s", code)},
		},
		StructuredContent: ShortenOutput{
			Code:     code,
			ShortURL: "http://localhost:8000/" + code,
		},
	}, nil
}

func (h *YasuMCPHandler) resolveHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Printf("Got resolve request: %v", req.Params)

	var input ResolveInput
	if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	if input.Code == "" {
		return nil, fmt.Errorf("code is required")
	}

	// Here you would add your URL resolving logic, e.g. look up the code and return the original URL
	originalURL := ""

	originalURL, err := h.svc.Get(ctx, input.Code, nil, ob.SourceMCP) // TODO: pass actual code and headers
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("original URL is: %s", originalURL)},
		},
		StructuredContent: ResolveOutput{
			OriginalURL: originalURL,
		},
	}, nil
}

func NewMCPServer(clients *initialize.Clients) (*mcp.Server, *ob.Worker, *ob.Drainer) {

	worker := ob.NewWorker(clients.PgClient)
	drainer := ob.NewDrainer(clients.PgClient, clients.ChClient)

	h := &YasuMCPHandler{
		svc:   svc.NewYasuService(clients.PgClient, clients.Cache, worker),
		stats: svc.NewAnalyticsService(clients.ChClient),
	}
	// TODO : add prompts and resources
	opts := &mcp.ServerOptions{
		Instructions: "This is the MCP server for Yasu, a URL shortening service. It provides tools to shorten URLs and resolve short codes back to original URLs.",
		HasTools:     true,
		HasPrompts:   false,
		HasResources: false,
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "yasen-mcp",
		Version: "v0.0.1",
	}, opts)

	server.AddReceivingMiddleware(loggingMiddleware)

	addTools(server, h)
	// addResources(server)
	// addPrompts(server)

	return server, worker, drainer
}
