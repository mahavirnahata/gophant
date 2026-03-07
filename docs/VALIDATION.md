# Validation

Basic rules:

```
v := validation.New(r).
	Field("email", validation.Required(), validation.Email())
```

Context rules:

```
v.FieldWith("password", validation.Confirmed("password_confirmation"))
```

Other rules:

```
validation.Numeric()
validation.Alpha()
validation.In("draft", "published")
validation.Regex(regexp.MustCompile("^foo"))
```

Unique rule:

```
conn, _ := db.Open("mysql", dsn, db.QuestionDialect{})

v.FieldWith("email", validation.Unique(conn, "users", "email"))
```
