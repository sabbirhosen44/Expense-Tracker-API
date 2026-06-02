# Expense Tracker API

A lightweight expense tracker REST API built with Go and [Beego](https://beego.me/). This service uses CSV files for persistence and exposes an interactive Swagger UI for API exploration.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Swagger UI](#swagger-ui)
- [API Reference](#api-reference)
- [Folder Structure](#folder-structure)
- [Notes](#notes)

## Overview

This API supports:

- user registration and login
- expense creation, listing, updating, deletion
- expense summary reporting
- filter, sort, and paginate expense results
- CSV-backed storage for users and expenses

## Features

- `GET /api/v1/health`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/expenses`
- `GET /api/v1/expenses`
- `GET /api/v1/expenses/{id}`
- `PUT /api/v1/expenses/{id}`
- `DELETE /api/v1/expenses/{id}`
- `GET /api/v1/summary`

## Prerequisites

- Go 1.26 or later
- `bee` CLI

## Quick Start

Install dependencies:

```bash
go mod tidy
```

Install the Beego CLI if you do not have it already:

```bash
go install github.com/beego/bee/v2@latest
```

Generate Swagger docs once before the first run:

```bash
bee generate docs
```

Run the service:

```bash
go run main.go
```

The server listens on port `8080` by default.

## Swagger UI

Application settings are stored in `conf/app.conf`.

Important values:

- `httpport = 8080`
- `runmode = dev`
- `users_csv = data/users.csv`
- `expenses_csv = data/expenses.csv`

## Swagger UI

Interactive API docs are served from:

- `http://localhost:8080/swagger`

Raw spec files are available at:

- `http://localhost:8080/swagger/swagger.json`
- `http://localhost:8080/swagger/swagger.yml`

### Regenerating docs after controller changes

Swagger docs are generated from controller annotations. When you change or add API comments in `controllers/*.go`, run:

```bash
bee generate docs
```

This command updates the generated spec files in `swagger/swagger.json` and `swagger/swagger.yml`.
These docs are not updated automatically at runtime.

## API Reference

### Authorization

Protected endpoints require the request header:

- `X-User-ID`: numeric user ID

### Endpoints

#### Health

- `GET /api/v1/health`

#### Authentication

- `POST /api/v1/auth/register`

  - Body: `name`, `email`, `password`

- `POST /api/v1/auth/login`
  - Body: `email`, `password`

#### Expenses

- `POST /api/v1/expenses`

  - Body: `title`, `amount`, `category`, `note`, `expense_date`

- `GET /api/v1/expenses`

  - Query: `category`, `date_from`, `date_to`, `sort_by`, `sort_order`, `limit`, `offset`

- `GET /api/v1/expenses/{id}`

- `PUT /api/v1/expenses/{id}`

  - Body: `title`, `amount`, `category`, `note`, `expense_date`

- `DELETE /api/v1/expenses/{id}`

- `GET /api/v1/summary`
  - Query: `date_from`, `date_to`

## Folder Structure

```text
Expense-Tracker-API/
├── conf/
│   ├── app.conf
│   └── app.conf.example
├── controllers/
│   ├── auth.go
│   ├── base.go
│   ├── expense.go
│   └── health.go
├── data/
│   ├── expenses.csv
│   └── users.csv
├── middlewares/
│   └── auth.go
├── models/
│   ├── expense.go
│   └── user.go
├── routers/
│   └── router.go
├── swagger/
│   ├── index.html
│   ├── swagger.json
│   └── swagger.yml
├── go.mod
├── main.go
└── README.md
```

## Notes

- This implementation uses a simple `X-User-ID` header for authentication.
- CSV files are created automatically in `data/` when needed.
- Swagger UI is served from the `/swagger` route.
- For a production-ready API, replace header auth with token-based authentication and persistent storage.

## New Developer Setup

After cloning the repository, follow these steps:

```bash
go mod tidy
go install github.com/beego/bee/v2@latest
bee generate docs
go run main.go
```

Then open `http://localhost:8080/swagger` in your browser.
