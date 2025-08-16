#!/bin/bash

# Generate Swagger documentation script
# This script generates Swagger docs using swag CLI

set -e

echo "🔄 Generating Swagger documentation..."

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo "📦 Installing swag CLI..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Add Go bin to PATH if not already there
export PATH=$PATH:$(go env GOPATH)/bin

# Generate swagger docs
echo "📝 Running swag init..."
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

echo "✅ Swagger documentation generated successfully!"
echo "📁 Documentation files created in ./docs/"
echo "🌐 Start the server and visit http://localhost:8080/swagger/index.html to view the API documentation"

# List generated files
echo ""
echo "Generated files:"
ls -la docs/