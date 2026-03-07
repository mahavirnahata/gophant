package container

import (
	"errors"
	"reflect"
)

type Container struct {
	factories map[reflect.Type]func() any
	values    map[reflect.Type]any
}

func New() *Container {
	return &Container{factories: map[reflect.Type]func() any{}, values: map[reflect.Type]any{}}
}

func (c *Container) Bind(sample any, factory func() any) {
	t := reflect.TypeOf(sample)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	c.factories[t] = factory
}

func (c *Container) Set(sample any) {
	t := reflect.TypeOf(sample)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	c.values[t] = sample
}

func (c *Container) Resolve(t reflect.Type) (any, error) {
	if t.Kind() == reflect.Pointer {
		val, ok := c.values[t.Elem()]
		if ok {
			return val, nil
		}
		if factory, ok := c.factories[t.Elem()]; ok {
			return factory(), nil
		}
	}
	val, ok := c.values[t]
	if ok {
		return val, nil
	}
	if factory, ok := c.factories[t]; ok {
		return factory(), nil
	}
	return nil, errors.New("no binding for type")
}
