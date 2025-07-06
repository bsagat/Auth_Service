# Auth Service

Simple **JWT Cookie-based Authentication Service** for user registration, login, token refreshing, role checks, and admin-level user management.

---

## Features

✅ User registration and login using **email + password**  
✅ **JWT access/refresh tokens** stored in HTTP-only cookies  
✅ Refresh token endpoint for seamless session renewal  
✅ Role check endpoint to verify `IsAdmin`  
✅ Admin-only endpoints:
- View user data (including hashed password)
- Update user name
- Delete user

---

## API Overview

- **Base URL:** `http://localhost:80`
- **Secured URL:** `https://localhost:80`
- **Spec:** [OpenAPI 3.0.3](https://swagger.io/specification/)

### Key Endpoints

| Method | Endpoint       | Description                              |
|--------|----------------|------------------------------------------|
| POST   | `/login`       | User login, returns JWT in cookies      |
| POST   | `/register`    | Register new user                       |
| POST   | `/refresh`     | Refresh JWT using refresh token cookie  |
| GET    | `/role`        | Check user role (`IsAdmin`)             |
| GET    | `/user/{id}`   | Get user data (Admin only)              |
| PUT    | `/user/{id}`   | Update user name (Admin only)           |
| DELETE | `/user/{id}`   | Delete user (Admin only)                |
| GET    | `/swagger/`    | Interactive API documentation           |
---

## Setup

### 1️⃣ Configure PostgreSQL Database

Create a PostgreSQL database matching your `.env` file settings before starting the service.

---

### 2️⃣ Create Configuration File

Create a `.env` file in the project root with:

```env
# Application server settings
PORT=80
HOST=localhost
ENV=dev

# Admin registration
ADMIN_NAME=BekaBratan
ADMIN_PASSWORD=SuperPassword
ADMIN_EMAIL=sagatbekbolat854@gmail.com

# Token settings
ACCESSTTL=15m
REFRESHTTL=168h # (7 days * 24 hours)
SECRET=exampleSecret

# Database configuration
DB_NAME=authDB
DB_USER=Bacoonti
DB_PASSWORD=SuperSecretPassword
DB_PORT=5432
```
