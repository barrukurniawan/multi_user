# Task Management API (Multi-User)

Created by Barru Kurniawan

## Tech Stack :
- Language : Golang 1.23
- Framework : Gin
- Database : PostgreSQL
- ORM : GORM

## Migration:
### Create database
#### Via Terminal
```
createdb user_multi
```
#### Via PSQL
```
psql -d postgres -c "CREATE DATABASE user_multi;"
```

### Run migration
```
psql -d user_multi -f migrations/schema.sql
```
