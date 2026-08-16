# Chirpy HTTP server

* A simple HTTP server and REST api example 
* Using standard GO libraries like database/sql, net/http. For database postgresql.
* Database migrations using goose (https://github.com/pressly/goose)

## ⚙️ Quick Start

Create new .env file:

```env
DB_URL="postgres://{user}:{password}@localhost:{port}/{dbname}?sslmode=disable"
PLATFORM="dev"
JWT_SECRET="{jwtkey}"
POLKA_KEY="{polkakey}"
```

Replace "{...}" with your own credentials and keys. For JWT_SECRET and POLKA_KEY you can generate keys with openssl:

```bash
openssl rand -base64 64
```

Install modules:

```bash
go mod download
```

Build and run:

```bash
go build -o out && ./out
```

## Options and configuring

### New migration

Inside ./sql/schema create new raw .sql file ("006_posts.sql")

File example:

```sql
-- +goose Up
CRATE TABLE posts (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    post TEXT NOT NULL,
    user_id UUID not null REFERENCES users (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts:
```

Migrate: 

```bash
cd ./sql/schema
goose postgres "postgres://{user}:{password}@localhost:{port}/{dbname}" up
```

### Database queries

Inside ./sql/queries create new raw .sql file ("posts.sql")

File example:

```sql
-- name: CreatePost :one
INSERT INTO users (id, created_at, updated_at, post, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetPosts :many
SELECT * FROM posts;

-- name: GetPost :one
SELECT * FROM posts WHERE id=$1;

-- name: UpdatePost :one
UPDATE posts SET post = $1, user_id = $2
WHERE id = $1
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1;
```

Generate methods:

```bash
sqlc generate
```