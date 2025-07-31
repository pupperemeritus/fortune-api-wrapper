# Fortune API

A production-ready REST API wrapper for the Unix `fortune` command, built with Go.

## Features

- **Full Fortune CLI Support**: All command-line options exposed as API parameters
- **RESTful API**: Clean HTTP endpoints with JSON responses
- **Production Ready**: Structured logging, graceful shutdown, health checks
- **Docker Support**: Multi-stage builds with health checks
- **Configurable**: Environment-based configuration
- **Error Handling**: Comprehensive error handling and logging
- **CORS Support**: Cross-origin request support

## API Endpoints

### Get Fortune

```
GET /fortune
```

Query Parameters:

- `all` (bool): Choose from all lists of maxims
- `show_cookie` (bool): Show the source file
- `equal` (bool): Consider all files equal size
- `long` (bool): Long fortunes only
- `short` (bool): Short fortunes only
- `ignore_case` (bool): Ignore case for pattern matching
- `wait` (bool): Wait before termination
- `length` (int): Maximum length for "short" fortunes
- `pattern` (string): Search pattern
- `files` (string): Comma-separated list of files
- `percentages` (string): Comma-separated list of percentages

### List Available Files

```
GET /fortune/files
```

### Search Fortunes

```
GET /fortune/search?pattern=wisdom
```

### Health Check

```
GET /health
```

## Quick Start

### Using Docker

```bash
docker build -t fortune-api .
docker run -p 8080:8080 fortune-api
```

### Using Docker Compose

```bash
docker-compose up
```

### Local Development

```bash
# Install dependencies
go mod tidy

# Run locally
make run

# Or directly
go run .
```

## Configuration

Environment variables:

- `SERVER_ADDRESS`: Server bind address (default: `:8080`)
- `FORTUNE_PATH`: Path to fortune binary (default: `fortune`)
- `READ_TIMEOUT`: HTTP read timeout (default: `15s`)
- `WRITE_TIMEOUT`: HTTP write timeout (default: `15s`)
- `IDLE_TIMEOUT`: HTTP idle timeout (default: `60s`)

## Examples

```bash
# Get a random fortune
curl http://localhost:8080/fortune

# Get a long fortune
curl "http://localhost:8080/fortune?long=true"

# Search for fortunes containing "wisdom"
curl "http://localhost:8080/fortune/search?pattern=wisdom"

# List available fortune files
curl http://localhost:8080/fortune/files

# Health check
curl http://localhost:8080/health
```

## Development

```bash
# Format code
make fmt

# Run tests
make test

# Build binary
make build

# Clean build artifacts
make clean
```

## Project Structure

```
fortune-api/
├── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go      # Configuration management
│   ├── handlers/
│   │   ├── handlers.go    # HTTP handlers
│   │   └── middleware.go  # HTTP middleware
│   └── service/
│       └── fortune.go     # Fortune service logic
├── Dockerfile             # Multi-stage Docker build
├── docker-compose.yml     # Docker Compose configuration
├── Makefile              # Build and development tasks
├── go.mod                # Go module definition
└── README.md             # This file
```

## License

This project is open source and available under the GPL V3 License.
