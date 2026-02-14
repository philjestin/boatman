package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// BugsnagMCPServer implements MCP protocol for Bugsnag with Okta OAuth
type BugsnagMCPServer struct {
	accessToken string
}

func main() {
	accessToken := os.Getenv("OKTA_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("OKTA_ACCESS_TOKEN environment variable is required")
	}

	server := &BugsnagMCPServer{
		accessToken: accessToken,
	}

	// Read MCP requests from stdin, write responses to stdout
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var request map[string]interface{}
		if err := decoder.Decode(&request); err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error decoding request: %v", err)
			continue
		}

		response := server.handleRequest(request)
		if err := encoder.Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

func (s *BugsnagMCPServer) handleRequest(request map[string]interface{}) map[string]interface{} {
	method, ok := request["method"].(string)
	if !ok {
		return s.errorResponse("invalid method")
	}

	switch method {
	case "initialize":
		return s.handleInitialize(request)
	case "tools/list":
		return s.handleToolsList()
	case "tools/call":
		return s.handleToolsCall(request)
	default:
		return s.errorResponse(fmt.Sprintf("unknown method: %s", method))
	}
}

func (s *BugsnagMCPServer) handleInitialize(request map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      request["id"],
		"result": map[string]interface{}{
			"protocolVersion": "1.0.0",
			"serverInfo": map[string]interface{}{
				"name":    "bugsnag-okta-mcp",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
		},
	}
}

func (s *BugsnagMCPServer) handleToolsList() map[string]interface{} {
	tools := []map[string]interface{}{
		{
			"name":        "bugsnag_list_projects",
			"description": "List all Bugsnag projects",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			"name":        "bugsnag_list_errors",
			"description": "List recent errors for a project",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Bugsnag project ID",
					},
					"filters": map[string]interface{}{
						"type":        "object",
						"description": "Optional filters (release_stage, severity, etc.)",
					},
				},
				"required": []string{"project_id"},
			},
		},
		{
			"name":        "bugsnag_get_error",
			"description": "Get detailed information about a specific error",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Bugsnag project ID",
					},
					"error_id": map[string]interface{}{
						"type":        "string",
						"description": "Error ID",
					},
				},
				"required": []string{"project_id", "error_id"},
			},
		},
		{
			"name":        "bugsnag_list_events",
			"description": "List events (occurrences) for a specific error",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Bugsnag project ID",
					},
					"error_id": map[string]interface{}{
						"type":        "string",
						"description": "Error ID",
					},
				},
				"required": []string{"project_id", "error_id"},
			},
		},
	}

	return map[string]interface{}{
		"jsonrpc": "2.0",
		"result": map[string]interface{}{
			"tools": tools,
		},
	}
}

func (s *BugsnagMCPServer) handleToolsCall(request map[string]interface{}) map[string]interface{} {
	params, ok := request["params"].(map[string]interface{})
	if !ok {
		return s.errorResponse("invalid params")
	}

	name, ok := params["name"].(string)
	if !ok {
		return s.errorResponse("missing tool name")
	}

	arguments, _ := params["arguments"].(map[string]interface{})

	var result interface{}
	var err error

	switch name {
	case "bugsnag_list_projects":
		result, err = s.listProjects()
	case "bugsnag_list_errors":
		result, err = s.listErrors(arguments)
	case "bugsnag_get_error":
		result, err = s.getError(arguments)
	case "bugsnag_list_events":
		result, err = s.listEvents(arguments)
	default:
		return s.errorResponse(fmt.Sprintf("unknown tool: %s", name))
	}

	if err != nil {
		return s.errorResponse(err.Error())
	}

	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      request["id"],
		"result": map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("%v", result),
				},
			},
		},
	}
}

func (s *BugsnagMCPServer) listProjects() (interface{}, error) {
	url := "https://api.bugsnag.com/user/organizations"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("X-Version", "2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Bugsnag API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *BugsnagMCPServer) listErrors(args map[string]interface{}) (interface{}, error) {
	projectID, ok := args["project_id"].(string)
	if !ok || projectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}

	url := fmt.Sprintf("https://api.bugsnag.com/projects/%s/errors", projectID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("X-Version", "2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Bugsnag API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *BugsnagMCPServer) getError(args map[string]interface{}) (interface{}, error) {
	projectID, _ := args["project_id"].(string)
	errorID, _ := args["error_id"].(string)

	if projectID == "" || errorID == "" {
		return nil, fmt.Errorf("project_id and error_id are required")
	}

	url := fmt.Sprintf("https://api.bugsnag.com/projects/%s/errors/%s", projectID, errorID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("X-Version", "2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Bugsnag API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *BugsnagMCPServer) listEvents(args map[string]interface{}) (interface{}, error) {
	projectID, _ := args["project_id"].(string)
	errorID, _ := args["error_id"].(string)

	if projectID == "" || errorID == "" {
		return nil, fmt.Errorf("project_id and error_id are required")
	}

	url := fmt.Sprintf("https://api.bugsnag.com/projects/%s/errors/%s/events", projectID, errorID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("X-Version", "2")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Bugsnag API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *BugsnagMCPServer) errorResponse(message string) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"error": map[string]interface{}{
			"code":    -32603,
			"message": message,
		},
	}
}
