#!/bin/bash

# Build script for Angular frontend integration with Go backend
# This script builds the Angular application and prepares it for embedding in Go

set -e

echo "🚀 Building Superintelligence Dependency Manager Frontend..."

# Change to the web directory
cd "$(dirname "$0")/../web"

# Check if Node.js and npm are installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18+ to continue."
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo "❌ npm is not installed. Please install npm to continue."
    exit 1
fi

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "📦 Installing Angular dependencies..."
    npm install
fi

# Build the Angular application for production
echo "🔨 Building Angular application..."
npm run build

# Verify the build output
if [ ! -d "dist" ]; then
    echo "❌ Build failed: dist directory not found"
    exit 1
fi

# Create the embedded filesystem directory structure for Go
echo "📁 Preparing embedded filesystem structure..."
mkdir -p ../internal/web/dist
cp -r dist/* ../internal/web/dist/

# Update the Go embed directive to include the built files
echo "🔧 Updating Go embed directive..."

# Create a temporary Go file with the correct embed directive
cat > ../internal/web/embed.go << 'EOF'
package web

import "embed"

//go:embed dist/*
var staticFiles embed.FS
EOF

echo "✅ Frontend build completed successfully!"
echo ""
echo "📊 Build Summary:"
echo "   - Angular application built for production"
echo "   - Static files prepared for Go embedding"
echo "   - Ready for unified deployment"
echo ""
echo "🚀 Next steps:"
echo "   - Run 'make build' to build the complete application"
echo "   - Run './superint-dep-manager' to start the unified server"
echo ""
