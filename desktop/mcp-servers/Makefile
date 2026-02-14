.PHONY: all clean datadog-okta bugsnag-okta

all: datadog-okta bugsnag-okta

datadog-okta:
	@echo "Building datadog-okta MCP server..."
	cd datadog-okta && go build -o datadog-okta main.go

bugsnag-okta:
	@echo "Building bugsnag-okta MCP server..."
	cd bugsnag-okta && go build -o bugsnag-okta main.go

clean:
	@echo "Cleaning MCP server binaries..."
	rm -f datadog-okta/datadog-okta bugsnag-okta/bugsnag-okta

install: all
	@echo "Installing MCP servers to ~/.claude/mcp-servers/..."
	mkdir -p ~/.claude/mcp-servers
	cp datadog-okta/datadog-okta ~/.claude/mcp-servers/
	cp bugsnag-okta/bugsnag-okta ~/.claude/mcp-servers/
	@echo "MCP servers installed successfully!"
