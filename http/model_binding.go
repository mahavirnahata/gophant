package http

import (
	"errors"
	"reflect"

	"github.com/mahavirnahata/gophant/db"
)

type ModelBinder func(id string, dest any) error

var binderRegistry = map[reflect.Type]ModelBinder{}

func RegisterModelBinder(sample any, binder ModelBinder) {
	t := reflect.TypeOf(sample)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	binderRegistry[t] = binder
}

func bindModel(c *Context, modelType reflect.Type) error {
	id := c.Param("id")
	if id == "" {
		return errors.New("missing id")
	}
	binder, ok := binderRegistry[modelType]
	if !ok {
		return errors.New("no binder for model")
	}
	inst := reflect.New(modelType).Interface()
	if err := binder(id, inst); err != nil {
		return err
	}
	c.SetModel(modelType.Name(), inst)
	return nil
}

// Default binder using db.Model + DefaultDB.
func DefaultModelBinder(table string, idColumn string) ModelBinder {
	return func(id string, dest any) error {
		if db.DefaultDB == nil {
			return errors.New("default db not set")
		}
		m := db.NewModel(db.DefaultDB, table)
		return m.Where(idColumn, "=", id).FirstStruct(dest)
	}
}
