package cache

import "time"

type Store interface {
	Get(key string) (any, bool)
	Set(key string, value any, ttl time.Duration) error
	Delete(key string) error
	Flush() error
	FlushTag(tag string) error
}

type Cache struct {
	Store Store
}

func New(store Store) *Cache {
	return &Cache{Store: store}
}

func (c *Cache) Get(key string) (any, bool) {
	return c.Store.Get(key)
}

func (c *Cache) Set(key string, value any, ttl time.Duration) error {
	return c.Store.Set(key, value, ttl)
}

func (c *Cache) Delete(key string) error {
	return c.Store.Delete(key)
}

func (c *Cache) Flush() error {
	return c.Store.Flush()
}

func (c *Cache) FlushTag(tag string) error {
	return c.Store.FlushTag(tag)
}

func (c *Cache) Remember(key string, ttl time.Duration, fn func() (any, error)) (any, error) {
	if val, ok := c.Get(key); ok {
		return val, nil
	}
	val, err := fn()
	if err != nil {
		return nil, err
	}
	_ = c.Set(key, val, ttl)
	return val, nil
}

func (c *Cache) RememberTagged(tag, key string, ttl time.Duration, fn func() (any, error)) (any, error) {
	tkey := tagKey(tag, key)
	return c.Remember(tkey, ttl, fn)
}

func (c *Cache) Tag(tag, key string) string {
	return tagKey(tag, key)
}
