# Python to Go Migration Guide

## Overview

This document outlines the migration from the Python AlfreDo scripts to a unified Go application.

## File Mapping

| Python File | Go Equivalent | Description |
|-------------|---------------|-------------|
| `alfredo-get.py` | `./alfredo-go get [today\|overdue]` | Get tasks from Todoist |
| `alfredo-ops.py` | `./alfredo-go complete [task-id]` | Complete a task |
| `alfredo-get-troubleshoot.py` | `./alfredo-go stats` | Get completion statistics |
| `alfredo_fun.py` | `pkg/utils/log.go` | Logging utilities |
| `config.py` | `pkg/config/config.go` | Configuration management |
| N/A | `main.go` | Single entry point |

## Command Equivalence

### Getting Tasks

**Python (alfredo-get.py):**
```bash
python3 alfredo-get.py today
python3 alfredo-get.py overdue
```

**Go:**
```bash
./alfredo-go get today
./alfredo-go get overdue
```

### Completing Tasks

**Python (alfredo-ops.py):**
```bash
python3 alfredo-ops.py TASK_ID
```

**Go:**
```bash
./alfredo-go complete TASK_ID
```

### Getting Statistics

**Python (alfredo-get-troubleshoot.py):**
```bash
python3 alfredo-get-troubleshoot.py
```

**Go:**
```bash
./alfredo-go stats
```

## Configuration

### Python
```python
# config.py
import os
TOKEN = os.path.expanduser(os.getenv('TOKEN', ''))
```

### Go
```go
// pkg/config/config.go
func LoadConfig() *Config {
    token := os.Getenv("TOKEN")
    if token == "" {
        token = os.Getenv("TODOIST_TOKEN")
    }
    return &Config{Token: token}
}
```

## API Client

### Python
```python
# Direct HTTP requests using requests library
headers = CaseInsensitiveDict()
headers["Authorization"] = "Bearer " + TOKEN
resp = requests.post(url_sync, headers=headers, data=data)
```

### Go
```go
// Structured client in pkg/todoist/client.go
type Client struct {
    token      string
    httpClient *http.Client
    baseURL    string
    syncURL    string
}

func (c *Client) GetTasks() (*SyncResponse, error) {
    // Structured HTTP requests with proper error handling
}
```

## Error Handling

### Python
```python
# Basic error handling
resp = requests.get(url, headers=headers)
myData = resp.json()
```

### Go
```go
// Comprehensive error handling
resp, err := c.httpClient.Do(req)
if err != nil {
    return nil, err
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
    return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
}
```

## Improvements in Go Version

### 1. Single Binary
- **Python**: Multiple script files
- **Go**: Single binary with subcommands

### 2. Performance
- **Python**: Sequential API calls
- **Go**: Parallel API requests using goroutines

### 3. Error Handling
- **Python**: Basic error handling
- **Go**: Comprehensive error handling with proper error types

### 4. Code Organization
- **Python**: Functional approach with shared utilities
- **Go**: Clean architecture with separated concerns:
  - `cmd/`: Command-line interface
  - `internal/`: Business logic
  - `pkg/`: Reusable packages

### 5. Type Safety
- **Python**: Dynamic typing
- **Go**: Static typing with compile-time checks

### 6. Deployment
- **Python**: Requires Python runtime + dependencies
- **Go**: Single binary, no runtime dependencies

### 7. Documentation
- **Python**: Comments in code
- **Go**: Built-in help system via Cobra

## Building and Deployment

### Python
```bash
# Requires Python 3 and pip dependencies
pip3 install requests
python3 alfredo-get.py today
```

### Go
```bash
# Build once, run anywhere
go build -o alfredo-go main.go
./alfredo-go get today
```

## Cross-Platform Support

### Python
- Requires Python interpreter on target system
- Dependency management needed

### Go
- Compile for different platforms:
```bash
GOOS=darwin GOARCH=amd64 go build -o alfredo-go-mac main.go
GOOS=linux GOARCH=amd64 go build -o alfredo-go-linux main.go
GOOS=windows GOARCH=amd64 go build -o alfredo-go.exe main.go
```

## Testing

### Python
- No existing tests in original codebase

### Go
- Structured testing with Go's built-in test framework
- Example: `pkg/config/config_test.go`

## Maintainability

### Python
- Code scattered across multiple files
- Repeated patterns and imports

### Go
- Clear separation of concerns
- Reusable packages
- Consistent patterns
- Better documentation through code structure

This migration provides a more robust, maintainable, and performant solution while preserving all the functionality of the original Python scripts.
