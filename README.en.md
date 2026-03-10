# Taskery REST API

[Версия на русском языке доступна здесь](./README.md)

REST API for **Taskery** — a task and todo management system with synchronization designed to be used by a CLI client.

## Purpose

This project is created to practice:
- Designing a backend application using DDD principles
- Separating domain logic from application and infrastructure layers
- Building a clean and maintainable REST API

## Architecture

The application is structured into layers:

- **Domain layer**  
  Core business logic: entities, value objects, domain rules

- **Application layer**  
  Use cases and services that orchestrate domain logic

- **Infrastructure layer**  
  Data access and external integrations (e.g. JWT provider)

- **Transport layer**  
  REST API (HTTP handlers, DTOs)

## Technologies

- Go 1.26+
- REST API
- PostgreSQL
- JWT Authorization
- Docker

## Getting Started

Follow these steps to set up and run the project locally.

### Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Docker and Docker Compose](https://www.docker.com/products/docker-desktop)
- PostgreSQL client (optional for debugging)

### Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/godlevsk1y/taskery-api.git
   cd taskery-api
   ```

2. **Configure environment variables**:
   ```bash
   cp example.env .env
   ```
   Edit `.env` with your database credentials and other settings.

3. **Start database and application services in Docker**:
   ```bash
   docker-compose up -d --build
   ```

4. **Build the application (on local machine)**:
   ```bash
   go build -o taskery-api ./cmd/taskery-api
   ```

5. **Run the API server (on local machine)**:
   ```bash
   ./taskery-api
   ```

The server will start at `http://localhost:8080`.

### API Documentation

Interactive API documentation is available at:
```
http://localhost:8080/swagger/index.html
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## Testing

Run all tests with:
```bash
go test ./...
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
