# api-server
test 4
### Introduction

This repository contains the API server for handling user interactions. It is built using Go and utilizes PostgreSQL as the relational database. Database migrations are managed using Flyway.

### Prerequisites

- **Go**: Ensure you have Go installed on your system.
- **PostgreSQL**: A PostgreSQL database is required for storing data.
- **Flyway**: Used for managing database schema migrations.
- **Docker**: Optional for running the application in a containerized environment.


### Build and Deploy Instructions

1. **Build the Application**:
   ```bash
   go build -o api-server cmd/main.go
   ```

2. **Run the Application**:
   ```bash
   ./api-server
   ```

3. **Run with Docker**:
   ```bash
   docker build -t api-server .
   docker run -p 8080:8080 api-server
   ```

4. **Database Migrations**:
   Use Flyway to manage database schema migrations. Ensure you have a `.env` file with your database credentials.

   ```bash
   docker run --rm --network=host -v $PWD/sql:/flyway/sql flyway/flyway -url=jdbc:postgresql://localhost:5432/mydb -user=myuser -password=mypassword migrate
   ```

### API Endpoints

API endpoints are documented using Swagger. You can find the API documentation at [SwaggerHub](https://app.swaggerhub.com/apis-docs/csye7125-fall2023/csye7125-spring2025-api-server/2025.05.01).

---

### Folder Structure

The project follows a layered approach for organization:

```plaintext
api-server/
├── cmd/
│   └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   └── user.go
│   │   └── routes.go
│   ├── model/
│   │   └── user.go
│   ├── repository/
│   │   └── user_repository.go
│   └── service/
│       └── user_service.go
├── pkg/
│   └── db/
│       └── postgres.go
├── sql/
│   └── migrations/
│       ├── V1.0__init.sql
│       └── V1.1__add_user_table.sql
├── .env
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

