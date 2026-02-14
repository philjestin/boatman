package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// DatadogMCPServer implements MCP protocol for Datadog with Okta OAuth
type DatadogMCPServer struct {
	accessToken string
	site        string
}

func main() {
	accessToken := os.Getenv("OKTA_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("OKTA_ACCESS_TOKEN environment variable is required")
	}

	site := os.Getenv("DD_SITE")
	if site == "" {
		site = "datadoghq.com"
	}

	server := &DatadogMCPServer{
		accessToken: accessToken,
		site:        site,
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

func (s *DatadogMCPServer) handleRequest(request map[string]interface{}) map[string]interface{} {
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

func (s *DatadogMCPServer) handleInitialize(request map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      request["id"],
		"result": map[string]interface{}{
			"protocolVersion": "1.0.0",
			"serverInfo": map[string]interface{}{
				"name":    "datadog-okta-mcp",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
		},
	}
}

func (s *DatadogMCPServer) handleToolsList() map[string]interface{} {
	tools := []map[string]interface{}{
		{
			"name":        "datadog_query_logs",
			"description": "Query Datadog logs with a search query",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Log search query",
					},
					"from": map[string]interface{}{
						"type":        "string",
						"description": "Start time (ISO 8601 or relative like '15m')",
					},
					"to": map[string]interface{}{
						"type":        "string",
						"description": "End time (ISO 8601 or 'now')",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "datadog_list_monitors",
			"description": "List Datadog monitors and their status",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"tags": map[string]interface{}{
						"type":        "string",
						"description": "Filter by tags (comma-separated)",
					},
				},
			},
		},
		{
			"name":        "datadog_get_metrics",
			"description": "Query Datadog metrics",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Metrics query (e.g., 'avg:system.cpu.user{*}')",
					},
					"from": map[string]interface{}{
						"type":        "string",
						"description": "Start time (unix timestamp or relative)",
					},
					"to": map[string]interface{}{
						"type":        "string",
						"description": "End time (unix timestamp or 'now')",
					},
				},
				"required": []string{"query", "from", "to"},
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

func (s *DatadogMCPServer) handleToolsCall(request map[string]interface{}) map[string]interface{} {
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
	case "datadog_query_logs":
		result, err = s.queryLogs(arguments)
	case "datadog_list_monitors":
		result, err = s.listMonitors(arguments)
	case "datadog_get_metrics":
		result, err = s.getMetrics(arguments)
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

func (s *DatadogMCPServer) queryLogs(args map[string]interface{}) (interface{}, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	url := fmt.Sprintf("https://api.%s/api/v2/logs/events/search", s.site)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Build request body
	body := map[string]interface{}{
		"filter": map[string]interface{}{
			"query": query,
			"from":  args["from"],
			"to":    args["to"],
		},
	}

	bodyBytes, _ := json.Marshal(body)
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Datadog API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *DatadogMCPServer) listMonitors(args map[string]interface{}) (interface{}, error) {
	url := fmt.Sprintf("https://api.%s/api/v1/monitor", s.site)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Datadog API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *DatadogMCPServer) getMetrics(args map[string]interface{}) (interface{}, error) {
	query, _ := args["query"].(string)
	from, _ := args["from"].(string)
	to, _ := args["to"].(string)

	if query == "" || from == "" || to == "" {
		return nil, fmt.Errorf("query, from, and to are required")
	}

	url := fmt.Sprintf("https://api.%s/api/v1/query?query=%s&from=%s&to=%s", s.site, query, from, to)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Datadog API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *DatadogMCPServer) errorResponse(message string) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"error": map[string]interface{}{
			"code":    -32603,
			"message": message,
		},
	}
}
