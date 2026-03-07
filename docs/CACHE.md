# Cache

Supported drivers:
- memory
- redis

Config (`.env`):

```
CACHE_DRIVER=memory
CACHE_PREFIX=gophant:cache:
```

Usage:

```
app := gophant.New()

_ = app.Cache.Set("key", "value", time.Minute)
val, ok := app.Cache.Get("key")

res, _ := app.Cache.Remember("count", time.Minute, func() (any, error) {
	return 123, nil
})
```

Tagged cache:

```
app := gophant.New()

key := app.Cache.Tag("users", "list")
_ = app.Cache.Set(key, []int{1, 2, 3}, time.Minute)

_ = app.Cache.FlushTag("users")
```

CLI:

```
gophant cache:clear
```
