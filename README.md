# PocketJSON 🚀

A lightweight, single-binary JSON storage service with built-in expiry and multi-tenant support. Perfect for developers who need a quick, reliable way to store and retrieve JSON data without the overhead of a full database setup.

## Features ✨

- **Zero Configuration**: Just run the binary and you're ready to go
- **Built-in Expiry**: All stored JSONs automatically expire (configurable)
- **Multi-tenant Support**: Managed API keys with isolated namespaces
- **Guest Mode**: Quick storage without authentication
- **Size Limits**: 100KB for guests, 1MB for authenticated users
- **Docker Ready**: Easy deployment with Docker and docker-compose
- **SQLite Backend**: Simple, reliable, and portable
- **Automatic Cleanup**: Background process removes expired data

## Quick Start 🚀

### Using Docker

1. Copy the `docker-compose.yml` file
2. Run `docker-compose up -d`
3. (optional) Set the `MASTER_API_KEY` env variable to a `.env` file

### Direct Usage

- Download the latest release from the releases page.
- Build it yourself:
  
```bash
# Build
go build
# Run
./pocketjson
```

## API Usage 📡

### Guest Mode

Store JSON (expires in 24 hours):
```bash
curl -X POST http://localhost:9819 \
  -H "Content-Type: application/json" \
  -d '{"hello":"world"}'

# Response:
{
  "id": "f7a8b9c0d1e2",
  "expires_at": "2024-01-21T15:30:45Z"
}
```

Retrieve JSON:
```bash
curl http://localhost:9819/f7a8b9c0d1e2
```

### Authenticated Mode

First, create an API key (requires master key):

```bash
curl -X POST http://localhost:9819/admin/keys \
  -H "X-API-Key: your-master-key" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Development API Key",
    "is_admin": false
  }'

# Response:
{
  "key": "924a98c84222ca4b2984e417c767c519",
  "client_id": "7f3d8",
  "description": "Development API Key",
  "is_admin": false,
  "created_at": "2024-01-20T15:30:45Z"
}
```

Store JSON with custom expiry:

```bash
# Expire after 3 days
curl -X POST "http://localhost:9819/my-data?expiry=72" \
  -H "X-API-Key: 924a98c84222ca4b2984e417c767c519" \
  -H "Content-Type: application/json" \
  -d '{"hello":"world"}'

# Response:
{
  "id": "7f3d8_my-data",
  "expires_at": "2024-01-23T15:30:45Z"
}
```

Note: Authenticated users' IDs are prefixed with their client_id (first 5 chars of MD5(api_key))

## API Reference 📚

### Endpoints

| Method | Path | Description | Auth Required |
|--------|------|-------------|---------------|
| POST | / | Store JSON with random ID | No |
| POST | /{id} | Store JSON with specific ID | Yes |
| GET | /{id} | Retrieve JSON | No |
| POST | /admin/keys | Create API key | Yes (Admin) |
| DELETE | /admin/keys/{key} | Delete API key | Yes (Admin) |
| GET | /health | Health check | No |

### Storage Limits

- Guest users: 100KB
- Authenticated users: 1MB

### Expiry Options

- Guest users: 48 hours
- Authenticated users can specify:
  - Custom hours: `?expiry=72` (72 hours)
  - Never expire: `?expiry=never`

## Development 🛠

### Prerequisites

- Go 1.23+
- SQLite3

### Setup

```bash
# Generate Ent code
go generate ./ent

# Build
go build
```

## Why PocketJSON? 🤔

PocketJSON was created to solve the common need for temporary JSON storage without the complexity of setting up and maintaining a full database system. It's perfect for:

- Development and testing
- Temporary data storage
- Webhook payload storage
- API response caching
- Cross-service data sharing

## License 📄

MIT License - see [LICENSE](LICENSE) for details

## Contributing 🤝

Contributions are welcome! Please feel free to submit a Pull Request.