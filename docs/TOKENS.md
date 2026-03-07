# API Tokens

Gophant includes a simple token service and middleware for API tokens.

## Migration (example)

```sql
CREATE TABLE personal_access_tokens (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  token_hash VARCHAR(64) UNIQUE NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  name VARCHAR(255),
  abilities TEXT,
  expires_at TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Create a token

```go
svc := auth.NewTokenService(app.DB)
plain, _ := svc.Create("1", "api", []string{"read"}, 24*time.Hour)
```

Return `plain` to the client **once**. Only the hash is stored.

## Protect routes

```go
svc := auth.NewTokenService(app.DB)

app.Router.Get("/api/me", handler, auth.Bearer(svc, auth.BearerOptions{}))
```

## Read user + abilities

```go
id, _ := c.Get("auth_user_id")
abilities, _ := c.Get("auth_abilities")
```
