# Rate Limited API Server

This repository contains a simple Go application that serves rate-limited APIs using Gorilla Mux.

## Overview

The application provides the following features:

- CRUD operations for managing clients (not exposed to the public)
- Rate limiting on specific endpoints
- Authentication middleware for validating tokens

## Dependencies

- [Gorilla Mux](https://github.com/gorilla/mux): A powerful URL router and dispatcher for Go.
- [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate): Rate limiting package for Go.

## Installation

1. Clone the repository:
git clone https://github.com/gokul1101/api-rate-limiting.git

2. Install dependencies:
go mod tidy

## Usage

To start the server, run:
go run .


The server will start listening on port `8080`.

### Endpoints

- `GET /clients`: Retrieve the list of clients.
- `POST /client`: Create a new client.
- `PATCH /client`: Update an existing client.
- `DELETE /client`: Delete a client.
- `POST /seed-client`: Seed client data.
- `GET /api/resource1`: Access resource 1 (rate-limited).
- `GET /api/resource2`: Access resource 2 (rate-limited).
- `GET /api/resource3`: Access resource 3 (rate-limited).

## Middleware

The application uses middleware for rate limiting and authentication.

### Rate Limiting Middleware

The `rateLimitMiddleware` limits the number of requests per client IP for specific endpoints. It uses a token bucket algorithm provided by the `golang.org/x/time/rate` package.

### Authentication Middleware

The `authMiddleware` checks for the presence of an Authorization header in incoming requests. Currently, it only checks for a specific token ("Zocket") as a placeholder for authentication.

## Client Management

The `client.go` file contains CRUD operations for managing clients.

### Endpoints

- `POST /client`: Create a new client.
- `GET /clients`: Retrieve the list of clients.
- `PATCH /client`: Update an existing client.
- `DELETE /client`: Delete a client.
- `POST /seed-client`: Seed client data.

## Testing

The `rate_limiter_test.go` file contains unit tests for rate limiting functionality.

## Docker

The provided Dockerfile allows you to containerize the application.

### Build Docker Image
docker build -t rate-limiter:latest .

### Run Docker Container
docker run -p 8080:8080 rate-limiter:latest

## License

This project is licensed under the [MIT License](LICENSE).

