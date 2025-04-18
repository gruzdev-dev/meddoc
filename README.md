# MedDoc

Medical documents manager project for SPbSTU R&D.

## Project Structure

```
.
├── api/                    # API specifications
├── cmd/                    # Application entry points
├── configs/                # Configuration files
├── internal/              # Internal code
│   ├── domain/           # Domain layer
│   └── app/              # Application layer
└── pkg/                  # Public packages
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

## Requirements

- Go 1.21+
- MongoDB 6.0+

## Setup

1. Clone and install:
```bash
git clone https://github.com/gruzdev-dev/meddoc.git
cd meddoc
go mod download
```

2. Run MongoDB:
```bash
docker run -d -p 27017:27017 --name mongodb mongo:latest
```

3. Start app:
```bash
go run cmd/main.go
```

## License

MIT
