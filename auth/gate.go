package auth

import (
	gomvchttp "github.com/mahavirnahata/gophant/http"
)

type Ability func(*gomvchttp.Context) bool

type Gate struct {
	abilities map[string]Ability
}

func NewGate() *Gate {
	return &Gate{abilities: map[string]Ability{}}
}

func (g *Gate) Define(name string, fn Ability) {
	g.abilities[name] = fn
}

func (g *Gate) Allows(name string, c *gomvchttp.Context) bool {
	fn, ok := g.abilities[name]
	if !ok {
		return false
	}
	return fn(c)
}

func (g *Gate) Require(name string) gomvchttp.Middleware {
	return func(next gomvchttp.Handler) gomvchttp.Handler {
		return func(c *gomvchttp.Context) {
			if g.Allows(name, c) {
				next(c)
				return
			}
			c.JSON(403, map[string]any{"error": "forbidden"})
		}
	}
}
