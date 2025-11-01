# Database Migrations

This directory contains versioned database migrations for the Core Banking Lab PostgreSQL database.

## Migration Tool

We use [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations.

## Migration Files

Migrations follow the naming convention: `{version}_{name}.{up|down}.sql`

- `000001_init_schema.up.sql` - Creates initial schema (accounts, transactions tables)
- `000001_init_schema.down.sql` - Rolls back initial schema

## Running Migrations Programmatically

Migrations are automatically applied when the application starts. See `internal/infrastructure/database/postgres/postgres.go` for implementation.

## Manual Migration (Development)

### Using migrate CLI

Install the CLI tool:
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Run migrations:
```bash
migrate -path internal/infrastructure/database/postgres/migrations \
        -database "postgres://banking:banking_secure_pass_2024@localhost:5432/banking?sslmode=disable" \
        up
```

Rollback last migration:
```bash
migrate -path internal/infrastructure/database/postgres/migrations \
        -database "postgres://banking:banking_secure_pass_2024@localhost:5432/banking?sslmode=disable" \
        down 1
```

Check migration version:
```bash
migrate -path internal/infrastructure/database/postgres/migrations \
        -database "postgres://banking:banking_secure_pass_2024@localhost:5432/banking?sslmode=disable" \
        version
```

## Docker Compose

For Docker Compose deployments, the schema is initialized via SQL scripts in:
`deployments/docker-compose/postgres/init/01-schema.sql`

This approach is simpler for local development and ensures a clean database state on each `docker-compose up`.

## Creating New Migrations

Generate a new migration:
```bash
migrate create -ext sql -dir internal/infrastructure/database/postgres/migrations -seq add_new_feature
```

This creates:
- `{version}_add_new_feature.up.sql`
- `{version}_add_new_feature.down.sql`

## Best Practices

1. **Always create both up and down migrations** - Ensure rollback capability
2. **Test migrations** - Run up and down to verify they work correctly
3. **Idempotent migrations** - Use `IF EXISTS` and `IF NOT EXISTS` where appropriate
4. **Avoid data loss** - Be careful with DROP statements in production
5. **Version control** - Never modify existing migration files, create new ones
6. **Sequential numbering** - Migrations are applied in order by version number
