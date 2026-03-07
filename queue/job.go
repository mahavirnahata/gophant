package queue

import (
	"encoding/json"
	"errors"
	"reflect"
)

type Envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Registry struct {
	constructors map[string]func() JobHandler
}

type JobHandler interface {
	Handle() error
}

func NewRegistry() *Registry {
	return &Registry{constructors: map[string]func() JobHandler{}}
}

func (r *Registry) Register(name string, ctor func() JobHandler) {
	r.constructors[name] = ctor
}

func (r *Registry) RegisterType(sample JobHandler, ctor func() JobHandler) {
	r.Register(typeName(sample), ctor)
}

func (r *Registry) TypeName(job JobHandler) string {
	return typeName(job)
}

func (r *Registry) Serialize(job JobHandler) ([]byte, error) {
	name := typeName(job)
	payload, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{Type: name, Payload: payload})
}

func (r *Registry) Deserialize(data []byte) (JobHandler, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	ctor, ok := r.constructors[env.Type]
	if !ok {
		return nil, errors.New("unknown job type: " + env.Type)
	}
	job := ctor()
	if err := json.Unmarshal(env.Payload, job); err != nil {
		return nil, err
	}
	return job, nil
}

func typeName(v any) string {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	return rv.Type().PkgPath() + "." + rv.Type().Name()
}
