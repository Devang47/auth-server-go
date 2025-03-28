# Project auth-server-go

This is a simple auth server written in Go. It uses a Postgres database to store user information and JWT tokens for authentication.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

## Features

- Password based authentication
- Google OAuth2 authentication
- JWT token generation, validation and refresh
- User account creation and management
- Rate limiting
- Docker support for Postgres database

## Routes

- GET `/` - Home
  @returns `auth-server-go`

- GET `/health` - Get Health
  @returns `OK`

- GET `/refresh-token` - Refresh token
  @header `Authorization: Bearer {token}`
  @returns `{token}`

- GET `/auth/google` - Google OAuth2 route, starts auth flow
  @returns `{token, account}`

- POST `/auth/register` - Register a new user
  @body `{email, password}`
  @returns `{token, account}`

- POST `/auth/login` - Login with email and password
  @body `{email, password}`
  @returns `{token, account}`

- GET `/logout` - Logout
  @redirects to `/`

- GET `/secure/account` - Get current user account from DB
  @header `Authorization: Bearer {token}`
  @returns `Account{}`

- GET `/secure/account/{id}` - Get account from DB using accountId
  @header `Authorization: Bearer {token}`
  @returns `Account{}`

## MakeFile

build the application

```bash
make build
```

run the application

```bash
make run
```

Create DB container

```bash
make docker-run
```

Shutdown DB container

```bash
make docker-down
```

live reload the application

```bash
make watch
```

run the test suite

```bash
make test
```

clean up binary from the last build

```bash
make clean
```
