#!/bin/bash
set -e

echo "🚢 Setting up Boatman Ecosystem..."
echo ""

# Check Go version
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.24.1 or later."
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✓ Go $GO_VERSION found"

# Check Node.js
if ! command -v node &> /dev/null; then
    echo "⚠️  Node.js not found. Desktop app requires Node.js 18+."
    echo "   You can still build the CLI."
else
    NODE_VERSION=$(node --version)
    echo "✓ Node.js $NODE_VERSION found"
fi

# Check Wails
if ! command -v wails &> /dev/null; then
    echo "⚠️  Wails not found. Desktop app requires Wails v2."
    echo "   Install: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    echo "   You can still build the CLI."
else
    WAILS_VERSION=$(wails version 2>&1 | grep "Wails" | awk '{print $3}' || echo "unknown")
    echo "✓ Wails $WAILS_VERSION found"
fi

echo ""
echo "📦 Syncing Go workspace..."
go work sync

echo ""
echo "📥 Downloading CLI dependencies..."
cd cli && go mod download && cd ..

echo ""
echo "🔨 Building CLI..."
make build-cli

if command -v node &> /dev/null; then
    echo ""
    echo "📥 Installing desktop frontend dependencies..."
    cd desktop/frontend && npm install && cd ../..
fi

echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "  • Build desktop: make build-desktop"
echo "  • Run CLI: ./cli/boatman work --prompt 'your task'"
echo "  • Dev mode: make dev"
echo "  • See all commands: make help"
