# AlfreDo Go

A Go implementation of the AlfreDo Todoist workflow application.

## Overview

AlfreDo Go is a command-line application that provides a single entry point with multiple commands for managing Todoist tasks. It's a migration from the original Python scripts.

## Installation

1. Make sure you have Go 1.21+ installed
2. Clone or navigate to this directory
3. Install dependencies:
   ```bash
   go mod tidy
   ```
4. Build the application:
   ```bash
   go build -o alfredo-go main.go
   ```

## Configuration

Set your Todoist API token as an environment variable:

```bash
export TOKEN="your_todoist_api_token_here"
# or
export TODOIST_TOKEN="your_todoist_api_token_here"
```

## Usage

### Get Tasks

Get tasks due today:
```bash
./alfredo-go get today
```

Get overdue tasks (including today):
```bash
./alfredo-go get overdue
```

### Complete a Task

Mark a task as completed:
```bash
./alfredo-go complete TASK_ID
```

### Get Statistics

View completion statistics:
```bash
./alfredo-go stats
```

### Help

View available commands:
```bash
./alfredo-go help
```

## Commands

- `get [today|overdue]` - Fetch tasks from Todoist
- `complete [task-id]` - Mark a task as completed
- `stats` - Display completion statistics
- `help` - Show help information

## Output Format

The `get` command outputs JSON in Alfred workflow format, compatible with the original Python scripts.

## Project Structure

```
alfredo-go/
├── main.go                    # Main entry point
├── cmd/                       # CLI commands
│   ├── root.go               # Root command
│   ├── get.go                # Get tasks command
│   ├── complete.go           # Complete task command
│   └── stats.go              # Statistics command
├── internal/                  # Internal packages
│   └── service/
│       └── task_service.go   # Business logic
├── pkg/                       # Public packages
│   ├── config/
│   │   └── config.go         # Configuration management
│   ├── todoist/
│   │   └── client.go         # Todoist API client
│   └── utils/
│       └── log.go            # Logging utilities
├── go.mod                     # Go module file
└── README.md                  # This file
```

## Features

- Single binary with multiple commands
- Parallel API requests for better performance
- Structured error handling
- Clean separation of concerns
- Compatible output format with original Python scripts
- Environment-based configuration
