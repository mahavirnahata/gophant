# Policies and Gates

Define abilities:

```
app := gophant.New()
app.Gate.Define("view-admin", func(c *gophanthttp.Context) bool {
	return app.Auth.HasRole(c, "admin")
})
```

Protect routes:

```
app.Router.Get("/admin", handler, app.Gate.Require("view-admin"))
```

Role middleware:

```
app.Router.Get("/billing", handler, app.Auth.RequireRole("billing"))
app.Router.Get("/staff", handler, app.Auth.RequireAnyRole("admin", "staff"))
```

Set roles on login:

```
app.Auth.SetRoles(c, []string{"admin"})
```

Policy scaffolding:

```
go run cmd/gophant/main.go make:policy UserPolicy
```

Auto-register a policy:

```
reg := auth.NewPolicyRegistrar(app.Gate)
reg.Register("user", &policies.UserPolicy{})

// abilities: user.viewAny, user.view, user.create, user.update, user.delete
```
