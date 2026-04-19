package mcp

import (
	"context"
	"encoding/json"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewMCPServer() *mcp.Server {
	// TODO : add prompts and resources
	opts := &mcp.ServerOptions{
		Instructions: "Use this server for fetching logged in user information in a ncx network via ncx-auth service for a tenant",
		HasTools:     true,
		HasPrompts:   false,
		HasResources: false,
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ncx-auth",
		Version: "v0.0.1",
	}, opts)

	addTools(server)
	addResources(server)
	addPrompts(server)

	return server
}

func addResources(server *mcp.Server) {
	// IDPs available
	server.AddResource(&mcp.Resource{
		URI:         "idps://available",
		Name:        "Available IDPs",
		Description: "ZTNA identity providers",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		idps := []map[string]string{{"name": "ZTNA", "type": "saml", "status": "active"}}
		data, _ := json.Marshal(idps)
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(data),
				},
			},
		}, nil
	})

	// Dashboard users
	server.AddResource(&mcp.Resource{
		URI:         "users://dashboard/{network_id}",
		Name:        "Dashboard Users",
		Description: "Users logged in via dashboard",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		users := []map[string]string{{"user_id": "user123", "email": "user@example.com", "status": "active"}}
		data, _ := json.Marshal(users)
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(data),
				},
			},
		}, nil
	})

	// PRA policies
	server.AddResource(&mcp.Resource{
		URI:         "policies://pra/{tenant_id}/{network_id}",
		Name:        "PRA Policies",
		Description: "Policies per tenant/network",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		policies := []map[string]string{{"policy_id": "pol123", "name": "Default Policy", "type": "access"}}
		data, _ := json.Marshal(policies)
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(data),
				},
			},
		}, nil
	})

	// User resources
	server.AddResource(&mcp.Resource{
		URI:         "resources://pra/{user_id}",
		Name:        "User Resources",
		Description: "Resources available to user",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		resources := []map[string]string{{"resource_id": "res123", "name": "App Server", "type": "application"}}
		data, _ := json.Marshal(resources)
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(data),
				},
			},
		}, nil
	})
}

func addTools(server *mcp.Server) {
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
	server.AddTool(pingTool, pingingToolHandler)

}

func addPrompts(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "auth_help",
		Description: "Get help with authentication flows",
		Arguments: []*mcp.PromptArgument{
			{Name: "flow_type", Description: "Type of auth flow (oauth2, saml)", Required: false},
		},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		flowType := "general"

		var content string
		switch flowType {
		case "oauth2":
			content = "OAuth2 flow: /oauth2/authorize -> /oauth2/tokens"
		case "saml":
			content = "SAML flow: IDP -> ACS endpoint"
		default:
			content = "NCX Auth supports OAuth2 and SAML flows"
		}

		return &mcp.GetPromptResult{
			Description: "Authentication guidance",
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: content},
				},
			},
		}, nil
	})
}

func pingingToolHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Printf("Got ping request: %v", req.Params)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "pong", Meta: map[string]any{"ping": "pong"}},
		},
	}, nil
}
