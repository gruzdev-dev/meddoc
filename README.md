# MedDoc

A medical document management system built with Go, providing secure document storage and user management capabilities.

## Features

- User authentication and authorization
- Document management and storage
- RESTful API
- MongoDB integration
- Docker support

## Prerequisites

- Go 1.24.1 or higher
- Docker and Docker Compose
- MongoDB (included in Docker setup)

## Quick Start

1. Clone the repository:
```bash
git clone https://github.com/gruzdev-dev/meddoc.git
cd meddoc
```

2. Start the application using Docker Compose:
```bash
docker compose up -d
```

The application will be available at `http://localhost:8080`

## Configuration

The application is configured via `config.yaml`. Key configuration options:

- Server settings (port, host)
- MongoDB connection
- Logging configuration
- Authentication settings

## Project Structure

```
.
├── api/            # API definitions and specifications
├── app/            # Application core
│   ├── config/     # Configuration management
│   ├── handlers/   # HTTP handlers
│   ├── repositories/ # Data access layer
│   ├── server/     # HTTP server setup
│   └── services/   # Business logic
├── database/       # Database connection and utilities
├── pkg/            # Shared packages
└── vendor/         # Dependencies
```

## Development

1. Install dependencies:
```bash
go mod download
```

2. Run the application locally:
```bash
go run main.go
```

## API Documentation

The API is documented using OpenAPI 3.0 specification in `api/openapi.yaml`. The specification includes:

- Authentication endpoints (register, login, refresh token)
- Document management endpoints (CRUD operations)
- Request/response schemas
- Security requirements

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
