# Migrations

Create a migration:

```
go run cmd/gophant/main.go make:migration create_users_table
```

This creates `database/migrations/<timestamp>_create_users_table.go`.

Register migrations in your app (example):

```
import (
	"github.com/mahavirnahata/gophant"
	"github.com/mahavirnahata/gophant/db/migrate"
)

func init() {
	gophant.RegisterMigrations([]migrate.Migration{
		// your migrations
	}...)
}
```

Run migrations from your app:

```
m := migrate.Migrator{DB: conn}
_ = m.Up(gophant.Migrations())
```

Rollback last batch:

```
_ = m.Down(gophant.Migrations(), 1)
```

CLI:

```
gophant migrate
gophant migrate:rollback 1
gophant migrate:fresh
gophant migrate:status
gophant migrate:status --json
```

Notes:
- Use `gophant migrate` during development after adding a migration file.
- For demo apps, prefer running migrations explicitly before `serve`.

Notes:
- `schema.New("mysql")` or `schema.New("postgres")` controls DDL type mapping.
- Use `Build` if you need indexes:

```

Extra columns:
- `Text`, `JSON`, `Decimal`, `Enum`
- `BigInteger`, `UnsignedBigInteger`
- Composite indexes: `CompositeIndex`, `CompositeUnique`
bp, sql := b.Build("users", func(t *schema.Blueprint) {
	t.Increments("id")
	t.String("email", 255)
	t.Unique("email")
	t.Index("email")
})
_, _ = db.Exec(sql)
for _, idx := range b.Indexes(bp) {
	_, _ = db.Exec(idx)
}
```
