# Code Execution Engine (CEE v2)

A secure, containerized code execution service that allows running code in multiple programming languages through a REST API. The engine uses Docker containers to provide isolated execution environments for Python, Node.js, and Go.

## Features

- **Multi-language Support**: Execute code in Python, Node.js, and Go
- **Isolated Execution**: Each code execution runs in its own Docker container
- **Timeout Control**: Configurable execution timeouts (max 20 seconds)
- **Input Support**: Accept stdin input for interactive programs
- **REST API**: Simple HTTP interface for code execution
- **Container Pooling**: Pre-warmed containers for faster execution

## Supported Languages

- **Python 3.11** (Alpine Linux)
- **Node.js** (Alpine Linux)
- **Go** (Alpine Linux)

## API Documentation

### Execute Code

**Endpoint**: `POST /execute`

**Request Body**:
```json
{
  "language": "python|node|go",
  "code": "print('Hello World')",
  "stdin": "optional input",
  "timeout": 10
}
```

**Parameters**:
- `language` (required): The programming language ("python", "node", or "go")
- `code` (required): The source code to execute
- `stdin` (optional): Input to pass to the program via stdin
- `timeout` (optional): Execution timeout in seconds (1-20, default: 20)

**Response**:
```json
{
  "stdout": "Hello World\n",
  "stderr": "",
  "exitCode": 0,
  "duration": 0.123
}
```

### Health Check

**Endpoint**: `GET /`

Returns: `"CEE Running"`

## Setup Instructions

### Prerequisites

- Go 1.24+
- Docker
- Make (optional, for convenience commands)

### 1. Clone and Build

```bash
git clone <repository-url>
cd ceev2
go mod download
```

### 2. Build Docker Images

Build all language runtime images:

```bash
make build-all
```

Or build individually:

```bash
make build-python
make build-node
make build-go
```

### 3. Run the Service

```bash
go run main.go
```

The server will start on port 3000.

## Usage Examples

### Python Execution

```bash
curl -X POST http://localhost:3000/execute \
  -H "Content-Type: application/json" \
  -d '{
    "language": "python",
    "code": "name = input()\nprint(f\"Hello, {name}!\")",
    "stdin": "World"
  }'
```

### Node.js Execution

```bash
curl -X POST http://localhost:3000/execute \
  -H "Content-Type: application/json" \
  -d '{
    "language": "node",
    "code": "console.log(\"Hello from Node.js!\");"
  }'
```

### Go Execution

```bash
curl -X POST http://localhost:3000/execute \
  -H "Content-Type: application/json" \
  -d '{
    "language": "go",
    "code": "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"Hello from Go!\") }"
  }'
```

## Architecture

The system consists of several components:

- **API Server**: HTTP server built with Fiber framework
- **Executor**: Manages code execution requests
- **Docker Client**: Interfaces with Docker API
- **Pool Manager**: Maintains a pool of pre-warmed containers
- **Sandbox**: Provides isolated execution environment
- **Language Loader**: Loads language configurations from JSON

## Security

- Code execution is isolated in Docker containers
- Non-root user execution in containers
- Automatic cleanup of temporary files
- Configurable execution timeouts
- Input validation and sanitization

## Development

### Running Tests

```bash
make tester
```

### Project Structure

```
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── Makefile               # Build automation
├── docker/                # Docker configurations
│   ├── python/
│   ├── node/
│   └── go/
├── internal/
│   ├── api/              # HTTP API handlers
│   ├── docker/           # Docker client and pooling
│   ├── Executer/         # Code execution logic
│   ├── languages/        # Language configuration
│   ├── sandbox/          # Execution sandbox
│   └── utils/            # Utility functions
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request


