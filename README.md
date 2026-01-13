# WellPaw Backend

A comprehensive pet health management backend API built with Go, providing authentication, pet management, and health tracking services.

## 📋 Table of Contents

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Docker Management](#docker-management)
- [Security Setup](#security-setup)
- [API Documentation](#api-documentation)
- [Project Structure](#project-structure)

## 🎯 Overview

WellPaw Backend is a RESTful API service designed for pet health management. It provides secure authentication with Google OAuth support, pet profile management, and comprehensive health tracking capabilities.

## 🚀 Tech Stack

### Core Framework
- **Go 1.25.4** - Primary programming language
- **Fiber v2** - Fast and lightweight web framework for building APIs

### Database & Caching
- **PostgreSQL 17** (Alpine) - Primary relational database
- **GORM** - ORM library for database operations
- **Redis 8.0.2** (Alpine) - In-memory caching and session management

### Authentication & Security
- **JWT (JSON Web Tokens)** - Secure token-based authentication
- **OAuth2** - Google OAuth integration
- **bcrypt** - Password hashing
- **RSA Key Pair** - Public/private key encryption

### API Documentation
- **Swagger/OpenAPI** - Interactive API documentation
- **Swaggo** - Automatic Swagger generation from Go annotations

### DevOps
- **Docker & Docker Compose** - Containerization and orchestration
- **Air** - Hot-reload for Go applications during development

## 📦 Prerequisites

Before starting this project, ensure you have the following installed:

- **Go** 1.25.4 or higher - [Download](https://golang.org/dl/)
- **Docker** - [Download](https://www.docker.com/get-started)
- **Docker Compose** - Usually bundled with Docker Desktop
- **OpenSSL** - For generating RSA key pairs
- **Make** - Build automation tool (pre-installed on macOS/Linux)
- **Swag CLI** - For generating Swagger documentation
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```

## 🏃 Getting Started

### 1. Clone the Repository
```bash
git clone <repository-url>
cd wellpaw-backend
```

### 2. Environment Setup
Create a `.env` file based on the example:
```bash
cp .env.example .env
```

Edit the `.env` file with your configuration:
```env
# Database Configuration
DB_HOST=pethealth_db
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=pethealth

# Redis Configuration
REDIS_HOST=pethealth_redis
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password

# JWT Configuration
JWT_SECRET=your_jwt_secret

# Google OAuth
GOOGLE_WEB_CLIENT_ID=your_google_client_id
GOOGLE_WEB_CLIENT_SECRET=your_google_client_secret

# Google Vision API, Google Cloud Storage, Firebase
GOOGLE_APPLICATION_CREDENTIALS=your_google_application_credentials

# Gemini API from goole ai studio
GEMINI_API_KEY=your_gemini_api_key

# Stirpe API Key
STRIPE_API_KEY=your_stripe_api_key
```

### 3. Generate RSA Key Pair
Generate private and public keys for JWT signing:

```bash
# For development environment
make keypair ENV=dev

# For production environment
make keypair ENV=prod
```


**Manual generation:**
```bash
# Generate private key (2048-bit RSA)
openssl genpkey -algorithm RSA -out private_key.pem -pkeyopt rsa_keygen_bits:2048

# Extract public key from private key
openssl rsa -in private_key.pem -out public_key.pem -pubout
```

This will create:
- `private_key_<ENV>.pem` - Private key for signing tokens
- `public_key_<ENV>.pem` - Public key for verifying tokens

### 4. Start the Application
```bash
make up
```

The application will be available at:
- **API Server**: http://localhost:50001
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379
- **Swagger UI**: http://localhost:50001/swagger/index.html

## 🐳 Docker Management

### Start Containers
Start all services in detached mode:
```bash
make up
```
or
```bash
docker-compose up -d
```

### Stop Containers
Stop all running containers:
```bash
make down
```
or
```bash
docker-compose down
```

### Delete Containers
Remove containers while keeping volumes:
```bash
docker-compose down
```

Remove containers and volumes (⚠️ WARNING: This deletes all data):
```bash
docker-compose down -v
```

## 📚 API Documentation

### Swagger UI
After starting the application, access interactive API documentation at:
```
http://localhost:50001/swagger/index.html
```

### Regenerate Documentation
After modifying API endpoints or annotations:
```bash
make swagger
```

## 📁 Project Structure

```
wellpaw-backend/
├── cmd/
│   └── main/
│       └── main.go              # Application entry point
├── internal/
│   ├── controller/              # HTTP request handlers
│   ├── model/                   # Database models
│   ├── service/                 # Business logic
│   ├── repository/              # Data access layer
│   ├── middleware/              # HTTP middlewares
│   └── config/                  # Configuration management
├── doc/                         # Swagger documentation
├── public/                      # Static files
├── docker-compose.yml           # Docker services configuration
├── Dockerfile                   # Application container definition
├── Makefile                     # Build automation
├── .env.example                 # Environment variables template
└── README.md                    # Project documentation
```

## 🛠️ Makefile Commands

| Command | Description |
|---------|-------------|
| `make keypair ENV=<env>` | Generate RSA key pair for specified environment |
| `make swagger` | Generate Swagger API documentation |
| `make up` | Start all Docker containers |
| `make down` | Stop all Docker containers |
