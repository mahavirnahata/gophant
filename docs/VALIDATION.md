# Validation

## Basic Usage

```go
v := validation.New(c.Request).
    Field("email", validation.Required(), validation.Email()).
    Field("password", validation.Required(), validation.Min(8))

if v.Fails() {
    c.JSON(422, map[string]any{"errors": v.Errors()})
    return
}
```

Errors are human-readable by default:

```json
{
  "errors": {
    "email": ["The email field must be a valid email address."],
    "password": ["The password field must be at least 8 characters."]
  }
}
```

## Validating Non-Request Data

```go
v := validation.NewFromMap(map[string]string{
    "name":  "Alice",
    "email": "not-an-email",
})
v.Field("email", validation.Required(), validation.Email())
```

## Cross-Field Rules (FieldWith)

```go
v.FieldWith("password", validation.Confirmed("password_confirmation"))
v.FieldWith("email",    validation.Unique(conn, "users", "email"))
```

## Custom Error Messages

```go
v := validation.New(c.Request).
    WithMessages(map[string]string{
        "email.required": "We need your email to create an account.",
        "email.email":    "That doesn't look like a valid email.",
        "required":       "The :field field cannot be blank.",
    }).
    Field("email", validation.Required(), validation.Email())
```

## All Built-in Rules

| Rule | Description |
|---|---|
| `Required()` | Value must be non-empty |
| `Email()` | Must be a valid email address |
| `Min(n)` | String length ≥ n characters |
| `Max(n)` | String length ≤ n characters |
| `MinValue(n)` | Numeric value ≥ n |
| `MaxValue(n)` | Numeric value ≤ n |
| `Numeric()` | Must be a number |
| `Alpha()` | Letters only |
| `AlphaNum()` | Letters and digits only |
| `Boolean()` | Must be true/false/1/0/yes/no |
| `In(a, b, ...)` | Value must be one of the listed options |
| `NotIn(a, b, ...)` | Value must NOT be one of the listed options |
| `URL()` | Must be a valid http/https URL |
| `UUID()` | Must be a valid UUID |
| `Regex(re)` | Must match the compiled regexp |
| `Confirmed(field)` | Value must equal another field |
| `Same(field)` | Alias for Confirmed |
| `Unique(db, table, col)` | Value must not exist in the database column |

## Reading Results

```go
v.Fails()           // bool — any errors?
v.Passes()          // bool — inverse
v.Errors()          // map[string][]string — all errors
v.First("email")    // string — first error for a field
v.Value("email")    // string — the submitted value
```
