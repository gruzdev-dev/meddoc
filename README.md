# MedDoc

Medical documents manager project for SPbSTU R&D.

## Project Structure

```
.
├── api/                    # API specifications and OpenAPI docs
├── cmd/                    # Application entry points
│   └── main.go            # Main application entry point
├── configs/               # Configuration files
├── docs/                  # Project documentation
├── internal/              # Internal code
│   ├── app/              # Application layer (services, handlers)
│   ├── database/         # Database layer (MongoDB implementation)
│   └── domain/           # Domain layer (entities, repositories)
├── pkg/                   # Public packages
└── vendor/               # Vendored dependencies
```

## Features

- Graceful shutdown
- Structured logging
- Request ID tracking
- Response compression
- Security headers
- MongoDB connection retry
- API versioning
- OpenAPI documentation
- Dependency injection
- Clean architecture

## Requirements

- Go 1.21+
- MongoDB 6.0+
- Docker (optional, for development)

## Setup

1. Clone and install:
```bash
git clone https://github.com/gruzdev-dev/meddoc.git
cd meddoc
go mod download
```

2. Run MongoDB (using Docker):
```bash
docker run -d -p 27017:27017 --name mongodb mongo:latest
```

3. Start application:
```bash
go run cmd/main.go
```

## Development

The project follows clean architecture principles with clear separation of concerns:
- Domain layer contains business logic and interfaces
- Application layer implements use cases and services
- Database layer handles data persistence
- API layer defines external interfaces

## License

MIT
