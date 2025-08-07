#!/bin/bash

# Install pre-commit hooks for aaa-service
set -e

echo "🔧 Setting up pre-commit hooks for aaa-service..."

# Check if pre-commit is installed
if ! command -v pre-commit &> /dev/null; then
    echo "❌ pre-commit not found. Installing via pip..."
    pip install pre-commit
fi

# Install pre-commit hooks
echo "📦 Installing pre-commit hooks..."
pre-commit install

# Install commit message hooks
echo "📝 Installing commit message hooks..."
pre-commit install --hook-type commit-msg

# Run pre-commit on all files to ensure everything works
echo "🧪 Running pre-commit on all files..."
pre-commit run --all-files || {
    echo "⚠️  Some pre-commit checks failed. This is expected if there are existing issues."
    echo "   Continue with fixing the issues and the hooks will work for future commits."
}

echo "✅ Pre-commit hooks installed successfully!"
echo ""
echo "📋 Available hooks:"
echo "  - go-fmt: Format Go code"
echo "  - go-imports: Fix Go imports"
echo "  - go-mod-tidy: Tidy Go modules"
echo "  - go-unit-tests: Run unit tests"
echo "  - golangci-lint: Lint Go code"
echo "  - gosec: Security scanning"
echo "  - conventional-pre-commit: Check commit message format"
echo ""
echo "🎯 Hooks will run automatically on each commit."
echo "   To run manually: pre-commit run --all-files" 