# Expense Tracker API

A lightweight expense tracker REST API built with Go and [Beego](https://beego.me/). This service uses CSV files for persistence and exposes an interactive Swagger UI for API exploration.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation & Setup](#installation--setup)
- [Configuration](#configuration)
- [Running the Server](#running-the-server)
- [Running Tests](#running-tests)
- [Swagger Documentation](#swagger-documentation)
- [API Reference](#api-reference)
- [Folder Structure](#folder-structure)
- [Notes](#notes)

---

## Overview

This API supports:

- User registration and login
- Expense creation, listing, updating, and deletion
- Expense summary reporting
- Filter, sort, and paginate expense results
- CSV-backed storage for users and expenses

---

## Features

| Method | Endpoint                | Description         |
| ------ | ----------------------- | ------------------- |
| GET    | `/api/v1/health`        | Health check        |
| POST   | `/api/v1/auth/register` | Register a new user |
| POST   | `/api/v1/auth/login`    | Login               |
| POST   | `/api/v1/expenses`      | Create an expense   |
| GET    | `/api/v1/expenses`      | List expenses       |
| GET    | `/api/v1/expenses/{id}` | Get expense by ID   |
| PUT    | `/api/v1/expenses/{id}` | Update an expense   |
| DELETE | `/api/v1/expenses/{id}` | Delete an expense   |
| GET    | `/api/v1/summary`       | Expense summary     |

---

## Prerequisites

- **Go 1.21** or later
- **Beego `bee` CLI** — used to run the server, hot-reload, and generate Swagger docs

---

## Installation & Setup

**1. Clone the repository**

```bash
git clone https://github.com/sabbirhosen44/Expense-Tracker-API.git
cd Expense-Tracker-API
```

**2. Install Go dependencies**

```bash
go mod tidy
```

**3. Install the Beego `bee` CLI**

```bash
go install github.com/beego/bee/v2@latest
```

Make sure `$GOPATH/bin` (or `$HOME/go/bin`) is in your `PATH`:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Verify the installation:

```bash
bee version
```

**4. Set up configuration**

Copy the example config and update values as needed:

```bash
cp conf/app.conf.example conf/app.conf
```

Key values in `conf/app.conf`:

```ini
httpport     = 8080
runmode      = dev
users_csv    = data/users.csv
expenses_csv = data/expenses.csv
```

**5. Generate Swagger docs**

This step is required before the first run:

```bash
bee generate docs
```

---

## Running the Server

Start the server with hot-reload using `bee run`:

```bash
bee run
```

The server starts on **http://localhost:8080** by default. `bee run` watches for file changes and automatically restarts the server.

To run without hot-reload:

```bash
bee run -runmode=prod
```

---

## Running Tests

Run all tests across the project:

```bash
bee test ./...
```

Or use the standard Go test runner:

```bash
go test ./...
```

Run tests for a specific package:

```bash
# Controllers
go test ./controllers/...

# Models
go test ./models/...

# Middlewares
go test ./middlewares/...

# Routers
go test ./routers/...
```

Run tests with verbose output:

```bash
go test -v ./...
```

Run tests with coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Swagger Documentation

Interactive API docs are available at:

```
http://localhost:8080/swagger
```

Raw spec files:

```
http://localhost:8080/swagger/swagger.json
http://localhost:8080/swagger/swagger.yml
```

### Generating Swagger docs

Before the first run, generate the docs once:

```bash
bee generate docs
```

### Regenerating after controller changes

Swagger docs are generated from annotations in `controllers/*.go`. After modifying or adding any API comments, regenerate and restart:

```bash
bee generate docs
bee run
```

> **Note:** Docs are **not** updated automatically at runtime — you must re-run `bee generate docs` and restart the server for changes to take effect.

---

## API Reference

### Authorization

Protected endpoints require the following request header:

```
X-User-ID: <numeric-user-id>
```

### Authentication

**Register**

```
POST /api/v1/auth/register
Body: { "name": "", "email": "", "password": "" }
```

**Login**

```
POST /api/v1/auth/login
Body: { "email": "", "password": "" }
```

### Expenses

**Create**

```
POST /api/v1/expenses
Body: { "title": "", "amount": 0, "category": "", "note": "", "expense_date": "" }
```

**List** (supports filtering, sorting, pagination)

```
GET /api/v1/expenses
Query: category, date_from, date_to, sort_by, sort_order, limit, offset
```

**Get by ID**

```
GET /api/v1/expenses/{id}
```

**Update**

```
PUT /api/v1/expenses/{id}
Body: { "title": "", "amount": 0, "category": "", "note": "", "expense_date": "" }
```

**Delete**

```
DELETE /api/v1/expenses/{id}
```

**Summary**

```
GET /api/v1/summary
Query: date_from, date_to
```

---

## Folder Structure

```text
Expense-Tracker-API/
├── conf/
│   ├── app.conf
│   └── app.conf.example
├── controllers/
│   ├── auth.go
│   ├── auth_test.go
│   ├── base.go
│   ├── expense.go
│   ├── expense_test.go
│   ├── health.go
│   └── health_test.go
├── data/
│   ├── expenses.csv
│   └── users.csv
├── middlewares/
│   ├── auth.go
│   └── auth_test.go
├── models/
│   ├── expense.go
│   ├── expense_test.go
│   ├── user.go
│   └── user_test.go
├── routers/
│   ├── router.go
│   └── router_test.go
├── swagger/
│   ├── index.html
│   ├── swagger.json
│   └── swagger.yml
├── tests/
├── go.mod
├── go.sum
├── main.go
└── README.md
```

---

## Notes

- Authentication uses a simple `X-User-ID` header. For production, replace with token-based auth (e.g. JWT).
- CSV files are created automatically under `data/` on first use.
- Swagger UI is served from the `/swagger` route and requires `bee generate docs` to stay up to date.
- For production, replace CSV storage with a proper database.
