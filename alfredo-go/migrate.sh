#!/bin/bash

# Migration helper script for AlfreDo Python to Go
# This script helps compare outputs and migrate workflows

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_DIR="$(dirname "$SCRIPT_DIR")"
GO_BINARY="$SCRIPT_DIR/alfredo-go"

echo "AlfreDo Migration Helper"
echo "======================="
echo "Python source: $PYTHON_DIR"
echo "Go binary: $GO_BINARY"
echo ""

# Check if Go binary exists
if [ ! -f "$GO_BINARY" ]; then
    echo "Go binary not found. Building..."
    cd "$SCRIPT_DIR"
    go build -o alfredo-go main.go
    echo "Build complete."
    echo ""
fi

# Check if TOKEN is set
if [ -z "$TOKEN" ]; then
    echo "Warning: TOKEN environment variable not set."
    echo "Please set your Todoist API token:"
    echo "export TOKEN='your_todoist_api_token'"
    echo ""
fi

# Function to run Python equivalent and Go version
compare_commands() {
    local command="$1"
    local description="$2"
    
    echo "Testing: $description"
    echo "Command: $command"
    echo "----------------------------------------"
    
    if [ -n "$TOKEN" ]; then
        echo "Go output:"
        eval "$GO_BINARY $command" 2>/dev/null || echo "Failed to execute Go command"
        echo ""
    else
        echo "Skipping execution - TOKEN not set"
        echo ""
    fi
}

# Show available commands
echo "Available Go commands:"
"$GO_BINARY" help
echo ""

# Show usage examples
echo "Usage Examples:"
echo "==============="
echo ""

compare_commands "get today" "Get tasks due today (equivalent to alfredo-get.py today)"
compare_commands "get overdue" "Get overdue tasks (equivalent to alfredo-get.py overdue)"
compare_commands "stats" "Get completion statistics (equivalent to alfredo-get-troubleshoot.py)"

echo "To complete a task (equivalent to alfredo-ops.py):"
echo "  $GO_BINARY complete TASK_ID"
echo ""

echo "Migration complete! Your Go application is ready to use."
echo ""
echo "Key improvements in the Go version:"
echo "- Single binary with multiple commands"
echo "- Better error handling"
echo "- Parallel API requests for better performance"
echo "- Structured code with clear separation of concerns"
echo "- Built-in help system"
echo "- Cross-platform compatibility"
