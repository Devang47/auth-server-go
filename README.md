# Project auth-server-go

This is a simple auth server written in Go. It uses a Postgres database to store user information and JWT tokens for authentication.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

## Routes

- `/` - Home
  @returns `auth-server-go`

- `/health` - Get Health
  @returns `OK`

- `/auth/google` - Google OAuth2 route, starts auth flow
  @returns `{token, account}`

- `/logout` - Logout
  @redirects to `/`

- `/account` - Get current user account from DB
  @returns `Account{}`

- `/secure/account/{id}` - Get account from DB using accountId
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
