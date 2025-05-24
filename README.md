# MedDoc

Medical document management service with REST API.

## Technologies

- Go 1.24.1
- MongoDB
- Docker & Docker Compose
- JWT Authentication

## Project Structure

```
.
├── api/            # API specifications and documentation
├── app/           # Application core
│   ├── errors/    # Error definitions and handling
│   ├── handlers/  # HTTP handlers
│   ├── models/    # Data models
│   ├── server/    # HTTP server
│   └── services/  # Business logic
├── config/        # Application configuration
├── database/      # Database layer
├── pkg/           # Shared utilities
└── tests/         # Integration tests
```

## Quick Start

```bash
git clone https://github.com/gruzdev-dev/meddoc.git
cd meddoc
docker compose up -d
```

Application will be available at: http://localhost:8080

## API Documentation

OpenAPI 3.0 specification in `api/` directory.

## Development

```bash
go mod download
docker compose up -d mongodb
go run main.go
```

## License

MIT License - see [LICENSE](LICENSE) file for details.
