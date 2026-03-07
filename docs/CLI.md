# gophant CLI

Usage:

```
gophant make:controller User
gophant make:controller User --resource

gophant make:service BillingService

gophant make:model User

gophant make:migration create_users_table

gophant make:policy UserPolicy

gophant make:job SendWelcomeEmail
gophant make:auth
gophant make:routes
gophant make:bootstrap
gophant make:schedule
gophant schedule:run
gophant schedule:work --interval 5

# runtime
gophant serve
gophant migrate
gophant migrate:rollback [steps]
gophant queue:work
gophant queue:work --once
gophant queue:work --sleep 5
gophant queue:retry --max 100
gophant migrate:fresh
gophant migrate:status
gophant migrate:status --json
gophant cache:clear
```

Files are created under `app/` with basic templates (in your app repo).
Migrations are created under `database/migrations` (in your app repo).
